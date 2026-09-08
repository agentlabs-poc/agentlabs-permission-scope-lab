package main

import (
	"agentlabs.local/abv/cli"
	"agentlabs.local/abv/internal/lab"
	"context"
	"os"
)

func main() {
	os.Exit(cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, lab.Connect, lab.Scenarios{}, lab.ConnectCatalog))
}
