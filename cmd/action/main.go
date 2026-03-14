package main

import (
	"fmt"
	"os"

	"github.com/saintedlama/goship/internal/action"
)

func main() {
	cfg := action.Config{
		Token:            getInput("GITHUB_TOKEN"),
		WorkingDirectory: getInputWithDefault("WORKING_DIRECTORY", "."),
		Test:             action.ParseBool(getInputWithDefault("TEST", "true")),
		Coverage:         action.ParseBool(getInputWithDefault("COVERAGE", "true")),
		Vet:              action.ParseBool(getInputWithDefault("VET", "true")),
		Fmt:              action.ParseBool(getInputWithDefault("FMT", "true")),
	}

	passed, err := action.Run(cfg)
	if err != nil {
		writeError(err.Error())
		os.Exit(1)
	}
	if !passed {
		os.Exit(1)
	}
}

func getInput(key string) string {
	return os.Getenv(key)
}

func getInputWithDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func setOutput(key, value string) error {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		fmt.Printf("::set-output name=%s::%s\n", key, value)
		return nil
	}
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open GITHUB_OUTPUT: %w", err)
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", key, value)
	return err
}

func writeError(msg string) {
	fmt.Fprintf(os.Stderr, "::error::%s\n", msg)
}

var _ = setOutput
