package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

var pm *ProcessManager
var target string
var ctx context.Context
var cancel context.CancelFunc

func main2() {
	//InitTerminal()
	//Loop()
	os.Exit(0)
	// Figure out what we are building
	flag.StringVar(&target, "target", "", "The ninja target to autobuild")
	flag.Parse()
	if target == "" {
		log.Fatal("Must specify a target")
	}

	out, err := exec.Command("ninja", "-t", "inputs", target).Output()
	if err != nil {
		log.Fatal(err)
	}

	// watch files for changes
	inputs := parseInputs(out)
	log.Println("Inputs: ", inputs)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range inputs {
		log.Println("Listening to", f)
		watcher.Add(f)
	}
	defer watcher.Close()

	// Example: run "ping" command with localhost
	// You can replace this with any other command
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	pm = NewProcessManager(target, []string{})
	if err := pm.startProcess(ctx); err != nil {
		log.Fatalf("Failed to start process: %v", err)
	}

	// Create a channel to receive errors
	done := make(chan bool)

	// Start watching for events
	go func() {
		// Debounce timer to prevent multiple rapid recompilations
		var debounceTimer *time.Timer
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Going away")
					return
				}
				log.Println("FS event", event)

				// Only trigger on write events
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Rename == fsnotify.Rename {
					if debounceTimer != nil {
						debounceTimer.Stop()
					}
					debounceTimer = time.AfterFunc(100*time.Millisecond, func() {
						recompile()
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("Going away")
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	fmt.Println("Watching for file changes. Press Ctrl+C to exit.")
	<-done
}

func parseInputs(out []byte) []string {
	content := string(out)
	results := make([]string, 0)
	parts := strings.Split(content, "\n")
	for _, p := range parts {
		if len(p) > 0 {
			results = append(results, p)
		}
	}
	return results
}

func recompile() {
	fmt.Println("\nFile change detected. Recompiling...")
	// Stop the current process
	if err := pm.stopProcess(); err != nil {
		log.Printf("Error stopping process: %v", err)
	}

	cmd := exec.Command("ninja")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Compilation failed: %v\n", err)
		return
	}

	fmt.Println("Compilation successful!")

	// Cancel the old context and create a new one
	cancel()
	ctx, cancel = context.WithCancel(context.Background())

	// Start the process again
	if err := pm.startProcess(ctx); err != nil {
		log.Fatalf("Failed to restart process: %v", err)
	}
}
