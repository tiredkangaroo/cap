package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sync"
	"syscall"
)

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
