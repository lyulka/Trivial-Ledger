package structs

import (
	"crypto/sha256"
	"fmt"
	"time"
)

const DEFAULT_BLOCK_SIZE int = 25

type ProposedTransaction struct {
	Content string `json:"content"`
}

type Transaction struct {
	ProposedTransaction

	// BlockNum+TxNumber uniquely identifies a transaction.
	BlockNum int `json:"blockNum"`
	TxNumber int `json:"txNumber"`
}

type Block struct {

	// Uniquely identifies a block in the blockchain.
	BlockNum int

	// Equal to the timestamp of the *last* transaction in TxList
	Timestamp string `json:"timestamp"`

	// SHA-256 hash.
	PreviousHash string `json:"previousHash"`

	// SHA-256 hash of the entire block.
	Hash string `json:"hash"`

	TxList [DEFAULT_BLOCK_SIZE]Transaction
}

// transactions need to be ordered in the same way they will be packed into the
// block
func NewBlock(blockNum int, prevHash string, transactions [DEFAULT_BLOCK_SIZE]Transaction) (*Block, error) {

	b := Block{
		BlockNum:     blockNum,
		Timestamp:    time.Now().String(),
		PreviousHash: prevHash,
		Hash:         "",
		TxList:       transactions,
	}

	hash := AsSha256(b)
	b.Hash = hash

	return &b, nil
}

func AsSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}
