package main

import (
	"os"

	"sdlaic/cmd"
)

// version is set at build time via:
// go build -ldflags "-X main.version=v0.1.0"
var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
