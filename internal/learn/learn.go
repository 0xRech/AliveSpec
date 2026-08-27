package learn

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sort"

	runtimeinfo "github.com/0xRech/AliveSpec/internal/runtime"
	"github.com/0xRech/AliveSpec/internal/spec"
)

type multiFlag []string

func (m *multiFlag) String() string { return fmt.Sprint([]string(*m)) }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func Run(args []string) error {
	fs := flag.NewFlagSet("learn", flag.ContinueOnError)
	name := fs.String("name", "baseline", "contract name")
	out := fs.String("out", "alivespec.yaml", "output YAML file")
	allListeners := fs.Bool("all-listeners", true, "capture local TCP listeners")
	minTLSValidity := fs.Int("min-tls-validity", 14, "minimum TLS certificate lifetime in days")

	var services, dnsNames, tlsEndpoints, filePaths multiFlag
	fs.Var(&services, "service", "systemd service to require (repeatable)")
	fs.Var(&dnsNames, "dns", "DNS name that must resolve (repeatable)")
	fs.Var(&tlsEndpoints, "tls", "TLS endpoint host[:port] (repeatable)")
	fs.Var(&filePaths, "file", "file to require and fingerprint (repeatable)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	host, _ := os.Hostname()
	c := spec.New(*name, host)

	for _, service := range unique(services) {
		if runtimeinfo.IsServiceActive(service) {
			c.Requires.Services = append(c.Requires.Services, spec.ServiceRequirement{Name: service, Active: true})
		} else {
			fmt.Fprintf(os.Stderr, "warning: service %s is not active; skipped\n", service)
		}
	}

	if *allListeners {
		ports, err := runtimeinfo.TCPListeners()
		if err != nil {
			return fmt.Errorf("discover TCP listeners: %w", err)
		}
		for _, port := range ports {
			c.Requires.Listeners = append(c.Requires.Listeners, spec.ListenerRequirement{Protocol: "tcp", Port: port})
		}
	}

	for _, name := range unique(dnsNames) {
		if _, err := net.LookupHost(name); err != nil {
			fmt.Fprintf(os.Stderr, "warning: DNS %s does not resolve; skipped\n", name)
			continue
		}
		c.Requires.DNS = append(c.Requires.DNS, spec.DNSRequirement{Name: name, Resolves: true})
	}

	for _, endpoint := range unique(tlsEndpoints) {
		host, port, err := runtimeinfo.ParseHostPort(endpoint, 443)
		if err != nil {
			return err
		}
		c.Requires.TLS = append(c.Requires.TLS, spec.TLSRequirement{Host: host, Port: port, ServerName: host, MinValidityDays: *minTLSValidity})
	}

	for _, path := range unique(filePaths) {
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: file %s does not exist; skipped\n", path)
			continue
		}
		hash, err := runtimeinfo.FileSHA256(path)
		if err != nil {
			return fmt.Errorf("hash %s: %w", path, err)
		}
		c.Requires.Files = append(c.Requires.Files, spec.FileRequirement{Path: path, Exists: true, SHA256: hash})
	}

	if err := spec.Save(*out, c); err != nil {
		return err
	}
	fmt.Printf("Learned operational contract %q -> %s\n", c.Metadata.Name, *out)
	return nil
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
