package main

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
