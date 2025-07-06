package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path"
	"runtime"
	"time"
)

type Command int

const (
	CommandInvalid Command = iota
	CommandRunDebug
	CommandRun
	CommandCompile
	CommandApp
	CommandGenCA
)

var command Command

const (
	ViteAddress       = "localhost:5173"
	CheckViteTimeout  = 1 * time.Second
	CheckViteInterval = 500 * time.Millisecond
	CheckViteRetries  = 5
)

var args []string

func init() {
	flag.Parse()

	cmd := flag.Arg(0)
	args = flag.Args()
	if len(args) > 1 {
		args = args[1:] // Remove the command from args
	} else {
		args = []string{} // No additional args
	}

	switch cmd {
	case "debug":
		command = CommandRunDebug
	case "run":
		command = CommandRun
	case "app":
		command = CommandApp
	case "gen-ca":
		command = CommandGenCA
	case "setup":
		command = CommandCompile
	default:
		slog.Error("Invalid command. Use debug, run, setup, app, or gen-ca.")
	}
}

func main() {
	switch command {
	case CommandRunDebug:
		runDebug()
	case CommandRun:
		run()
	case CommandApp:
		app()
	case CommandGenCA:
		genCA()
	case CommandCompile:
		compile()
	default:
		fmt.Println("Invalid command. Use debug, run, app, setup, or gen-ca.")
	}
}

// CA:
func genCA() {
	var CERTS_DIR = "./certs"
	if len(args) > 0 {
		CERTS_DIR = args[0]
	}
	if err := os.MkdirAll(CERTS_DIR, 0744); err != nil {
		slog.Error("error creating certs directory", "error", err)
		return
	}
	cmd("01", fmt.Sprintf("openssl req -x509 -newkey rsa:4096 -keyout %s/ca.key -out %s/ca.crt -days 3650 -nodes -subj \"/CN=CAP\"", CERTS_DIR, CERTS_DIR))
	fmt.Println("CA certificate and key generated in", CERTS_DIR)
	fmt.Println("NOTE: this certificate expires in 3650 day (~10 years). You can regenerate it at any time by running this command again.")
	fmt.Printf("NOTE: you NEED to trust this CA certificate in your system/browser for the proxy to work properly. prompt for trust (Y/n)? ")
	if getYN(true) {
		trustCA(path.Join(CERTS_DIR, "ca.crt"))
	} else {
		fmt.Println("Skipping CA trust prompt.")
	}
}

func trustCA(certPath string) {
	switch runtime.GOOS {
	case "windows":
		fmt.Printf("This command requires admin privileges to run. Continue? (Y/n) ")
		if !getYN(true) {
			fmt.Println("Skipping CA trust on Windows.")
			return
		}
		cmd("02", fmt.Sprintf("certutil -addstore -f Root %s", certPath))
	case "darwin":
		fmt.Printf("This command will prompt you for the admin password used for sudo. Continue? (Y/n) ")
		if !getYN(true) {
			fmt.Println("Skipping CA trust on macOS.")
			return
		}
		cmd("02", fmt.Sprintf("sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s", certPath))
	case "linux":
		fmt.Printf("This command will prompt you for the admin password used for sudo. Continue? (Y/n) ")
		if !getYN(true) {
			fmt.Println("Skipping CA trust on macOS.")
			return
		}
		cmd("02", fmt.Sprintf("sudo cp %s /usr/local/share/ca-certificates/my-ca.crt", certPath))
		cmd("03", "sudo update-ca-certificates")
	default:
		// NOTE: we shouldn't even prompt for trust on unsupported OSes.
		fmt.Println("Unsupported OS for CA trust:", runtime.GOOS)
		fmt.Println("You need to trust the CA certificate manually in your system/browser for the proxy to work properly.")
		fmt.Println("Please refer to the documentation for your OS or browser on how to trust a CA certificate.")
		fmt.Println("The CA certificate is located at:", certPath)
		fmt.Println("Skipping CA trust prompt.")
	}
}

// build app for macos
func app() {
	if runtime.GOOS != "darwin" {
		log.Fatalf("FATAL: GOOS %s NOT SUPPORTED", runtime.GOOS)
	}
	cmd("01", "npm run build --prefix ./manager -- --outDir vitedist")
	cmd("02", "go build -o ./proxy/proxy-app ./proxy")
	cmd("03", "cd manager && npx electron-builder build")
	cmd("04", "cp ./proxy/proxy-app ./manager/dist/mac-arm64/cap.app/Contents/Resources")
	cmd("05", "cp -R ./manager/vitedist ./manager/dist/mac-arm64/cap.app/Contents/Resources/dist")
	var certsDir = os.Getenv("CERTS_DIR")
	if certsDir == "" {
		fmt.Println("CERTS_DIR environment variable not set, using default: certs")
		certsDir = "certs"
	}
	_, err := os.Stat(certsDir)
	if err != nil {
		fmt.Printf("certs directory not found, generate CA? (Y/n): ")
		var response string
		fmt.Scanln(&response)
		if response == "" || response == "Y" || response == "y" {
			args = []string{certsDir} // set args to the certs directory for genCA
			genCA()
		} else {
			fmt.Println("Skipping CA generation.")
		}
	} else {
		fmt.Println("Using existing certs directory.")
		cmd("06", fmt.Sprintf("cp -R ./%s ./manager/dist/mac-arm64/cap.app/Contents/Resources/certs", certsDir))
	}
	cmd("07", "cp -R ./manager/dist/mac-arm64/cap.app .")
}

func compile() {
	cmd("01", "go mod tidy")
	cmd("02", "npm i --prefix manager")
	cmd("03", "go build -o ./proxy/proxy-app ./proxy")
	cmd("04", "npm run build --prefix ./manager --outDir vitedist")
	fmt.Printf("Generate certificates required for MITM functionality? (Y/n): ")
	if getYN(true) {
		genCA()
	}
}

func run() {
	fmt.Println("Make sure you've run the compile command before running this command.")

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
