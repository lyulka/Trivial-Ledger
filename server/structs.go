package server

import "github.com/lyulka/trivial-ledger/structs"

type ProposeTransactionRequest structs.ProposedTransaction

type ProposeTransactionResponse struct {
	BlockNum int `json:"blockNum"`
	TxNumber int `json:"txNumber"`
}

type GetTransactionRequest struct {
	BlockNum int `json:"blockNum"`
	TxNumber int `json:"txNumber"`
}

type GetTransactionResponse structs.Transaction

type GetBlockRequest struct {
	BlockNum int `json:"blockNum"`
}

type GetBlockResponse structs.Block
