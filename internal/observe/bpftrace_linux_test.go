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

func TestParseTCP4Event(t *testing.T) {
	e, err := parseLine("AS|TCP4|321|myapp|10.0.0.5|5432")
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != KindTCP || e.Host != "10.0.0.5" || e.Port != 5432 {
		t.Fatalf("unexpected event: %#v", e)
	}
}

func TestParseTCP6Event(t *testing.T) {
	e, err := parseLine("AS|TCP6|321|myapp|2001:db8::10|8443")
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != KindTCP || e.Host != "2001:db8::10" || e.Port != 8443 {
		t.Fatalf("unexpected event: %#v", e)
	}
}

func TestParseDNSEvent(t *testing.T) {
	e, err := parseLine("AS|DNS|42|myapp|db01.internal")
	if err != nil {
		t.Fatal(err)
	}
	if e.Kind != KindDNS || e.Name != "db01.internal" || e.Process != "myapp" {
		t.Fatalf("unexpected event: %#v", e)
	}
}

func TestFilterRejectsUnsafeComm(t *testing.T) {
	if _, err := NewBPFTrace([]string{`bad/name`}); err == nil {
		t.Fatal("expected unsafe comm to be rejected")
	}
}
