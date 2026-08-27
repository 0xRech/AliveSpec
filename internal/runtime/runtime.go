package runtime

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

func IsServiceActive(name string) bool {
	return exec.Command("systemctl", "is-active", "--quiet", name).Run() == nil
}

func TCPListeners() ([]int, error) {
	ports := map[int]struct{}{}
	var firstErr error

	for _, path := range []string{"/proc/net/tcp", "/proc/net/tcp6"} {
		file, err := os.Open(path)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		scanner := bufio.NewScanner(file)
		first := true
		for scanner.Scan() {
			if first {
				first = false
				continue
			}
			fields := strings.Fields(scanner.Text())
			if len(fields) < 4 || fields[3] != "0A" {
				continue
			}
			parts := strings.Split(fields[1], ":")
			if len(parts) != 2 {
				continue
			}
			p, err := strconv.ParseInt(parts[1], 16, 32)
			if err == nil {
				ports[int(p)] = struct{}{}
			}
		}
		if err := scanner.Err(); err != nil && firstErr == nil {
			firstErr = err
		}
		_ = file.Close()
	}

	if len(ports) == 0 && firstErr != nil {
		return nil, firstErr
	}
	result := make([]int, 0, len(ports))
	for p := range ports {
		result = append(result, p)
	}
	sort.Ints(result)
	return result, nil
}

func IsTCPListening(port int) bool {
	ports, err := TCPListeners()
	if err != nil {
		return false
	}
	for _, p := range ports {
		if p == port {
			return true
		}
	}
	return false
}

func FileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func ParseHostPort(value string, defaultPort int) (string, int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", 0, fmt.Errorf("empty endpoint")
	}
	if !strings.Contains(value, ":") {
		return value, defaultPort, nil
	}
	idx := strings.LastIndex(value, ":")
	host := value[:idx]
	port, err := strconv.Atoi(value[idx+1:])
	if err != nil || host == "" || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("invalid endpoint %q", value)
	}
	return host, port, nil
}
