//go:build linux

package observe

import "testing"

func TestParseProcessEvent(t *testing.T) {
	e, err := parseLine("AS|EXEC|123|myapp|/opt/myapp/app")
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != KindProcess || e.PID != 123 || e.Process != "myapp" || e.Path != "/opt/myapp/app" {
		t.Fatalf("unexpected event: %#v", e)
	}
}

func TestParseTCPEvent(t *testing.T) {
	e, err := parseLine("AS|TCP4|321|myapp|10.0.0.5|5432")
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != KindTCP || e.Host != "10.0.0.5" || e.Port != 5432 {
		t.Fatalf("unexpected event: %#v", e)
	}
}

func TestFilterRejectsUnsafeComm(t *testing.T) {
	if _, err := NewBPFTrace([]string{`bad/name`}); err == nil {
		t.Fatal("expected unsafe comm to be rejected")
	}
}
