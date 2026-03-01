package main

import (
	"flag"
	"fmt"
	"os"
)

const usageText = `usage: ft_turing [-h] jsonfile input
positional arguments:
  jsonfile   json description of the machine
  input      input of the machine
optional arguments:
  -h, --help show this help message and exit
`

func printUsage() {
	fmt.Fprint(os.Stderr, usageText)
}

func main() {
	// Custom flag set so we can support both -h and --help
	fs := flag.NewFlagSet("ft_turing", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	help := fs.Bool("h", false, "show this help message and exit")
	helpLong := fs.Bool("help", false, "show this help message and exit")

	// We don't actually define any other flags, just parse the args
	if err := fs.Parse(os.Args[1:]); err != nil {
		printUsage()
		os.Exit(1)
	}

	if *help || *helpLong {
		printUsage()
		return
	}

	args := fs.Args()
	if len(args) != 2 {
		printUsage()
		os.Exit(1)
	}

	jsonPath := args[0]
	input := args[1]

	machine, err := LoadMachineFromFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := machine.ValidateInput(input); err != nil {
		fmt.Fprintf(os.Stderr, "invalid input: %v\n", err)
		os.Exit(1)
	}

	if err := RunMachine(machine, input, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		os.Exit(1)
	}
}



