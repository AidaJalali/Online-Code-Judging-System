package code_runner

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ExecutionResult represents the result of code execution
type ExecutionResult struct {
	Output      string
	Error       string
	ExitCode    int
	TimeUsed    time.Duration
	MemoryUsed  int64 // in bytes
	CompileTime time.Duration
}

// Language represents a programming language supported by the code runner
type Language string

const (
	Python Language = "python"
	Java   Language = "java"
	Cpp    Language = "cpp"
)

// CodeRunner is the interface for the code execution service
type CodeRunner interface {
	// ExecuteCode runs the given code with the specified language and input
	ExecuteCode(ctx context.Context, code string, language Language, input string) (*ExecutionResult, error)

	// Cleanup removes any temporary files or resources
	Cleanup() error
}

// Config holds the configuration for the code runner
type Config struct {
	// TimeLimit is the maximum time allowed for code execution
	TimeLimit time.Duration

	// MemoryLimit is the maximum memory allowed for code execution (in bytes)
	MemoryLimit int64

	// WorkDir is the directory where temporary files will be stored
	WorkDir string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		TimeLimit:   2 * time.Second,
		MemoryLimit: 256 * 1024 * 1024, // 256 MB
		WorkDir:     filepath.Join(os.TempDir(), "code_runner"),
	}
}

// NewCodeRunner creates a new instance of the code runner
func NewCodeRunner(config *Config) (CodeRunner, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create work directory if it doesn't exist
	if err := os.MkdirAll(config.WorkDir, 0755); err != nil {
		return nil, err
	}

	return &codeRunner{
		config: config,
	}, nil
}

// codeRunner is the implementation of the CodeRunner interface
type codeRunner struct {
	config *Config
}

// ExecuteCode implements the CodeRunner interface
func (cr *codeRunner) ExecuteCode(ctx context.Context, code string, language Language, input string) (*ExecutionResult, error) {
	// Create a temporary file for the code
	tempFile, err := cr.createTempFile(code, language)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile)

	// Compile the code if needed
	executable, compileTime, err := cr.compileCode(tempFile, language)
	if err != nil {
		return nil, err
	}
	if executable != "" {
		defer os.Remove(executable)
	}

	// Execute the code
	result, err := cr.executeCode(ctx, executable, language, input)
	if err != nil {
		return nil, err
	}

	result.CompileTime = compileTime
	return result, nil
}

// Cleanup implements the CodeRunner interface
func (cr *codeRunner) Cleanup() error {
	return os.RemoveAll(cr.config.WorkDir)
}

// createTempFile creates a temporary file with the given code
func (cr *codeRunner) createTempFile(code string, language Language) (string, error) {
	var extension string
	switch language {
	case Python:
		extension = ".py"
	case Java:
		extension = ".java"
	case Cpp:
		extension = ".cpp"
	default:
		return "", errors.New("unsupported language")
	}

	// Create a temporary file in the work directory
	tempFile, err := os.CreateTemp(cr.config.WorkDir, "code_*"+extension)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	// Write the code to the file
	if _, err := tempFile.WriteString(code); err != nil {
		os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// compileCode compiles the code if needed and returns the path to the executable
func (cr *codeRunner) compileCode(filePath string, language Language) (string, time.Duration, error) {
	startTime := time.Now()

	switch language {
	case Python:
		// Python doesn't need compilation
		return filePath, 0, nil

	case Java:
		// Compile Java code
		cmd := exec.Command("javac", filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", 0, errors.New(string(output))
		}

		// Return the path to the .class file
		classFile := filePath[:len(filePath)-5] + ".class"
		return classFile, time.Since(startTime), nil

	case Cpp:
		// Compile C++ code
		outputFile := filePath[:len(filePath)-4] // Remove .cpp extension
		cmd := exec.Command("g++", "-o", outputFile, filePath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", 0, errors.New(string(output))
		}

		return outputFile, time.Since(startTime), nil

	default:
		return "", 0, errors.New("unsupported language")
	}
}

// executeCode runs the code and returns the execution result
func (cr *codeRunner) executeCode(ctx context.Context, executable string, language Language, input string) (*ExecutionResult, error) {
	var cmd *exec.Cmd

	switch language {
	case Python:
		cmd = exec.CommandContext(ctx, "python", executable)
	case Java:
		// Get the class name from the file path
		className := filepath.Base(executable)
		className = className[:len(className)-6] // Remove .class extension
		cmd = exec.CommandContext(ctx, "java", "-cp", filepath.Dir(executable), className)
	case Cpp:
		cmd = exec.CommandContext(ctx, executable)
	default:
		return nil, errors.New("unsupported language")
	}

	// Set up input/output pipes
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	defer stdin.Close()

	// Write input to stdin
	go func() {
		io.WriteString(stdin, input)
		stdin.Close()
	}()

	// Start the process
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Create a channel to receive the result
	resultChan := make(chan *ExecutionResult, 1)
	errorChan := make(chan error, 1)

	// Monitor the process
	go func() {
		// Wait for the process to finish
		err := cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				resultChan <- &ExecutionResult{
					ExitCode: exitErr.ExitCode(),
					Error:    string(exitErr.Stderr),
				}
				return
			}
			errorChan <- err
			return
		}

		// Get the output
		output, _ := cmd.CombinedOutput()
		resultChan <- &ExecutionResult{
			Output:   string(output),
			ExitCode: 0,
		}
	}()

	// Wait for either the result or an error
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-ctx.Done():
		// Kill the process
		cmd.Process.Kill()
		return nil, ctx.Err()
	}
}
