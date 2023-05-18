package main

import (
	"fmt"
	"github.com/coderwangke/kubectl-debugger-plugin/cmd"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use: "kubectl-debugger",
	Short: "A kubectl plugin to create debugger pods",
}

func main() {
	rootCmd.AddCommand(cmd.NewPodCmd())
	rootCmd.AddCommand(cmd.NewNodeCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
