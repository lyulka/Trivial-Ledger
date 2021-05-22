package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lyulka/trivial-ledger/structs"
	"go.etcd.io/etcd/clientv3"
)

var DEFAULT_ENDPOINTS []string = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}

var DEFAULT_DIAL_TIMEOUT time.Duration = 1 * time.Second

type Server struct {
	Router     *httprouter.Router
	etcdClient *clientv3.Client

	// TODO: Refactor into type BlockchainEngine
	latestTxIndex int
}

func New() (*Server, error) {

	client, err := clientv3.New(clientv3.Config{
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

	server.Initialize()

	router.GET("/helloWorld", server.HelloWorldGet)
	router.GET("/getTransaction", server.getTransactionGet)
	router.GET("/getBlock", server.getBlockGet)
	router.POST("/proposeTransaction", server.proposeTransactionPost)

	return &server, nil
}

// Initialize's main purpose is to populate BlockCache
func (s *Server) Initialize() {
	// Determine the latest transaction index (this is distinct from within-block txNum)

}

func (s *Server) Teardown() {
	s.etcdClient.Close()

	fmt.Println("RBDNS: Tearing down server. Goodbye!")
}

func (s *Server) HelloWorldGet(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	fmt.Println("Hello world: in")

	fmt.Fprintln(w, "Hello world!")
}

func (s *Server) proposeTransaction(propTx structs.ProposedTransaction) (blockHash string, txNumber int, err error) {

}

func (s *Server) getTransaction(blockHash string, txNum int) (structs.Transaction, error) {

}

func (s *Server) getBlock(blockHash string) (structs.Block, error) {

}

func (s *Server) proposeTransactionPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	proposedTx := structs.ProposedTransaction{}
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
