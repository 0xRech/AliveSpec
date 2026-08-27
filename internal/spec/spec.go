package spec

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	APIVersion = "alivespec.dev/v1alpha1"
	Kind       = "OperationalContract"
)

type Contract struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Requires   Requirements `yaml:"requires"`
}

type Metadata struct {
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"createdAt,omitempty"`
	Host      string    `yaml:"host,omitempty"`
}

type Requirements struct {
	Processes   []ProcessRequirement    `yaml:"processes,omitempty"`
	Services    []ServiceRequirement    `yaml:"services,omitempty"`
	Listeners   []ListenerRequirement   `yaml:"listeners,omitempty"`
	Connections []ConnectionRequirement `yaml:"connections,omitempty"`
	DNS         []DNSRequirement        `yaml:"dns,omitempty"`
	TLS         []TLSRequirement        `yaml:"tls,omitempty"`
	Files       []FileRequirement       `yaml:"files,omitempty"`
}

type Evidence struct {
	Source       string   `yaml:"source,omitempty"`
	Observations int      `yaml:"observations,omitempty"`
	Confidence   float64  `yaml:"confidence,omitempty"`
	Processes    []string `yaml:"processes,omitempty"`
}

type ProcessRequirement struct {
	Name       string   `yaml:"name"`
	Executable string   `yaml:"executable,omitempty"`
	Running    bool     `yaml:"running"`
	Evidence   Evidence `yaml:"evidence,omitempty"`
}

type ServiceRequirement struct {
	Name   string `yaml:"name"`
	Active bool   `yaml:"active"`
}

type ListenerRequirement struct {
	Protocol string `yaml:"protocol"`
	Port     int    `yaml:"port"`
}

type ConnectionRequirement struct {
	Protocol string   `yaml:"protocol"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	Evidence Evidence `yaml:"evidence,omitempty"`
}

type DNSRequirement struct {
	Name     string `yaml:"name"`
	Resolves bool   `yaml:"resolves"`
}

type TLSRequirement struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	ServerName      string `yaml:"serverName,omitempty"`
	MinValidityDays int    `yaml:"minValidityDays,omitempty"`
}

type FileRequirement struct {
	Path     string   `yaml:"path"`
	Exists   bool     `yaml:"exists"`
	SHA256   string   `yaml:"sha256,omitempty"`
	Evidence Evidence `yaml:"evidence,omitempty"`
}

func New(name, host string) *Contract {
	return &Contract{APIVersion: APIVersion, Kind: Kind, Metadata: Metadata{Name: name, CreatedAt: time.Now().UTC(), Host: host}}
}

func Load(path string) (*Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Contract
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func Save(path string, c *Contract) error {
	if err := c.Validate(); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c Contract) Validate() error {
	if c.APIVersion != APIVersion {
		return fmt.Errorf("unsupported apiVersion %q", c.APIVersion)
	}
	if c.Kind != Kind {
		return fmt.Errorf("unsupported kind %q", c.Kind)
	}
	if c.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	for _, r := range c.Requires.Connections {
		if r.Protocol == "" || r.Host == "" || r.Port < 1 || r.Port > 65535 {
			return fmt.Errorf("invalid connection requirement %s://%s:%d", r.Protocol, r.Host, r.Port)
		}
	}
	return nil
}
