package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"

func main() {
	var sbCmd = &cobra.Command{
		Use:   "sb",
		Short: "Simone's blockchain command line interface",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	sbCmd.AddCommand(versionCmd)
	sbCmd.AddCommand(balancesCmd())
	sbCmd.AddCommand(txCmd())

	err := sbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDefaultRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().String(
		flagDataDir,
		"",
		"Absolute path to the node data dir where the DB will/is stored")
	cmd.MarkFlagRequired(flagDataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
