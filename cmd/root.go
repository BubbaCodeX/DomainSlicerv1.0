/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.DomainSlicer_V1.1.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	addSubCommands()
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
