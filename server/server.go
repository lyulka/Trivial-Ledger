package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	st "github.com/lyulka/trivial-ledger/structs"
	cv3 "go.etcd.io/etcd/clientv3"
)

const TLEDGER_PREFIX = "tledger/"

const DEFAULT_TLEDGER_SERVER_ENDPOINT = "http://localhost:8081"

var DEFAULT_ENDPOINTS []string = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}
var DEFAULT_DIAL_TIMEOUT time.Duration = 1 * time.Second

type Server struct {
	Router     *httprouter.Router
	etcdClient *cv3.Client

	blockCache BlockCache

	// TODO: Refactor into type BlockchainEngine
	latestTxIndex int // >= 1
}

func New() (*Server, error) {

	client, err := cv3.New(cv3.Config{
		Endpoints:   DEFAULT_ENDPOINTS,
		DialTimeout: DEFAULT_DIAL_TIMEOUT,
	})
	if err != nil {
		return nil, err
	}

	router := httprouter.New()
	server := Server{
		Router:     router,
		etcdClient: client,
	}

	server.blockCache = make(BlockCache)

	err = server.Initialize()
	if err != nil {
		return nil, err
	}

	router.GET("/helloWorld", server.HelloWorldGet)
	router.GET("/getTransaction", server.getTransactionGet)
	router.GET("/getBlock", server.getBlockGet)
	router.POST("/proposeTransaction", server.proposeTransactionPost)

	return &server, nil
}

// Initialize's main purpose is to quickly populate BlockCache on startup
func (s *Server) Initialize() error {

	fmt.Println("Initialize: in")

	// Determine the latest transaction index (this is distinct from within-block txNum)
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)

	resp, err := s.etcdClient.Get(
		ctx, TLEDGER_PREFIX,
		cv3.WithPrefix(),
		cv3.WithSort(cv3.SortByKey, cv3.SortDescend),
		cv3.WithLimit(1))
	cancel()
	if err != nil {
		return err
	}

	// No transactions have been committed yet
	if len(resp.Kvs) == 0 {
		s.latestTxIndex = 0
	} else {
		s.latestTxIndex, err = strconv.Atoi(string(resp.Kvs[0].Key))
		if err != nil {
			return err
		}
	}

	// Determine the block number up to which transactions have been committed (floor div.)
	var latestBlockCommitted int = (s.latestTxIndex + 1) / st.DEFAULT_BLOCK_SIZE

	err = s.blockCache.PopulateWithBlocks(*s.etcdClient, 0, int(latestBlockCommitted))
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Teardown() {
	s.etcdClient.Close()

	fmt.Println("RBDNS: Tearing down server. Goodbye!")
}

func (s *Server) proposeTransaction(propTx st.ProposedTransaction) (blockNum string, txNumber int, err error) {
}

// Queries block cache
func (s *Server) getTransaction(blockNum string, txNum int) (st.Transaction, error) {

	// First check if blockNum is in BlockCache
}

// Queries block cache
func (s *Server) getBlock(blockNum string) (st.Block, error) {

}

func (s *Server) proposeTransactionPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	fmt.Println("proposeTransactionPost: in")

	proposedTx := st.ProposedTransaction{}
	err := json.NewDecoder(r.Body).Decode(&proposedTx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Request body should be properly formatted JSON")
		return
	}

	blockHash, txNumber, err := s.proposeTransaction(proposedTx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "TODO: more granular errors")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ProposeTransactionResponse{
		blockHash: blockHash,
		txNumber:  txNumber,
	})
}

func (s *Server) getTransactionGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	fmt.Println("getTransactionGet: in")

	reqBody := GetTransactionRequest{}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Request body should be properly formatted JSON")
		return
	}

	transaction, err := s.getTransaction(reqBody.blockHash, reqBody.txNumber)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "TODO: more granular errors")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetTransactionResponse(transaction))
}

func (s *Server) getBlockGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	fmt.Println("getBlockGet: in")

	reqBody := GetBlockRequest{}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Request body should be properly formatted JSON")
		return
	}

	block, err := s.getBlock(reqBody.blockHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "TODO: more granular errors")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetBlockResponse(block))
}

func (s *Server) HelloWorldGet(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	fmt.Println("Hello world: in")
	fmt.Fprintln(w, "Hello world!")
}
