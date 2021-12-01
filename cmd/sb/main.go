package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var sbCmd = &cobra.Command{
		Use:   "sb",
		Short: "Simone's blockchain command line interface",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	sbCmd.AddCommand(versionCmd)

	err := sbCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
