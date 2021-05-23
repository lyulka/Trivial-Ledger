package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	st "github.com/lyulka/trivial-ledger/structs"
	cv3 "go.etcd.io/etcd/clientv3"
)

const PREFIX = "tledger/"

var TLEDGER_SERVER_ENDPOINT = "localhost:9090"

var DEFAULT_ENDPOINTS []string = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}
var DEFAULT_DIAL_TIMEOUT time.Duration = 1 * time.Second

type Server struct {
	Router     *httprouter.Router
	etcdClient *cv3.Client

	blockCache BlockCache

	// TODO: Refactor into type BlockchainEngine
	latestTxIndex int64
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
		Router:        router,
		etcdClient:    client,
		blockCache:    make(BlockCache),
		latestTxIndex: -1,
	}

	server.blockCache = make(BlockCache)

	err = server.BringBlockCacheAndLatestTxNumUpToDate()
	if err != nil {
		return nil, err
	}

	router.GET("/helloWorld", server.HelloWorldGet)
	router.GET("/getTransaction", server.getTransactionGet)
	router.GET("/getBlock", server.getBlockGet)
	router.POST("/proposeTransaction", server.proposeTransactionPost)

	return &server, nil
}

func (s *Server) BringBlockCacheAndLatestTxNumUpToDate() error {

	// Determine the latest transaction index (note that this is distinct from within-block txNum)
	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)

	resp, err := s.etcdClient.Get(
		ctx, PREFIX,
		cv3.WithPrefix(),
		cv3.WithSort(cv3.SortByKey, cv3.SortDescend),
		cv3.WithLimit(1))
	cancel()
	if err != nil {
		return err
	}

	if len(resp.Kvs) != 0 {
		// Store latestTxIndex
		keyPathSegments := strings.Split(string(resp.Kvs[0].Key), "/")
		s.latestTxIndex, err = binaryToInt(keyPathSegments[len(keyPathSegments)-1])
		if err != nil {
			return err
		}
	}

	// Determine the block number up to which transactions have been committed (floor div.)
	latestBlockNumCommitted := int(s.latestTxIndex+1) / st.DEFAULT_BLOCK_SIZE

	err = s.blockCache.PopulateWithBlocks(*s.etcdClient,
		s.blockCache.latestBlockNumInCache()+1, latestBlockNumCommitted)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Teardown() {
	s.etcdClient.Close()

	outDir := fmt.Sprintf("/logs" + TLEDGER_SERVER_ENDPOINT)
	timeStamp := time.Now().String()
	fmt.Println("RBDNS: SIGINT received. Printing BlockCache into " + outDir + "/" + timeStamp)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		os.Mkdir(outDir, 0755)
	}

	outFile, err := os.Create(outDir + "/" + timeStamp)
	if err != nil {
		fmt.Println("RBDNS: Failed to log results.")
		fmt.Println(err)
	}
	outFile.Write([]byte(fmt.Sprintf("%v", s.blockCache)))
	outFile.Close()

	fmt.Println("RBDNS: Tearing down server. Goodbye!")
}

func (s *Server) proposeTransaction(propTx st.ProposedTransaction) (blockNum int, txNumber int, err error) {

	accepted := false
	for !accepted {
		s.BringBlockCacheAndLatestTxNumUpToDate()

		// If the etcd TXN coming up succeeds, the new transaction will have the following
		// blockNum and txNumber
		blockNum = s.blockCache.latestBlockNumInCache() + 1
		txNumber = int(s.latestTxIndex+1) % st.DEFAULT_BLOCK_SIZE

		// Generate KV pair
		key := fmt.Sprintf("%s%s", PREFIX, intToBinary(s.latestTxIndex+1))
		valueBytes, err := json.Marshal(st.Transaction{
			ProposedTransaction: propTx,
			BlockNum:            blockNum,
			TxNumber:            txNumber,
		})
		value := string(valueBytes)

		if err != nil {
			return -1, -1, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)
		// CreateRevision == 0 if key does not exist
		txnResp, err := s.etcdClient.Txn(ctx).If(
			cv3.Compare(cv3.CreateRevision(key), "=", 0),
		).Then(
			cv3.OpPut(key, value),
		).Commit()

		cancel()
		if err != nil {
			return -1, -1, err
		}

		if txnResp.Succeeded {
			accepted = true
		}
	}

	return blockNum, txNumber, nil

}

// Queries block cache
// txNum must be smaller than configured block size
func (s *Server) getTransaction(blockNum int, txNum int) *st.Transaction {

	block := s.getBlock(blockNum)
	if block == nil {
		return nil
	}

	return &block.Transactions[txNum]
}

// Queries block cache
func (s *Server) getBlock(blockNum int) *st.Block {

	// Check if block is in BlockCache
	if block, ok := s.blockCache[blockNum]; ok {
		return &block
	}

	s.BringBlockCacheAndLatestTxNumUpToDate()

	// Check again
	if block, ok := s.blockCache[blockNum]; ok {
		return &block
	}

	return nil
}

func (s *Server) proposeTransactionPost(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	proposedTx := st.ProposedTransaction{}
	err := json.NewDecoder(r.Body).Decode(&proposedTx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Request body should be properly formatted JSON")
		return
	}

	blockNum, txNumber, err := s.proposeTransaction(proposedTx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "TODO: more granular errors")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(ProposeTransactionResponse{
		BlockNum: blockNum,
		TxNumber: txNumber,
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

	transaction := s.getTransaction(reqBody.BlockNum, reqBody.TxNumber)
	if transaction == nil {
		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetTransactionResponse(*transaction))
}

func (s *Server) getBlockGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	reqBody := GetBlockRequest{}
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Request body should be properly formatted JSON")
		return
	}

	block := s.getBlock(reqBody.BlockNum)
	if block == nil {
		w.WriteHeader(http.StatusNoContent)
		io.WriteString(w, "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetBlockResponse(*block))
}

func (s *Server) HelloWorldGet(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {

	fmt.Println("Hello world: in")
	fmt.Fprintln(w, "Hello world!")
}

func intToBinary(n int64) string {
	return fmt.Sprintf("%064s", strconv.FormatInt(int64(n), 2))
}

func binaryToInt(binary string) (int64, error) {
	i, err := strconv.ParseInt(binary, 2, 64)
	if err != nil {
		return -1, err
	}

	return i, nil
}
