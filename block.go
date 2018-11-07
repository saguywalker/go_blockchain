package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

func (b *Block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.HashTransaction(), timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

func (b *Block) Serialization() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Error to serialize block:%x.\n", b.Hash)
		os.Exit(1)
	}
	return result.Bytes()
}

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func NewBlock(txs []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), txs, prevBlockHash, []byte{}, 0, height}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Solve()
	block.Nonce = nonce
	block.Hash = hash[:]
	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
