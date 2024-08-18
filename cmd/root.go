package cmd

import (
	"fmt"
	"os"

	"github.com/savioxavier/termlink"
	"github.com/spf13/cobra"

	"github.com/artmoskvin/hide/pkg/config"
)

func init() {
	cobra.EnableTraverseRunHooks = true

	rootCmd.AddCommand(runCmd)
}

var rootCmd = &cobra.Command{
	Use: "hide",
	Long: fmt.Sprintf(`%s
%s is a headless IDE for coding agents.
	`, splash, termlink.ColorLink("Hide", "https://hide.sh", "blue")),
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	Version: config.Version(),
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
