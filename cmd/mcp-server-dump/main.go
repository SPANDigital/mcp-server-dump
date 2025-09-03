package main

import (
	"log"

	"github.com/alecthomas/kong"

	"github.com/spandigital/mcp-server-dump/internal/app"
)

func main() {
	var cli app.CLI
	ctx := kong.Parse(&cli, kong.Vars{"version": app.GetVersion()})

	if err := app.Run(&cli); err != nil {
		log.Fatalf("Error: %v", err)
	}

	_ = ctx
}
