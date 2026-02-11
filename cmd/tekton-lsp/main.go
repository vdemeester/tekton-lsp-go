package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tliron/commonlog"
	_ "github.com/tliron/commonlog/simple"

	"github.com/vdemeester/tekton-lsp-go/pkg/server"
)

// Set via -ldflags at build time.
var version = "dev"

const name = "tekton-lsp"

var (
	showVersion = flag.Bool("version", false, "Show version and exit")
	logLevel    = flag.Int("log-level", 0, "Logging verbosity (0=errors, 1=info, 2=debug, 3=trace)")
	tcp         = flag.String("tcp", "", "Use TCP transport (e.g., 'localhost:8080')")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", name, version)
		os.Exit(0)
	}

	// Configure logging
	commonlog.Configure(*logLevel, nil)

	// Create and run LSP server
	srv := server.New(name, version)

	if *tcp != "" {
		if err := srv.RunTCP(*tcp); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TCP server: %v\n", err)
			os.Exit(1)
		}
	} else {
		if err := srv.RunStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running stdio server: %v\n", err)
			os.Exit(1)
		}
	}
}
