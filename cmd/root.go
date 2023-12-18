package cmd

import (
	"github.com/spf13/cobra"
)

var kubeConfig string

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kubectl-debugger",
		Short: "A kubectl plugin to create debugger pods",
	}

	flags := rootCmd.PersistentFlags()
	flags.StringVar(&kubeConfig, "kubeconfig", ".kube/config", "Path to the kubeconfig file")

	// Add subcommands
	rootCmd.AddCommand(newPodCmd())
	rootCmd.AddCommand(newNodeCmd())

	return rootCmd
}
