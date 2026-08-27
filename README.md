# AliveSpec

> **Learn how your system works while it works.**

AliveSpec turns a known-good runtime into an **executable operational contract**.

Traditional monitoring can tell you that a server is online while a real user journey is broken. AliveSpec asks a different question:

> **What conditions were true when this function worked — and are they still true now?**

## Status

**v0.2.0-alpha.1** — Linux runtime-learning prototype.

The new `record` command observes a successful journey through an eBPF-backed `bpftrace` adapter and compiles runtime evidence into readable YAML.

Currently observed:

- process execution / participating process names
- outgoing IPv4 TCP connections
- opened configuration-like files

Existing contract checks also support:

- systemd services
- local TCP listeners
- DNS resolution
- TLS trust and certificate lifetime
- file existence / SHA-256 fingerprints

## Quick start

Requirements for runtime recording:

- Linux
- root or equivalent BPF capabilities
- `bpftrace`

Build AliveSpec:

```bash
go build -o alivespec ./cmd/alivespec
```

Record a successful journey:

```bash
sudo ./alivespec record login \
  --comm nginx \
  --comm myapp
```

Or use a fixed window:

```bash
sudo ./alivespec record document-upload \
  --comm myapp \
  --duration 30s
```

Then verify what was learned:

```bash
./alivespec verify login.alivespec.yaml
```

Compare two contracts:

```bash
./alivespec diff before.yaml after.yaml
```

## Runtime learning

A recording looks like this:

```text
 ALIVESPEC  /  RUNTIME LEARN
────────────────────────────────────────────────────────────
  journey     login
  observer    eBPF / bpftrace
  window      until Ctrl+C
  processes   nginx, myapp
  contract    login.alivespec.yaml
────────────────────────────────────────────────────────────
● recording runtime evidence…

  20:13:04  PROC  myapp            /opt/myapp/app
  20:13:06  TCP   myapp            10.10.20.15:5432
  20:13:06  FILE  myapp            /etc/myapp/app.yaml
```

The generated contract keeps evidence provenance instead of pretending every observation is automatically a guaranteed dependency:

```yaml
requires:
  connections:
    - protocol: tcp
      host: 10.10.20.15
      port: 5432
      evidence:
        source: observed
        observations: 4
        confidence: 0.9
        processes:
          - myapp
```

See [`docs/RUNTIME_RECORDING.md`](docs/RUNTIME_RECORDING.md) for the current recording model and limitations.

## Manual learning

The original explicit-hint mode remains available:

```bash
sudo ./alivespec learn \
  --name login \
  --service nginx.service \
  --dns ldap.example.internal \
  --tls ldap.example.internal:636 \
  --file /etc/myapp/config.yaml \
  --out login.alivespec.yaml
```

## Design principles

1. **Observed beats assumed.** Runtime evidence is first-class.
2. **Evidence is not certainty.** Observed, declared and later confirmed dependencies stay distinguishable.
3. **No black box.** Generated contracts remain readable YAML.
4. **Local-first.** No cloud service is required.
5. **Metadata only.** No packet payloads, credentials, private keys or file contents should be captured.
6. **Useful before AI.** Core learning and verification remain deterministic.

## Backend architecture

`record` talks to an observer interface. The first Linux backend uses `bpftrace`, which keeps the Go build simple while the event and contract model stabilizes. A native `cilium/ebpf` backend is planned without changing the CLI or contract semantics.

## Roadmap

See [`docs/ROADMAP.md`](docs/ROADMAP.md).

Long-term AliveSpec should be able to answer:

```text
What changed since the last known-good journey?
Which journeys depend on this endpoint?
What breaks if this certificate, DNS record or service changes?
```

## Security

See [`SECURITY.md`](SECURITY.md).

## License

MIT
