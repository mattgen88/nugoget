package cmd

import (
	"github.com/spf13/cobra"
)

var (
	debug   = false
	rootCmd = &cobra.Command{
		Use:   "nugoget",
		Short: "Update dependencies of a dotnet project",
		Long:  `This program is designed to help with updating dependencies`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "set to true to output debug data")
	rootCmd.AddCommand(updateCmd)
}
