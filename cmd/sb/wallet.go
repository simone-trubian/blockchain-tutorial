package main

import (
	"bytes"
	"fmt"
	"os"
	"syscall"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/simone-trubian/blockchain-tutorial/wallet"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func walletCmd() *cobra.Command {
	var walletCmd = &cobra.Command{
		Use:   "wallet",
		Short: "Manages accounts, keys, cryptography.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return incorrectUsageErr()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	walletCmd.AddCommand(walletNewAccountCmd())

	return walletCmd
}

func walletNewAccountCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new-account",
		Short: "Creates a new account with a new set of a elliptic-curve Private + Public keys.",
		Run: func(cmd *cobra.Command, args []string) {
			password := getPassPhrase("Please enter a password to encrypt the new wallet:", true)

			dataDir := getDataDirFromCmd(cmd)

			ks := keystore.NewKeyStore(wallet.GetKeystoreDirPath(dataDir), keystore.StandardScryptN, keystore.StandardScryptP)
			acc, err := ks.NewAccount(password)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			fmt.Printf("New account created: %s\n", acc.Address.Hex())
		},
	}

	addDefaultRequiredFlags(cmd)

	return cmd
}

func getPassPhrase(prompt string, confirmation bool) string {
	fmt.Println(prompt)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		utils.Fatalf("Failed to read password: %v", err)
	}

	if confirmation {
		confirm, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			utils.Fatalf("Failed to read password confirmation: %v", err)
		}
		if bytes.Compare(bytePassword, confirm) != 0 {
			utils.Fatalf("Passwords do not match")
		}
	}

	return string(bytePassword[:])
}
