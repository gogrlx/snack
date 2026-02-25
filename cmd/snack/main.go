// snack is a unified CLI for system package managers.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snack",
	Short: "A unified CLI for system package managers",
	Long: `snack wraps system package managers (apt, pacman, apk, dnf, and more)
behind a single, consistent interface.`,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
