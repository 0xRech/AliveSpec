//go:build !linux

package observe

import "fmt"

func NewBPFTrace(processes []string) (Observer, error) {
	return nil, fmt.Errorf("runtime recording is currently supported on Linux only")
}
