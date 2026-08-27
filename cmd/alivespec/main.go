package main

import (
	"fmt"
	"os"

	"github.com/0xRech/AliveSpec/internal/diff"
	"github.com/0xRech/AliveSpec/internal/learn"
	"github.com/0xRech/AliveSpec/internal/verify"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "learn":
		err = learn.Run(os.Args[2:])
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Print(`AliveSpec — runtime-to-spec operational contracts

Usage:
  alivespec learn  [flags]
  alivespec verify <contract.yaml>
  alivespec diff   <before.yaml> <after.yaml>
  alivespec version
`)
}
