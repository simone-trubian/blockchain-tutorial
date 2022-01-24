package main

import (
	"context"
	"fmt"
	"os"

	"github.com/simone-trubian/blockchain-tutorial/database"
	"github.com/simone-trubian/blockchain-tutorial/node"
	"github.com/spf13/cobra"
)

var migrateCmd = func() *cobra.Command {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the blockchain database according to new business rules.",
		Run: func(cmd *cobra.Command, args []string) {
			state, err := database.NewStateFromDisk(getDataDirFromCmd(cmd))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			defer state.Close()

			pendingBlock := node.NewPendingBlock(
				database.Hash{},
				state.NextBlockNumber(),
				[]database.Tx{
					database.NewTx("simone", "simone", 3, ""),
					database.NewTx("simone", "simone", 700, "reward"),
					database.NewTx("simone", "tanya", 2000, ""),
					database.NewTx("simone", "simone", 100, "reward"),
					database.NewTx("tanya", "simone", 1, ""),
					database.NewTx("tanya", "ugo", 1000, ""),
					database.NewTx("tanya", "simone", 50, ""),
					database.NewTx("simone", "simone", 600, "reward"),
					database.NewTx("simone", "andrej", 24700, "reward"),
				},
			)

			_, err = node.Mine(context.Background(), pendingBlock)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(migrateCmd)

	return migrateCmd
}
