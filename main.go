package main

import (
	"math"
	"strconv"
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

func IntToHex(num int64) []byte {
	return []byte(strconv.FormatInt(num, 16))
}
