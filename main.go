package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SCKelemen/cpkg/internal/cmd"
)

func main() {
	app := cmd.NewApp()
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
