package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"time"
)

type Command int

const (
	CommandInvalid Command = iota
	CommandRunDebug
	CommandRun
	CommandCompile
	CommandApp
)

var command Command

const (
	ViteAddress       = "localhost:5173"
	CheckViteTimeout  = 1 * time.Second
	CheckViteInterval = 500 * time.Millisecond
	CheckViteRetries  = 5
)

func init() {
	flag.Parse()

	cmd := flag.Arg(0)

	switch cmd {
	case "debug":
		command = CommandRunDebug
	case "run":
		command = CommandRun
	case "compile":
		command = CommandCompile
	case "app":
		command = CommandApp
	default:
		slog.Error("Invalid command. Use debug, run, compile, or app.")
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
	case CommandApp:
		app()
	default:
		fmt.Println("Invalid command. Use -debug, -run, or -compile.")
	}
}

// ProcessGroup manages multiple child processes.
type ProcessGroup struct {
	cleanupOnce sync.Once
	processes   []*ProcessHandle
	mx          sync.Mutex
}

// ProcessHandle holds a started process with its control functions.
type ProcessHandle struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

// Add adds a new process to the group.
func (pg *ProcessGroup) Add(p *ProcessHandle) {
	pg.mx.Lock()
	defer pg.mx.Unlock()
	pg.processes = append(pg.processes, p)
}

// Cleanup cancels all processes in the group. It will only run once.
func (pg *ProcessGroup) Cleanup() {
	pg.cleanupOnce.Do(func() {
		for _, p := range pg.processes {
			p.cancel()
		}
	})
}

func (pg *ProcessGroup) Wait() {
	wg := &sync.WaitGroup{}
	for _, p := range pg.processes {
		wg.Add(1)
		go func(proc *ProcessHandle) {
			defer wg.Done()
			if err := proc.cmd.Wait(); err != nil {
				pg.mx.Lock()
				// remove the process from the list
				pg.processes = slices.DeleteFunc(pg.processes, func(e *ProcessHandle) bool {
					return e.cmd.Process.Pid == proc.cmd.Process.Pid
				})
				pg.mx.Unlock()
				fmt.Printf("Process (pid=%d) exited with error: %v\n", proc.cmd.Process.Pid, err)
			} else {
				pg.mx.Lock()
				// remove the process from the list
				pg.processes = slices.DeleteFunc(pg.processes, func(e *ProcessHandle) bool {
					return e.cmd.Process.Pid == proc.cmd.Process.Pid
				})
				pg.mx.Unlock()
				fmt.Printf("Process (pid=%d) exited.\n", proc.cmd.Process.Pid)
			}
			// trigger cleanup
			pg.Cleanup()
		}(p)
	}
	wg.Wait()
}

func cmd(id, command string) {
	fmt.Printf("Running (%s): %s\n", id, command)
	c := exec.Command("bash", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Start(); err != nil {
		fmt.Printf("cmd (%s) command failed: %s\n", id, err.Error())
		runtime.Goexit()
	}
	if err := c.Wait(); err != nil {
		fmt.Printf("cmd (%s) wait failed: %s\n", id, err.Error())
		runtime.Goexit()
	}
}

func startProcess(command string) (*ProcessHandle, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("command start failed: %w", err)
	}
	fmt.Printf("Started %d: %s\n", cmd.Process.Pid, command)

	return &ProcessHandle{
		cmd: cmd,
		cancel: func() {
			fmt.Printf("cancelling cmd pid: %d\n", cmd.Process.Pid)
			pgid, err := syscall.Getpgid(cmd.Process.Pid)
			if err != nil {
			}
			if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
				fmt.Printf("kill failed (%d): %s\n", cmd.Process.Pid, err.Error())
			}
			cancel()
		},
	}, nil
}

func compile() {
	cmd("04", "npm run build --prefix ./manager")
	cmd("05", "go build -o ./proxy/proxy-app ./proxy")
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

func app() {
	cmd("01", "npm run build --prefix ./manager --outDir vitedist")
	cmd("02", "go build -o ./proxy/proxy-app ./proxy")
	cmd("03", "cd manager && npx electron-builder build")
	cmd("04", "cp ./proxy/proxy-app ./manager/dist/mac-arm64/cap.app/Contents/Resources")
	cmd("05", "cp -R ./manager/vitedist ./manager/dist/mac-arm64/cap.app/Contents/Resources/dist")
	cmd("06", "cp -R ./manager/dist/mac-arm64/cap.app .")
}

func run() {
	// run requires compilation.
	compile()

	pg := new(ProcessGroup)
	handleSignals(pg.Cleanup)

	proxyApp, err := startProcess("DEBUG=false ./proxy/proxy-app")
	if err != nil {
		fmt.Println("error starting proxy app:", err)
		return
	}
	pg.Add(proxyApp)
	defer pg.Cleanup()

	electron, err := startProcess("DEBUG=false BUILT=false electron ./manager")
	if err != nil {
		fmt.Println("error starting proxy app:", err)
		return
	}

	pg.Add(electron)
	pg.Wait()
}

func waitForVite() bool {
	for range CheckViteRetries {
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

	group := &ProcessGroup{}
	handleSignals(group.Cleanup)

	// Start proxy
	proxy, err := startProcess("DEBUG=true go run ./proxy")
	if err != nil {
		fmt.Println("Error starting proxy:", err)
		return
	}
	group.Add(proxy)
	defer group.Cleanup()

	// Start Vite
	vite, err := startProcess("DEBUG=true npm run dev --prefix ./manager")
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
	electron, err := startProcess("DEBUG=true BUILT=false electron ./manager")
	if err != nil {
		fmt.Println("Error starting Electron:", err)
		return
	}
	group.Add(electron)

	group.Wait()
}
