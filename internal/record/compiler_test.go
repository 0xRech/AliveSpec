package record

import (
	"testing"
	"time"

	"github.com/0xRech/AliveSpec/internal/observe"
)

func TestCollectorDeduplicatesConnectionEvidence(t *testing.T) {
	c := NewCollector(false)
	event := observe.Event{Time: time.Now(), Kind: observe.KindTCP, PID: 10, Process: "myapp", Host: "10.0.0.5", Port: 5432}
	c.Add(event)
	c.Add(event)

	contract := c.Contract("login")
	if len(contract.Requires.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(contract.Requires.Connections))
	}
	got := contract.Requires.Connections[0]
	if got.Evidence.Observations != 2 {
		t.Fatalf("expected two observations, got %d", got.Evidence.Observations)
	}
	if len(got.Evidence.Processes) != 1 || got.Evidence.Processes[0] != "myapp" {
		t.Fatalf("unexpected processes: %#v", got.Evidence.Processes)
	}
}

func TestCollectorFiltersRuntimeFileNoise(t *testing.T) {
	c := NewCollector(false)
	c.Add(observe.Event{Kind: observe.KindFile, Process: "myapp", Path: "/usr/lib/libc.so"})
	c.Add(observe.Event{Kind: observe.KindFile, Process: "myapp", Path: "/etc/myapp/app.yaml"})

	contract := c.Contract("login")
	if len(contract.Requires.Files) != 1 {
		t.Fatalf("expected one interesting file, got %d", len(contract.Requires.Files))
	}
	if contract.Requires.Files[0].Path != "/etc/myapp/app.yaml" {
		t.Fatalf("unexpected file %q", contract.Requires.Files[0].Path)
	}
}

func TestInterestingFile(t *testing.T) {
	cases := map[string]bool{
		"/etc/nginx/nginx.conf":       true,
		"/opt/app/config/app.yaml":    true,
		"/etc/resolv.conf":            true,
		"/proc/123/status":            false,
		"/usr/lib/x86_64-linux/a.so":  false,
		"relative/config.yaml":        false,
	}
	for path, want := range cases {
		if got := interestingFile(path); got != want {
			t.Errorf("interestingFile(%q) = %t, want %t", path, got, want)
		}
	}
}
