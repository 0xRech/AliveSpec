package verify

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	runtimeinfo "github.com/0xRech/AliveSpec/internal/runtime"
	"github.com/0xRech/AliveSpec/internal/spec"
	"github.com/0xRech/AliveSpec/internal/ui"
)

type result struct {
	ok     bool
	kind   string
	target string
	detail string
}

type verificationFailed int

func (e verificationFailed) Error() string { return fmt.Sprintf("%d contract check(s) failed", int(e)) }
func (e verificationFailed) Quiet() bool   { return true }

func Run(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	timeout := fs.Duration("timeout", 5*time.Second, "network timeout")
	color := fs.String("color", "auto", "terminal color: auto, always, never")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: alivespec verify <contract.yaml>")
	}
	if *color != "auto" && *color != "always" && *color != "never" {
		return fmt.Errorf("invalid --color %q; use auto, always or never", *color)
	}

	contractPath := fs.Arg(0)
	c, err := spec.Load(contractPath)
	if err != nil {
		return err
	}

	printer := ui.New(os.Stdout, *color)
	printer.VerifyHeader(c.Metadata.Name, contractPath, c.Metadata.Host)
	started := time.Now()

	var results []result
	for _, req := range c.Requires.Processes {
		ok := runtimeinfo.IsProcessRunning(req.Name)
		detail := "running"
		if !ok {
			detail = "not running"
		}
		results = append(results, result{ok: ok, kind: "process", target: req.Name, detail: detail})
	}
	for _, req := range c.Requires.Services {
		ok := runtimeinfo.IsServiceActive(req.Name)
		detail := "active"
		if !ok {
			detail = "inactive or unavailable"
		}
		results = append(results, result{ok: ok, kind: "service", target: req.Name, detail: detail})
	}
	for _, req := range c.Requires.Listeners {
		ok := req.Protocol == "tcp" && runtimeinfo.IsTCPListening(req.Port)
		detail := "listening"
		if !ok {
			detail = "not listening"
		}
		results = append(results, result{
			ok:     ok,
			kind:   "listener",
			target: fmt.Sprintf("%s/%d", strings.ToUpper(req.Protocol), req.Port),
			detail: detail,
		})
	}
	for _, req := range c.Requires.Connections {
		results = append(results, verifyConnection(req, *timeout))
	}
	for _, req := range c.Requires.DNS {
		_, lookupErr := net.LookupHost(req.Name)
		detail := "resolved successfully"
		if lookupErr != nil {
			detail = "resolution failed: " + lookupErr.Error()
		}
		results = append(results, result{ok: lookupErr == nil, kind: "dns", target: req.Name, detail: detail})
	}
	for _, req := range c.Requires.TLS {
		results = append(results, verifyTLS(req, *timeout))
	}
	for _, req := range c.Requires.Files {
		results = append(results, verifyFile(req))
	}

	failures := 0
	for _, r := range results {
		printer.Check(r.ok, r.kind, r.target, r.detail)
		if !r.ok {
			failures++
		}
	}
	printer.VerifySummary(len(results)-failures, len(results), time.Since(started))

	if failures > 0 {
		return verificationFailed(failures)
	}
	return nil
}

func verifyConnection(req spec.ConnectionRequirement, timeout time.Duration) result {
	address := net.JoinHostPort(req.Host, fmt.Sprint(req.Port))
	if req.Protocol != "tcp" {
		return result{ok: false, kind: "connection", target: address, detail: "unsupported protocol: " + req.Protocol}
	}
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return result{ok: false, kind: "connection", target: address, detail: "unreachable: " + err.Error()}
	}
	_ = conn.Close()
	return result{ok: true, kind: "connection", target: address, detail: "TCP reachable"}
}

func verifyTLS(req spec.TLSRequirement, timeout time.Duration) result {
	serverName := req.ServerName
	if serverName == "" {
		serverName = req.Host
	}
	address := net.JoinHostPort(req.Host, fmt.Sprint(req.Port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12})
	if err != nil {
		return result{ok: false, kind: "tls", target: address, detail: "handshake failed: " + err.Error()}
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return result{ok: false, kind: "tls", target: address, detail: "peer returned no certificate"}
	}
	remaining := time.Until(state.PeerCertificates[0].NotAfter)
	min := time.Duration(req.MinValidityDays) * 24 * time.Hour
	if remaining < min {
		return result{
			ok:     false,
			kind:   "tls",
			target: address,
			detail: fmt.Sprintf("trusted, but only %.1f days remain (minimum %d)", remaining.Hours()/24, req.MinValidityDays),
		}
	}
	return result{
		ok:     true,
		kind:   "tls",
		target: address,
		detail: fmt.Sprintf("trusted · certificate valid for %.1f more days", remaining.Hours()/24),
	}
}

func verifyFile(req spec.FileRequirement) result {
	if _, err := os.Stat(req.Path); err != nil {
		return result{ok: false, kind: "file", target: req.Path, detail: "missing or inaccessible"}
	}
	if req.SHA256 == "" {
		return result{ok: true, kind: "file", target: req.Path, detail: "exists"}
	}
	hash, err := runtimeinfo.FileSHA256(req.Path)
	if err != nil {
		return result{ok: false, kind: "file", target: req.Path, detail: "hash failed: " + err.Error()}
	}
	if hash != req.SHA256 {
		return result{ok: false, kind: "file", target: req.Path, detail: "SHA-256 changed"}
	}
	return result{ok: true, kind: "file", target: req.Path, detail: "exists · SHA-256 matches"}
}
