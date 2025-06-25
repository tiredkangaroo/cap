package main

import (
	"encoding/json"
	"fmt"
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
func toURL(s string, secure bool) (*url.URL, error) {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		if secure {
			s = "https://" + s
		} else {
			s = "http://" + s
		}
	}
	return url.Parse(s)
}

func getClientProcessInfo(clientIP, clientPort string, pid *int, name *string) {
	if clientIP != ThisDevice {
		*pid = 0
		*name = ""
		return
	}
	switch runtime.GOOS {
	case "darwin", "linux":
		*pid, *name = getMacLinuxProcessInfo(clientPort)
	}
}

func ipIsLocalhost(ip string) bool {
	// NOTE: make matching ip configurable
	switch ip {
	case "::1", "localhost", "127.0.0.1", myLocalIP:
		return true
	}
	return false
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
		if len(line) < 2 {
			continue
		}
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

func marshal(data any) []byte {
	m, _ := json.Marshal(data)
	return m
}
