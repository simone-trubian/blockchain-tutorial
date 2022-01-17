package main

import (
	"fmt"
	"os"

	"github.com/simone-trubian/blockchain-tutorial/node"
	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Launches the SB node and its HTTP API.",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetUint64(flagPort)

			fmt.Println("Launching SB node and its HTTP API...")

			bootstrap := node.NewPeerNode(
				"18.184.213.146",
				8080,
				true,
				true,
			)

			n := node.New(getDataDirFromCmd(cmd), port, bootstrap)
			err := n.Run()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	addDefaultRequiredFlags(runCmd)
	runCmd.Flags().Uint64(
		flagPort,
		node.DefaultHTTPort,
		"exposed HTTP port for communication with peers")

	return runCmd
}
