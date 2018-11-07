package goblockchain

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address who want to know their balance.")
	addBlockAddress := addBlockCmd.String("address", "", "The address to send block reward to.")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to.")

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "createblockchain", err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "printchain", err)
		}
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "addblock", err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, "getbalance", err)
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

	if addBlockCmd.Parsed() {
		if *addBlockAddress == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock([]*Transaction{}, *addBlockAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
}
func (cli *CLI) addBlock(txs []*Transaction, address string) {
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

func (cli *CLI) getBalance(address string) {
	bc := NewBlockchain(address)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
