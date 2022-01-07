package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	communityKey = "community"
)

var rootCmd = &cobra.Command{
	Use:   "entangle",
	Short: "A library for building peer-to-peer file sharing solutions.",
	Long: `A library for building peer-to-peer file sharing solutions.

	For more information, please visit https://github.com/alphahorizon/libentangle`,
}

func Execute() {

	rootCmd.PersistentFlags().String(communityKey, "testCommunityName", "Community to join")

	// Bind env variables
	if err := viper.BindPFlags(serverCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}

	viper.AutomaticEnv()

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.AddCommand(signalCmd)
	rootCmd.AddCommand(clientCmd)
	rootCmd.AddCommand(serverCmd)
}
