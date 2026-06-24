package main

import (
	"flag"
	"fmt"
	"os"

	"social-network/internal/gates"
)

//nolint:nestif
func main() {
	all := flag.Bool("all", false, "run all gates")
	gate := flag.String("gate", "", "run a specific gate by name")
	jsonOutput := flag.Bool("json", false, "output in JSON format")
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
	runner.Register(&gates.FormatGate{})
	runner.Register(&gates.LintGate{})
	runner.Register(&gates.UnitTestGate{})
	runner.Register(&gates.FrontendGate{})

	if *gate != "" {
		result, err := runner.RunOne(*gate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		report := gates.Report{Overall: result.Status, Gates: []gates.Result{result}}

		if *jsonOutput {
			if err := gates.WriteJSON(os.Stdout, report); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
				os.Exit(2)
			}
		} else {
			printTextReport(report)
		}

		if result.Status == "FAIL" {
			os.Exit(1)
		}
		return
	}

	if *all || flag.NArg() == 0 {
		if !*jsonOutput {
			runner.OnResult = func(result gates.Result) {
				printResult(result)
			}
		}

		report := runner.RunAll()

		if *jsonOutput {
			if err := gates.WriteJSON(os.Stdout, report); err != nil {
				fmt.Fprintf(os.Stderr, "error writing JSON: %v\n", err)
				os.Exit(2)
			}
		}

		if report.Overall == "FAIL" {
			os.Exit(1)
		}
		return
	}

	fmt.Fprintln(os.Stderr, "usage: gates --all | --gate=<name> [--json]")
	os.Exit(2)
}

func printResult(result gates.Result) {
	switch result.Status {
	case "PASS":
		fmt.Printf("[PASS] %s\n", result.Gate)
	case "FAIL":
		fmt.Printf("[FAIL] %s: %s\n", result.Gate, result.Message)
	case "SKIP":
		fmt.Printf("[SKIP] %s: %s\n", result.Gate, result.Message)
	default:
		fmt.Printf("[%s] %s: %s\n", result.Status, result.Gate, result.Message)
	}
}

func printTextReport(report gates.Report) {
	for _, result := range report.Gates {
		printResult(result)
	}
}
