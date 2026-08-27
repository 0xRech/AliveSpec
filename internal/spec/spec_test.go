package spec

import (
	"path/filepath"
	"testing"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "contract.yaml")
	c := New("login", "test-host")
	c.Requires.Listeners = append(c.Requires.Listeners, ListenerRequirement{Protocol: "tcp", Port: 443})

	if err := Save(path, c); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "login" {
		t.Fatalf("unexpected name: %s", got.Metadata.Name)
	}
	if len(got.Requires.Listeners) != 1 || got.Requires.Listeners[0].Port != 443 {
		t.Fatalf("listener round-trip failed")
	}
}

func TestValidateRejectsWrongVersion(t *testing.T) {
	c := New("x", "host")
	c.APIVersion = "wrong/v1"
	if err := c.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}
