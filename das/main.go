package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := cobra.Command{
		Use:     "das",
		Aliases: []string{"das"},
	}

	rootCmd.AddCommand(
		sampleCmd(),
		initCmd(),
		addHydraCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
