package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
)

func main() {
	cli := CLI{}
	cli.Run()
}

const targetBits = 22
const maxNonce = math.MaxInt64
const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}
