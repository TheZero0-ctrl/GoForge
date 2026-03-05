package main

import (
	"context"
	"os"

	"goforge/internal/cli"
)

func main() {
	code := cli.Run(context.Background(), os.Stdout, os.Stderr)
	os.Exit(code)
}
