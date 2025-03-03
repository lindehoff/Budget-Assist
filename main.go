/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"os"

	"github.com/lindehoff/Budget-Assist/cmd"
)

func main() {
	fmt.Fprintf(os.Stderr, "Starting Budget-Assist...\n")
	cmd.Execute()
}
