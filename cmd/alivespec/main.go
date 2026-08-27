package main

import (
	"fmt"
	"os"

	"github.com/0xRech/AliveSpec/internal/diff"
	"github.com/0xRech/AliveSpec/internal/learn"
	"github.com/0xRech/AliveSpec/internal/record"
	"github.com/0xRech/AliveSpec/internal/verify"
)

const version = "0.2.0-alpha.1"

type quietError interface {
	error
	Quiet() bool
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "learn":
		err = learn.Run(os.Args[2:])
	case "record":
		err = record.Run(os.Args[2:])
	case "verify":
		err = verify.Run(os.Args[2:])
	case "diff":
		err = diff.Run(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("AliveSpec %s\n", version)
		return
	case "help", "--help", "-h":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}

	if err != nil {
		if quiet, ok := err.(quietError); !ok || !quiet.Quiet() {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`AliveSpec — learn executable operational contracts from healthy systems

Usage:
  alivespec record [journey] [flags]        Observe a successful runtime journey (Linux/eBPF)
  alivespec learn  [flags]                  Build a contract from explicit hints
  alivespec verify <contract.yaml>          Verify a contract against the current system
  alivespec diff   <before.yaml> <after.yaml>
  alivespec version

Examples:
  sudo alivespec record login --comm nginx --comm myapp
  alivespec verify login.alivespec.yaml
`)
}
