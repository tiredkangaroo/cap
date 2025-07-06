package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func getYN(preselected bool) bool {
	var response string
	fmt.Scanln(&response)
	return response == "Y" || response == "y" || (response == "" && preselected)
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

func cmd(id, command string) {
	fmt.Printf("Running (%s): %s\n", id, command)
	c := exec.Command("bash", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Start(); err != nil {
		fmt.Printf("cmd (%s) command failed: %s\n", id, err.Error())
		os.Exit(1)
	}
	if err := c.Wait(); err != nil {
		fmt.Printf("cmd (%s) wait failed: %s\n", id, err.Error())
		os.Exit(1)
	}
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
