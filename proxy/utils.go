package main

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// toURL converts a string to a URL. If the string does not start with "http://" or
// "https://", it will prepend "http://" or "https://" based on the https parameter.
// It returns an error if the string is not a valid URL after conversion.
//
// This function is used to ensure that the URL is in a valid format before
// performing any net/http Do operations.
func toURL(s string, https bool) (*url.URL, error) {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		if https {
			s = "https://" + s
		} else {
			s = "http://" + s
		}
	}
	return url.Parse(s)
}

func getClientProcessInfo(clientIP string, pid *int, name *string) {
	ip, port, err := net.SplitHostPort(clientIP)
	if err != nil {
		// fmt.Println("net split")
		return
	}
	// NOTE: make matching ip configurable
	if ip != "::1" && ip != "localhost" && ip != "127.0.0.1" && ip != myLocalIP {
		// fmt.Println("req not from localhost", ip)
		return
	}
	switch runtime.GOOS {
	case "darwin":
		*pid, *name = getMacLinuxProcessInfo(port)
	}
}

func getMacLinuxProcessInfo(port string) (pid int, pname string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("lsof -i :%s", port))
	out, err := cmd.Output()
	if err != nil {
		// fmt.Println("lsof error:", err)
		return 0, ""
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 3 {
		// fmt.Println("lsof output too short:", len(lines))
		return 0, ""
	}

	// fmt.Println("output\n", strings.Join(lines, "\n"))
	pid, pname = getProcessFromLsofOutputLine(lines[1], myPID)
	if pid != 0 {
		return pid, pname
	}
	return getProcessFromLsofOutputLine(lines[2], myPID)
}

func getProcessFromLsofOutputLine(out string, proxyPID int) (pid int, pname string) {
	parts := strings.Fields(out)
	if len(parts) < 2 {
		return 0, ""
	}

	var err error
	pid, err = strconv.Atoi(parts[1])
	// fmt.Println("76", parts[0], pid, proxyPID)
	if err == nil && pid != proxyPID {
		return pid, parts[0]
	}

	return 0, ""
}
