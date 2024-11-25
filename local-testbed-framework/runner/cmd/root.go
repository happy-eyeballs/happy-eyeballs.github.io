package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.lrz.de/netintum/projects/gino/students/happy-eyeballs-measurement-framework/runner/cmd/internal"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "runner",
	Short: "Automation Framework for test cases",
	Long:  "Automatically execute a user-defined list of test cases.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Persistent flags
	rootCmd.PersistentFlags().Int8VarP(
		&internal.Verbosity,
		"verbosity", "v", 2,
		"Verbosity (log level) of the program's log output",
	)
	rootCmd.PersistentFlags().StringVarP(
		&internal.LogFilePath,
		"log-file", "l", "",
		"Path to the file where this program's logging output will be written to (default stdout)",
	)
	rootCmd.PersistentFlags().Lookup("verbosity").NoOptDefVal = "1"
}
