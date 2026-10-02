//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("=== Running Go Tests ===")
	fmt.Println()

	// Run go test
	fmt.Println("[1/4] Running go test...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("✗ Tests failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Tests passed")
	fmt.Println()

	// Run go vet
	fmt.Println("[2/4] Running go vet...")
	cmd = exec.Command("go", "vet", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠ Go vet found issues: %v\n", err)
	} else {
		fmt.Println("✓ Go vet passed")
	}
	fmt.Println()

	// Run go fmt check
	fmt.Println("[3/4] Running go fmt...")
	cmd = exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("⚠ Go fmt failed: %v\n", err)
	} else {
		fmt.Println("✓ Go fmt passed")
	}
	fmt.Println()

	// Build
	fmt.Println("[4/4] Building application...")
	cmd = exec.Command("go", "build", "-o", "../bin/presensigo-test.exe", "./cmd/api")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("✗ Build failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Build passed")
	fmt.Println()

	fmt.Println("=== All automated tests PASSED! ===")
}
