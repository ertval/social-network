package main

import (
	"flag"
	"fmt"
	"os"
	"social-network/internal/gates"
)

func main() {
	all := flag.Bool("all", false, "run all gates")
	gate := flag.String("gate", "", "run a specific gate by name")
	flag.Parse()

	runner := gates.NewRunner()

	// Register all gates
	runner.Register(&gates.StackGate{})
	runner.Register(&gates.LayoutGate{})
	runner.Register(&gates.BoundariesGate{})
	runner.Register(&gates.DAGGate{})
	runner.Register(&gates.TDDGate{})
	runner.Register(&gates.MigrationsGate{})
	runner.Register(&gates.SecurityGate{})
	runner.Register(&gates.BranchGate{})
	runner.Register(&gates.ScopeDriftGate{})
	runner.Register(&gates.CoverageGate{})

	if *gate != "" {
		result, err := runner.RunOne(*gate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		report := gates.Report{Overall: result.Status, Gates: []gates.Result{result}}
		if err := gates.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			os.Exit(2)
		}
		if result.Status == "FAIL" {
			os.Exit(1)
		}
		return
	}

	if *all || flag.NArg() == 0 {
		report := runner.RunAll()
		if err := gates.WriteJSON(os.Stdout, report); err != nil {
			fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
			os.Exit(2)
		}
		if report.Overall == "FAIL" {
			os.Exit(1)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "usage: gates --all | --gate=<name>")
	os.Exit(2)
}
