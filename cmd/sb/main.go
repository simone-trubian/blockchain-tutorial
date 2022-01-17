package main

import (
	"fmt"
	"os"

	"github.com/simone-trubian/blockchain-tutorial/fs"
	"github.com/spf13/cobra"
)

const flagDataDir = "datadir"
const flagPort = "port"

func main() {
	var sbCmd = &cobra.Command{
		Use:   "sb",
		Short: "Simone's blockchain command line interface",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	sbCmd.AddCommand(versionCmd)
	sbCmd.AddCommand(runCmd())
	sbCmd.AddCommand(balancesCmd())
	sbCmd.AddCommand(migrateCmd())

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

func getDataDirFromCmd(cmd *cobra.Command) string {
	dataDir, _ := cmd.Flags().GetString(flagDataDir)

	return fs.ExpandPath(dataDir)
}

func incorrectUsageErr() error {
	return fmt.Errorf("incorrect usage")
}
