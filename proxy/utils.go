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
	// NOTE: make matching ip configurable; add switcg case
	if ip != "::1" && ip != "localhost" && ip != "127.0.0.1" && ip != myLocalIP {
		// fmt.Println("req not from localhost", ip)
		return
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		*pid, *name = getMacLinuxProcessInfo(port)
	}
}

func getMacLinuxProcessInfo(port string) (pid int, pname string) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("lsof -i :%s -F c", port))
	out, err := cmd.Output()
	if err != nil {
		// fmt.Println("lsof error:", err)
		return 0, ""
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) < 4 || (len(lines)-1)%3 != 0 {
		// fmt.Println("lsof output too short:", len(lines))
		return 0, ""
	}

	for i, line := range lines {
		if line[0] != 'c' {
			continue
		}
		pid, err := strconv.Atoi(lines[i-1][1:]) // previous line 'p[PID]' - remove 'p' prefix
		if err == nil && pid != myPID {
			return pid, line[1:] // remove 'c' prefix
		}
	}
	// fmt.Println("no matching process found")
	return 0, ""
}
