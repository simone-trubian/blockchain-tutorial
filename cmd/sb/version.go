package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Major = "0"
const Minor = "7"
const Fix = "0"
const Verbal = "Sync"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Describes version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s.%s.%s-beta %s", Major, Minor, Fix, Verbal)
	},
}
