package main

import (
	"fmt"
	"os"

	"github.com/craftslab/cleansource-sca-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
