package main

import (
	"fmt"
	"os"
	"time"

	"github.com/simone-trubian/blockchain-tutorial/database"
)

func main() {
	cwd, _ := os.Getwd()
	state, err := database.NewStateFromDisk(cwd)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
	defer state.Close()

	block0 := database.NewBlock(
		database.Hash{},
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("simone", "simone", 3, ""),
			database.NewTx("simone", "simone", 700, "reward"),
		},
	)

	state.AddBlock(block0)
	block0hash, _ := state.Persist()

	block1 := database.NewBlock(
		block0hash,
		uint64(time.Now().Unix()),
		[]database.Tx{
			database.NewTx("simone", "tanya", 2000, ""),
			database.NewTx("simone", "simone", 100, "reward"),
			database.NewTx("tanya", "simone", 1, ""),
			database.NewTx("tanya", "ugo", 1000, ""),
			database.NewTx("tanya", "simone", 50, ""),
			database.NewTx("simone", "simone", 600, "reward"),
		},
	)

	state.AddBlock(block1)
	state.Persist()
}
