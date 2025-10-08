package ansible

import (
	"fmt"
	"strings"
	"time"
)

// ExecutionResult tracks the result of a playbook execution
type ExecutionResult struct {
	PlaybookName string
	Status       string // "success", "failed", "skipped"
	Duration     time.Duration
	Error        error
}

// ExecutionSummary holds multiple execution results
type ExecutionSummary struct {
	Results []ExecutionResult
}

// Add adds a result to the summary
func (s *ExecutionSummary) Add(result ExecutionResult) {
	s.Results = append(s.Results, result)
}

// HasFailures returns true if any execution failed
func (s *ExecutionSummary) HasFailures() bool {
	for _, r := range s.Results {
		if r.Status == "failed" {
			return true
		}
	}
	return false
}

// SuccessCount returns the number of successful executions
func (s *ExecutionSummary) SuccessCount() int {
	count := 0
	for _, r := range s.Results {
		if r.Status == "success" {
			count++
		}
	}
	return count
}

// Print displays the execution summary
func (s *ExecutionSummary) Print() {
	if len(s.Results) == 0 {
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("PLAYBOOK EXECUTION SUMMARY")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("%-40s %-10s %-15s\n", "PLAYBOOK", "STATUS", "DURATION")
	fmt.Println(strings.Repeat("-", 70))

	for _, r := range s.Results {
		status := r.Status
		if r.Status == "success" {
			status = "OK"
		} else if r.Status == "failed" {
			status = "FAILED"
		}
		fmt.Printf("%-40s %-10s %-15s\n", r.PlaybookName, status, r.Duration.Round(time.Second))
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Printf("Total: %d | Success: %d | Failed: %d\n",
		len(s.Results),
		s.SuccessCount(),
		len(s.Results)-s.SuccessCount())
	fmt.Println(strings.Repeat("=", 70))
}
