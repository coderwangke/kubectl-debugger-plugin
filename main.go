package main

import (
	"github.com/coderwangke/kubectl-debugger-plugin/cmd"
	"log"
	"os"
)

func main() {
	// Execute
	if err := cmd.NewRootCmd().Execute(); err != nil {
		log.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
