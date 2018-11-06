package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	cli := CLI{}
	cli.Run()
}

const targetBits = 16
const maxNonce = math.MaxInt64
const dbfile = "blockchain_db"
const blocksBucket = "first_chain"
const genesisCoinbaseData = "Thx, miner(s)!"

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

type BlockchainIterator struct {
	currentBlockHash []byte
	db               *bolt.DB
}

type CLI struct {
	bc *Blockchain
}

func IntToHex(num int64) []byte {
	return []byte(strconv.FormatInt(num, 16))
}

func (b *Block) setHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.HashTransaction(), timestamp}, []byte{})
	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

func NewBlock(txs []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), txs, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Solve()
	block.Nonce = nonce
	block.Hash = hash[:]
	return pow.block
}

func (bc *Blockchain) AddBlock(txs []*Transaction) {
	var tip []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	newBlock := NewBlock(txs, tip)
	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialization())
		err = b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error addblock", err)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error with bolt db", err)
	}
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func NewBlockchain(address string) *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbfile, 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if b == nil {
			cbtx := NewCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(cbtx)
			b, err = tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialization())
			err = b.Put([]byte("l"), genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})
	bc := Blockchain{tip, db}
	return &bc
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(3)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransaction(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
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

func (pow *ProofOfWork) Solve() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("Mining the block...")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println("\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	var hashInt big.Int
	hashInt.SetBytes(hash[:])
	isValid := hashInt.Cmp(pow.target) == -1
	return isValid
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

func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Printf("Error to bytes: %s", d)
	}
	return &block
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	fmt.Println(string(bc.tip))
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

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to.")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "addblock", err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "printchain", err)
		}
	default:
		//cli.printUsage()
		fmt.Printf("Unknown argument: %s\n", string(os.Args[1]))
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) addBlock(txs []*Transaction) {
	cli.bc.AddBlock(txs)
	fmt.Println("Success!")
}

func (cli *CLI) printChain() {
	fmt.Println(&cli)
	fmt.Println(&cli.bc.tip)
	bci := cli.bc.Iterator()
	for {
		block := bci.Next()
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.HashTransaction())
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()
	fmt.Println("Done!")
}
