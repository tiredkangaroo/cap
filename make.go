package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

const ViteAddress = "localhost:5173"
const CheckViteTimeout = 1 * time.Second
const CheckViteInterval = 500 * time.Millisecond
const CheckViteRetries = 5

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
		println("Invalid command. Use -debug, -run, or -compile.")
	}
}

func cmdWaitKill(id, cmd string) (wait func() error, kill func() error) {
	fmt.Printf("running (%s) (with waitkill): %s\n", id, cmd)
	c := exec.Command("bash", "-c", cmd)
	if err := c.Start(); err != nil {
		println("command failed: ", err.Error())
		runtime.Goexit()
	}

	out, err := c.StdoutPipe()
	if err != nil {
		println("stdout pipe failed: ", err.Error())
	}
	go func() {
		defer out.Close()
		io.Copy(os.Stdout, out)
	}()

	wait = c.Wait
	kill = func() error {
		if err := c.Process.Kill(); err != nil {
			return fmt.Errorf("cmd (%s) kill failed: %w", id, err)
		}
		return nil
	}
	return wait, kill
}

func cmd(id, command string) {
	fmt.Printf("running (%s): %s\n", id, command)
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

// whenOneKillAll waits for the first command to finish and kills all other commands.
func whenOneKillAll(f ...func() error) {
	killAll := sync.OnceFunc(func() {
		for j := range len(f) / 2 {
			killFunc := f[2*j+1]
			if err := killFunc(); err != nil {
				fmt.Printf("error killing command (index %d): %s\n", j, err.Error())
			}
		}
	})

	var wg = &sync.WaitGroup{}
	for i := range len(f) / 2 {
		waitFunc := f[2*i]
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := waitFunc(); err != nil {
				fmt.Printf("error waiting for command (index %d): %s\n", i, err.Error())
			}
			killAll()
		}()
	}

	wg.Wait()
}

// run runs the program in production mode.
func run() {
	println("Running in production mode...")
	// Add your production run logic here
}

// compile compiles the program.
func compile() {
	cmd("04", "npm run build --prefix ./manager")
	cmd("05", "go build -o ./proxy/proxy-app ./proxy")
}

// runDebug runs the program in debug mode.
func runDebug() {
	println("Running in debug mode...")
	runProxyWait, runProxyKill := cmdWaitKill("01", "DEBUG=true go run ./proxy")
	runNPMWait, runNPMKill := cmdWaitKill("02", "DEBUG=true npm run dev --prefix ./manager")

	working := false
	for range 5 {
		_, err := net.DialTimeout("tcp", ViteAddress, time.Second)
		if err == nil {
			working = true
			break
		}
		time.Sleep(CheckViteInterval)
	}

	if !working {
		println("Vite server is not running. Exiting...")
		return
	}

	runElectronWait, runElectronKill := cmdWaitKill("03", "DEBUG=true electron ./manager")

	defer whenOneKillAll(runProxyWait, runProxyKill, runNPMWait, runNPMKill, runElectronWait, runElectronKill)
}
