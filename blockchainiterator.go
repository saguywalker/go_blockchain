package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

type BlockchainIterator struct {
	currentBlockHash []byte
	db               *bolt.DB
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

func (it *BlockchainIterator) Next() *Block {
	var block *Block
	err := it.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockHash := b.Get(it.currentBlockHash)
		block = DeserializeBlock(blockHash)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with bolt db file", err)
	}

	it.currentBlockHash = block.PrevBlockHash

	return block
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}
