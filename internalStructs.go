package main

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

// Why this is necessary is complicated. It is necessary for the correctness
// of TLedger's protocol and is inspired by cumulative acknowledgement in network protocols.
type transactionEntry struct {
	structs.Transaction
	isCommitted bool
}
