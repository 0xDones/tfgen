package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type TestEvent struct {
	Time    time.Time
	Action  string
	Package string
	Test    string
	Elapsed float64
	Output  string
}

type TestResult struct {
	Name    string
	Package string
	Status  string
	Elapsed float64
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	testResults := make(map[string]*TestResult)
	packageResults := make(map[string]string)

	// Read JSON test output line by line
	for scanner.Scan() {
		var event TestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		// Track test results
		if event.Test != "" {
			key := event.Package + "::" + event.Test
			if event.Action == "run" {
				testResults[key] = &TestResult{
					Name:    event.Test,
					Package: event.Package,
					Status:  "running",
				}
			} else if event.Action == "pass" || event.Action == "fail" || event.Action == "skip" {
				if result, exists := testResults[key]; exists {
					result.Status = event.Action
					result.Elapsed = event.Elapsed
				} else {
					testResults[key] = &TestResult{
						Name:    event.Test,
						Package: event.Package,
						Status:  event.Action,
						Elapsed: event.Elapsed,
					}
				}
			}
		}

		// Track package results
		if event.Test == "" && (event.Action == "pass" || event.Action == "fail" || event.Action == "skip") {
			packageResults[event.Package] = event.Action
		}
	}

	// Get GitHub Step Summary file path
	summaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryFile == "" {
		fmt.Println("GITHUB_STEP_SUMMARY not set, skipping summary generation")
		return
	}

	// Open the summary file
	f, err := os.OpenFile(summaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening summary file: %v\n", err)
		return
	}
	defer f.Close()

	// Write summary header
	fmt.Fprintf(f, "## 🧪 Test Results\n\n")

	// Count results
	passed := 0
	failed := 0
	skipped := 0
	for _, result := range testResults {
		switch result.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "skip":
			skipped++
		}
	}

	// Write summary stats
	fmt.Fprintf(f, "**Summary:** ")
	if failed > 0 {
		fmt.Fprintf(f, "❌ %d failed, ", failed)
	}
	if passed > 0 {
		fmt.Fprintf(f, "✅ %d passed, ", passed)
	}
	if skipped > 0 {
		fmt.Fprintf(f, "⏭️  %d skipped, ", skipped)
	}
	fmt.Fprintf(f, "**Total:** %d tests\n\n", len(testResults))

	// Write test results table
	if len(testResults) > 0 {
		fmt.Fprintf(f, "### Test Cases\n\n")
		fmt.Fprintf(f, "| Status | Package | Test | Duration |\n")
		fmt.Fprintf(f, "|--------|---------|------|----------|\n")

		for _, result := range testResults {
			statusIcon := ""
			switch result.Status {
			case "pass":
				statusIcon = "✅"
			case "fail":
				statusIcon = "❌"
			case "skip":
				statusIcon = "⏭️"
			}
			duration := fmt.Sprintf("%.3fs", result.Elapsed)
			fmt.Fprintf(f, "| %s | %s | %s | %s |\n", statusIcon, result.Package, result.Name, duration)
		}
		fmt.Fprintf(f, "\n")
	}

	// Write package results
	if len(packageResults) > 0 {
		fmt.Fprintf(f, "### Package Results\n\n")
		fmt.Fprintf(f, "| Status | Package |\n")
		fmt.Fprintf(f, "|--------|---------|\n")

		for pkg, status := range packageResults {
			statusIcon := ""
			switch status {
			case "pass":
				statusIcon = "✅"
			case "fail":
				statusIcon = "❌"
			case "skip":
				statusIcon = "⏭️"
			}
			fmt.Fprintf(f, "| %s | %s |\n", statusIcon, pkg)
		}
	}
}
