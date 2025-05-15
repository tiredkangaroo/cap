package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const (
	ViteAddress       = "localhost:5173"
	CheckViteTimeout  = 1 * time.Second
	CheckViteInterval = 500 * time.Millisecond
	CheckViteRetries  = 5
)

type Command int

const (
	CommandInvalid Command = iota
	CommandRunDebug
	CommandRun
	CommandCompile
)

var command Command

func init() {
	runDebug := flag.Bool("debug", false, "Run the program (debug mode)")
	run := flag.Bool("run", false, "Run the program (production mode)")
	compile := flag.Bool("compile", false, "Compile the program")
	flag.Parse()

	if *runDebug {
		command = CommandRunDebug
	}
	if *run {
		command = CommandRun
	}
	if *compile {
		command = CommandCompile
	}
}

func main() {
	switch command {
	case CommandRunDebug:
		runDebug()
	case CommandRun:
		run()
	case CommandCompile:
		compile()
	default:
		fmt.Println("Invalid command. Use -debug, -run, or -compile.")
	}
}

// ProcessHandle holds a started process with its control functions
type ProcessHandle struct {
	cmd    *exec.Cmd
	wait   func() error
	cancel context.CancelFunc
	kill   func() error
}

// ProcessGroup manages multiple child processes
type ProcessGroup struct {
	mu          sync.Mutex
	cleanupOnce sync.Once
	processes   []*ProcessHandle
}

func (pg *ProcessGroup) Add(p *ProcessHandle) {
	pg.mu.Lock()
	defer pg.mu.Unlock()
	pg.processes = append(pg.processes, p)
}

func (pg *ProcessGroup) Cleanup() {
	pg.cleanupOnce.Do(func() {
		fmt.Println("cleaning up processes...")
		pg.mu.Lock()
		defer pg.mu.Unlock()
		for _, p := range pg.processes {
			p.cancel()
			if err := p.kill(); err != nil {
				fmt.Printf("failed to kill process (pid=%d): %v\n", p.cmd.Process.Pid, err)
			}
		}
	})
}

func handleSignals(cleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nSignal received. Cleaning up...")
		cleanup()
		os.Exit(1)
	}()
}

func startProcess(id, command string) (*ProcessHandle, error) {
	fmt.Printf("Starting (%s): %s\n", id, command)
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stdout pipe failed: %w", err)
	}
	go func() {
		defer stdout.Close()
		io.Copy(os.Stdout, stdout)
	}()

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("command start failed: %w", err)
	}

	return &ProcessHandle{
		cmd:    cmd,
		wait:   cmd.Wait,
		cancel: cancel,
		kill: func() error {
			return cmd.Process.Kill()
		},
	}, nil
}

func waitForVite() bool {
	for i := 0; i < CheckViteRetries; i++ {
		conn, err := net.DialTimeout("tcp", ViteAddress, CheckViteTimeout)
		if err == nil {
			conn.Close()
			return true
		}
		time.Sleep(CheckViteInterval)
	}
	return false
}

func runDebug() {
	fmt.Println("Running in debug mode...")

	var wg sync.WaitGroup
	group := &ProcessGroup{}
	handleSignals(group.Cleanup)

	// Start proxy
	proxy, err := startProcess("01", "DEBUG=true go run ./proxy")
	if err != nil {
		fmt.Println("Error starting proxy:", err)
		return
	}
	group.Add(proxy)
	defer group.Cleanup()

	// Start Vite
	vite, err := startProcess("02", "DEBUG=true npm run dev --prefix ./manager")
	if err != nil {
		fmt.Println("Error starting Vite:", err)
		return
	}
	group.Add(vite)

	if !waitForVite() {
		fmt.Println("Vite server did not start. Exiting...")
		return
	}

	// Start Electron
	electron, err := startProcess("03", "DEBUG=true electron ./manager")
	if err != nil {
		fmt.Println("Error starting Electron:", err)
		return
	}
	group.Add(electron)

	// Wait for any to finish, then cleanup all
	for _, p := range []*ProcessHandle{proxy, vite, electron} {
		wg.Add(1)
		go func(proc *ProcessHandle) {
			defer wg.Done()
			if err := proc.wait(); err != nil {
				fmt.Printf("Process (pid=%d) exited with error: %v\n", proc.cmd.Process.Pid, err)
			} else {
				fmt.Printf("Process (pid=%d) exited.\n", proc.cmd.Process.Pid)
			}
			group.Cleanup() // First one to exit triggers cleanup
		}(p)
	}

	wg.Wait()
}

func run() {
	fmt.Println("Running in production mode...")
	// Add your production run logic here
}

func compile() {
	cmd("04", "npm run build --prefix ./manager")
	cmd("05", "go build -o ./proxy/proxy-app ./proxy")
}

func cmd(id, command string) {
	fmt.Printf("Running (%s): %s\n", id, command)
	c := exec.Command("bash", "-c", command)
	out, err := c.StdoutPipe()
	if err != nil {
		fmt.Printf("cmd (%s) stdout pipe failed: %s\n", id, err.Error())
	}
	defer out.Close()
	if err := c.Start(); err != nil {
		fmt.Printf("cmd (%s) command failed: %s\n", id, err.Error())
		runtime.Goexit()
	}
	go io.Copy(os.Stdout, out)
	if err := c.Wait(); err != nil {
		fmt.Printf("cmd (%s) wait failed: %s\n", id, err.Error())
	}
}
