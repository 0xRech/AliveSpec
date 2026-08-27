package observe

import (
	"context"
	"time"
)

type Kind string

const (
	KindProcess Kind = "process"
	KindTCP     Kind = "tcp"
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
}

type Observer interface {
	Name() string
	Start(context.Context) (<-chan Event, <-chan error, error)
}
