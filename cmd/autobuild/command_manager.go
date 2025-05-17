package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// ProcessOutput represents a single output line from the process
type ProcessOutput struct {
	Data     string
	IsStderr bool
	IsReset  bool
}

type ProcessManager struct {
	cmd         *exec.Cmd
	processLock sync.Mutex
	stopChan    chan struct{}
	processName string
	processArgs []string
	outputChan  chan ProcessOutput
}

func NewProcessManager(processName string, args []string) *ProcessManager {
	return &ProcessManager{
		processName: processName,
		processArgs: args,
		stopChan:    make(chan struct{}),
		outputChan:  make(chan ProcessOutput, 100), // Buffered channel to prevent blocking
	}
}

func (pm *ProcessManager) handleOutput(reader io.Reader, isStderr bool, wg *sync.WaitGroup) {
	defer wg.Done()

	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			output := ProcessOutput{
				Data:     string(buffer[:n]),
				IsStderr: isStderr,
			}

			select {
			case pm.outputChan <- output:
				// Output sent successfully
			case <-pm.stopChan:
				// Process is being stopped
				return
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from %s: %v",
					map[bool]string{true: "stderr", false: "stdout"}[isStderr],
					err)
			}
			return
		}
	}
}

func (pm *ProcessManager) startProcess(ctx context.Context) error {
	pm.processLock.Lock()
	defer pm.processLock.Unlock()

	// Create the command with context
	pm.cmd = exec.CommandContext(ctx, pm.processName, pm.processArgs...)

	// Set up pipes for stdout and stderr
	stdout, err := pm.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := pm.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Start the process
	if err := pm.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %v", err)
	}

	// Use WaitGroup to track output handling goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	// Handle stdout
	go pm.handleOutput(stdout, false, &wg)

	// Handle stderr
	go pm.handleOutput(stderr, true, &wg)

	// Wait for the process in a goroutine
	go func() {
		// Wait for output handlers to complete
		wg.Wait()
		// Close output channel when both handlers are done
		close(pm.outputChan)

		err := pm.cmd.Wait()
		if err != nil {
			log.Printf("Process exited with error: %v", err)
		}
	}()

	log.Printf("Started process %s with PID %d", pm.processName, pm.cmd.Process.Pid)
	return nil
}

func (pm *ProcessManager) stopProcess() error {
	pm.processLock.Lock()
	defer pm.processLock.Unlock()

	// Signal output handlers to stop
	close(pm.stopChan)
	message := ProcessOutput{
		Data:     "",
		IsStderr: false,
		IsReset:  true,
	}
	pm.outputChan <- message

	if pm.cmd != nil && pm.cmd.Process != nil {
		return pm.cmd.Process.Kill()
	}
	return nil
}

// outputProcessor handles the process output in a separate goroutine
func (pm *ProcessManager) outputProcessor(ctx context.Context, terminalChannel chan string) {
	stdout := ""
	stderr := ""
	for {
		select {
		case output, ok := <-pm.outputChan:
			if !ok {
				// Channel closed, exit processor
				return
			}
			if output.IsReset {
				stdout = ""
				stderr = ""
			} else if output.IsStderr {
				//fmt.Fprintf(os.Stderr, "STDERR: %s", output.Data)
				stderr += output.Data
				terminalChannel <- stderr
			} else {
				//fmt.Printf("STDOUT: %s", output.Data)
				stdout += output.Data
				terminalChannel <- stdout
			}
		case <-ctx.Done():
			return
		}
	}
}
