package main

import (
	"context"
	"os"

	"github.com/charmbracelet/fang"

	"ayo/internal/version"
)

func main() {
	ctx := context.Background()
	cmd := newRootCmd()
	if err := fang.Execute(ctx, cmd, fang.WithVersion(version.Version)); err != nil {
		os.Exit(1)
	}
}
