/*
Copyright © 2024 BubbaCodeX
*/
package cmd

import (
	"os"

	"DomainSlicer_V1.1/cmd/network"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "DomainSlicer_V1.1",
	Short: "Pings hosts from a list",
	Long: `Pings hosts from a list and sorts them based on the http responses,
	enabling a smoother bug bounty workflow.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
func addSubCommands() {
	rootCmd.AddCommand(network.ParseCmd)
}
func init() {

	addSubCommands()
}
