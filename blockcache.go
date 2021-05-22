package main

import "github.com/lyulka/trivial-ledger/structs"

type blockHash string
type txNum string

type BlockCache map[blockHash]structs.Block
