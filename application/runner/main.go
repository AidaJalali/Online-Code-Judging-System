package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

type Result struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	TimeMs   int64  `json:"time_ms"`
	Error    string `json:"error,omitempty"`
}

func main() {
	codePath := "/code/code.go"
	inputPath := "/code/input.txt"
	resultPath := "/code/result.json"
	binPath := "/code/bin"

	timeoutSec := 2 // default
	if tStr := os.Getenv("TIMEOUT_SEC"); tStr != "" {
		if t, err := strconv.Atoi(tStr); err == nil {
			timeoutSec = t
		}
	}

	// Compile
	cmd := exec.Command("go", "build", "-o", binPath, codePath)
	var buildStderr bytes.Buffer
	cmd.Stderr = &buildStderr
	if err := cmd.Run(); err != nil {
		writeResult(resultPath, Result{Stderr: buildStderr.String(), ExitCode: 1, Error: "build failed"})
		os.Exit(1)
	}

	// Run
	input, _ := ioutil.ReadFile(inputPath)
	runCmd := exec.Command(binPath)
	runCmd.Stdin = bytes.NewReader(input)
	var outBuf, errBuf bytes.Buffer
	runCmd.Stdout = &outBuf
	runCmd.Stderr = &errBuf
	start := time.Now()
	done := make(chan error, 1)
	go func() { done <- runCmd.Run() }()
	var err error
	select {
	case err = <-done:
		// finished
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		runCmd.Process.Kill()
		err = &TimeoutError{}
	}
	dur := time.Since(start)
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
			}
		}
	}
	result := Result{
		Stdout:   outBuf.String(),
		Stderr:   errBuf.String(),
		ExitCode: exitCode,
		TimeMs:   dur.Milliseconds(),
	}
	if err != nil {
		result.Error = err.Error()
	}
	writeResult(resultPath, result)
}

type TimeoutError struct{}

func (t *TimeoutError) Error() string { return "timeout" }

func writeResult(path string, r Result) {
	b, _ := json.Marshal(r)
	_ = ioutil.WriteFile(path, b, 0644)
}
