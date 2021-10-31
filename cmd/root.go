package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "entangle",
	Short: "A library for building peer-to-peer file sharing solutions.",
	Long: `A library for building peer-to-peer file sharing solutions.

	For more information, please visit https://github.com/alphahorizon/libentangle`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(signalCmd)
	rootCmd.AddCommand(clientCmd)
}
