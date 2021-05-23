package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	st "github.com/lyulka/trivial-ledger/structs"
	cv3 "go.etcd.io/etcd/clientv3"
)

// BlockCache is a map between blockNum and Block.
// It contains only committed blocks.
type BlockCache map[int]st.Block

// start block, end block. end is exclusive
// Only blocks that have been 'committed' are supposed to be added into BlockCache
func (bc BlockCache) PopulateWithBlocks(client cv3.Client, start int, end int) error {

	if start == end {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DEFAULT_DIAL_TIMEOUT)

	resp, err := client.Get(
		ctx,
		fmt.Sprintf("%s%s", PREFIX, intToBinary(int64(start*st.DEFAULT_BLOCK_SIZE))),              // Start key
		cv3.WithRange(fmt.Sprintf("%s%s", PREFIX, intToBinary(int64(end*st.DEFAULT_BLOCK_SIZE)))), // End key
		cv3.WithSort(cv3.SortByKey, cv3.SortAscend))

	cancel()
	if err != nil {
		return err
	}

	if len(resp.Kvs)%st.DEFAULT_BLOCK_SIZE != 0 {
		return errors.New("PopulateWithBlocks: MAJOR ARITHMETIC PROBLEM IN PROGRAMMING")
	}

	// Generate blocks and add into BlockCache
	blocksAdded := 0
	for blockNum := start; blockNum < end; blockNum += 1 {

		var prevHash string

		// Special case for first block, no prevHash
		if blockNum == 0 {
			prevHash = ""
		} else {
			prevBlock, ok := bc[blockNum-1]
			if !ok {
				return errors.New("PopulateWithBlocks: fail to get prevBlock")
			}
			prevHash = prevBlock.Hash
		}

		// Extract values into an array of transactions
		// transactions := extractTransactions(resp.Kvs, blocksAdded*st.DEFAULT_BLOCK_SIZE)
		var transactions [st.DEFAULT_BLOCK_SIZE]st.Transaction
		for i := blocksAdded * st.DEFAULT_BLOCK_SIZE; i < st.DEFAULT_BLOCK_SIZE; i += 1 {
			txBytes := resp.Kvs[i].Value

			transaction := st.Transaction{}
			err := json.Unmarshal(txBytes, &transaction)

			if err != nil {
				return err
			}
			transactions[i] = transaction
		}

		block, err := st.NewBlock(blockNum, prevHash, transactions)
		if err != nil {
			return err
		}

		if _, ok := bc[blockNum]; ok {
			fmt.Printf("PopulateWithBlocks: blockNum: %d is already in cache\n", blockNum)
		}

		bc[blockNum] = *block
	}

	return nil
}

func (bc BlockCache) latestBlockNumInCache() int {
	return len(bc) - 1
}

// TODO: figure out why types are not allowed for arguments
// func extractTransactions(kvs []*mvccpb.KeyValue, start int) (*[st.DEFAULT_BLOCK_SIZE]st.Transaction, error) {
// 	var transactions [st.DEFAULT_BLOCK_SIZE]st.Transaction

// 	for i := start; i < st.DEFAULT_BLOCK_SIZE; i += 1 {
// 		txBytes := kvs[i].Value

// 		transaction := st.Transaction{}
// 		err := json.Unmarshal(txBytes, &transaction)

// 		if err != nil {
// 			return nil, err
// 		}
// 		transactions[i] = transaction
// 	}

// 	return &transactions, nil
// }
