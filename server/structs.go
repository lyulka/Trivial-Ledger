package server

import "github.com/lyulka/trivial-ledger/structs"

type ProposeTransactionRequest structs.ProposedTransaction

type ProposeTransactionResponse struct {
	blockHash string `json:"blockHash"`
	txNumber  int    `json:"txNumber"`
}

type GetTransactionRequest struct {
	blockHash string `json:"blockHash"`
	txNumber  int    `json:"txNumber"`
}

type GetTransactionResponse structs.Transaction

type GetBlockRequest struct {
	blockHash string `json:"blockHash"`
}

type GetBlockResponse structs.Block
