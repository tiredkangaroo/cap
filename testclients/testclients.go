package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

var names = []string{
	"dog_client",
	"cat_client",
	"fish_client",
	"bird_client",
	"hamster_client",
	"turtle_client",
	"snake_client",
	"rabbit_client",
	"lizard_client",
	"ferret_client",
	"guinea_pig_client",
	"chinchilla_client",
	"hedgehog_client",
	"gerbil_client",
	"mouse_client",
	"rat_client",
	"parrot_client",
	"canary_client",
}

var X_TIMES = os.Getenv("X_TIMES")

func main() {
	wg := &sync.WaitGroup{}
	for _, name := range names {
		wg.Add(1)
		go runClient(wg, name)
	}
	wg.Wait()
}

func runClient(wg *sync.WaitGroup, name string) {
	defer wg.Done()
	err := exec.Command("go", "build", "-o", name, "client/client.go").Run()
	if err != nil {
		slog.Error("error building client %s: %v", name, err)
		return
	}

	defer os.Remove(name)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("X_TIMES=%s ./%s", X_TIMES, name))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		slog.Error("error running client %s: %v", name, err)
		return
	}
}
