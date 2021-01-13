package cmd

import (
	"fmt"
	"os"

	"github.com/soldatov-s/go-garage/app"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// RootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:     appName,
		Short:   description,
		Version: appFullVersion,
	}

	rootCmd.AddCommand(app.CreateServeCmd(serveHandler))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
