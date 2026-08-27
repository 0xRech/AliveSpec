package verify

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	runtimeinfo "github.com/0xRech/AliveSpec/internal/runtime"
	"github.com/0xRech/AliveSpec/internal/spec"
)

type result struct {
	ok      bool
	message string
}

func Run(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 5*time.Second, "network timeout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: alivespec verify <contract.yaml>")
	}

	c, err := spec.Load(fs.Arg(0))
	if err != nil {
		return err
	}
	fmt.Printf("AliveSpec: %s\n\n", c.Metadata.Name)

	var results []result
	for _, req := range c.Requires.Services {
		results = append(results, result{runtimeinfo.IsServiceActive(req.Name), fmt.Sprintf("service %s active", req.Name)})
	}
	for _, req := range c.Requires.Listeners {
		ok := req.Protocol == "tcp" && runtimeinfo.IsTCPListening(req.Port)
		results = append(results, result{ok, fmt.Sprintf("%s/%d listening", req.Protocol, req.Port)})
	}
	for _, req := range c.Requires.DNS {
		_, err := net.LookupHost(req.Name)
		results = append(results, result{err == nil, fmt.Sprintf("DNS %s resolves", req.Name)})
	}
	for _, req := range c.Requires.TLS {
		ok, detail := verifyTLS(req, *timeout)
		results = append(results, result{ok, detail})
	}
	for _, req := range c.Requires.Files {
		ok, detail := verifyFile(req)
		results = append(results, result{ok, detail})
	}

	failures := 0
	for _, r := range results {
		if r.ok {
			fmt.Printf("✓ %s\n", r.message)
		} else {
			fmt.Printf("✗ %s\n", r.message)
			failures++
		}
	}
	fmt.Printf("\nResult: %d/%d checks passed\n", len(results)-failures, len(results))
	if failures > 0 {
		return fmt.Errorf("%d contract check(s) failed", failures)
	}
	return nil
}

func verifyTLS(req spec.TLSRequirement, timeout time.Duration) (bool, string) {
	serverName := req.ServerName
	if serverName == "" {
		serverName = req.Host
	}
	address := net.JoinHostPort(req.Host, fmt.Sprint(req.Port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12})
	if err != nil {
		return false, fmt.Sprintf("TLS %s failed: %v", address, err)
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return false, fmt.Sprintf("TLS %s returned no certificate", address)
	}
	remaining := time.Until(state.PeerCertificates[0].NotAfter)
	min := time.Duration(req.MinValidityDays) * 24 * time.Hour
	if remaining < min {
		return false, fmt.Sprintf("TLS %s certificate validity %.1f days < required %d days", address, remaining.Hours()/24, req.MinValidityDays)
	}
	return true, fmt.Sprintf("TLS %s trusted; certificate valid for %.1f more days", address, remaining.Hours()/24)
}

func verifyFile(req spec.FileRequirement) (bool, string) {
	if _, err := os.Stat(req.Path); err != nil {
		return false, fmt.Sprintf("file %s exists", req.Path)
	}
	if req.SHA256 == "" {
		return true, fmt.Sprintf("file %s exists", req.Path)
	}
	hash, err := runtimeinfo.FileSHA256(req.Path)
	if err != nil {
		return false, fmt.Sprintf("file %s hash failed: %v", req.Path, err)
	}
	if hash != req.SHA256 {
		return false, fmt.Sprintf("file %s SHA-256 changed", req.Path)
	}
	return true, fmt.Sprintf("file %s exists and SHA-256 matches", req.Path)
}
