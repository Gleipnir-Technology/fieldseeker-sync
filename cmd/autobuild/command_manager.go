package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

type ProcessManager struct {
	cmd         *exec.Cmd
	processLock sync.Mutex
	stopChan    chan struct{}
	processName string
	processArgs []string
}

func NewProcessManager(processName string, args []string) *ProcessManager {
	return &ProcessManager{
		processName: processName,
		processArgs: args,
		stopChan:    make(chan struct{}),
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

	// Handle stdout in a goroutine
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stdout.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// Handle stderr in a goroutine
	go func() {
		buffer := make([]byte, 1024)
		for {
			n, err := stderr.Read(buffer)
			if n > 0 {
				fmt.Print(string(buffer[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	// Wait for the process in a goroutine
	go func() {
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

	if pm.cmd != nil && pm.cmd.Process != nil {
		return pm.cmd.Process.Kill()
	}
	return nil
}
