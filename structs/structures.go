package structs

const BLOCK_SIZE int = 25

type ProposedTransaction struct {
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

type Transaction struct {
	ProposedTransaction

	// BlockHash+TxNumber uniquely identifies a transaction.
	BlockHash string `json:"blockHash"`
	TxNumber  string `json:"txNumber"`
}

type Block struct {

	// Equal to the timestamp of the *last* transaction in TxList
	Timestamp string `json:"timestamp"`

	// SHA-256 hash.
	PreviousHash string `json:"previousHash"`

	// SHA-256 hash of the entire block (Timestamp included).
	// Uniquely identifies a block in the blockchain.
	// This is used to cache the block in-memory for quick response
	// to getTransaction() and getBlock()
	Hash string `json:"hash"`

	TxList [BLOCK_SIZE]Transaction
}
