package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "FIleMove-cli",
	Short: "A CLI tool for file Upload and Download.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Please use one of the available commands (signup, login).")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
