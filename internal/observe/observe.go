package observe

import (
	"context"
	"time"
)

type Kind string

const (
	KindProcess Kind = "process"
	KindTCP     Kind = "tcp"
	KindDNS     Kind = "dns"
	KindFile    Kind = "file"
)

type Event struct {
	Time    time.Time
	Kind    Kind
	PID     int
	Process string
	Path    string
	Host    string
	Port    int
	Name    string
}

type Observer interface {
	Name() string
	Start(context.Context) (<-chan Event, <-chan error, error)
}
