package main

import (
	"context"
	"fmt"
	"os"

	"mdnsmap/internal/cli"
	"mdnsmap/internal/output"
	"mdnsmap/internal/service"
)

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	result := service.NewScanService().Run(context.Background(), cfg)
	if err := output.WriteText(os.Stdout, result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
