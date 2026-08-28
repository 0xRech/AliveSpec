# AliveSpec

> **Learn how your system works while it works.**

AliveSpec turns a known-good runtime into an **executable operational contract**.

Traditional monitoring can tell you that a server is online while a real user journey is broken. AliveSpec asks a different question:

> **What conditions were true when this function worked — and are they still true now?**

> [!WARNING]
> AliveSpec is currently **experimental alpha software**. The runtime observer is Linux-only, requires elevated BPF permissions, and currently uses `bpftrace` as its first eBPF backend.

## Status

**v0.2.0-alpha.1** — Linux runtime-learning prototype.

The `record` command observes a successful journey and compiles runtime evidence into readable YAML.

Currently observed:

- process execution / participating process names
- outgoing IPv4 and IPv6 TCP connections
- best-effort DNS lookups via glibc `getaddrinfo()`
- opened configuration-like files
- evidence counts and confidence metadata

Existing contract checks also support:

- systemd services
- local TCP listeners
- learned TCP dependencies
- DNS resolution
- TLS trust and certificate lifetime
- process presence
- file existence / SHA-256 fingerprints

## Requirements

Runtime recording currently requires:

- Linux
- Go 1.23+ to build from source
- `bpftrace`
- root or equivalent BPF capabilities

Example package installation on Debian/Ubuntu:

```bash
sudo apt update
sudo apt install bpftrace
```

## Quick start

Build AliveSpec:

```bash
git clone https://github.com/0xRech/AliveSpec.git
cd AliveSpec
go build -o alivespec ./cmd/alivespec
```

Record a successful journey:

```bash
sudo ./alivespec record login \
  --comm nginx \
  --comm myapp
```

Or use a fixed recording window:

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

A recording uses AliveSpec's compact **Modern Ops** terminal UI:

```text
╭─ AliveSpec · Runtime Learning ──────────────────────────────────────╮
│  ● RECORDING                                                       │
│                                                                    │
│  Journey       login                                               │
│  Window        until Ctrl+C                                        │
│  Observer      eBPF / bpftrace · DNS enabled                       │
│  Processes     nginx, myapp                                        │
│  Contract      login.alivespec.yaml                                │
╰────────────────────────────────────────────────────────────────────╯

  DISCOVERED

  20:23:04  ◈ PROCESS     myapp
            └─ /opt/myapp/app

  20:23:05  ⌁ DNS         myapp  confidence 65%
            └─ db01.internal

  20:23:05  ↗ CONNECTION  myapp
            └─ 10.10.20.15:5432

  20:23:06  ↗ CONNECTION  myapp
            └─ [2001:db8::20]:443

  20:23:07  ◫ FILE        myapp  confidence 70%
            └─ /etc/myapp/app.yaml
```

Repeated events do not flood the terminal. They increase the evidence count in the generated contract instead:

```yaml
requires:
  connections:
    - protocol: tcp
      host: 10.10.20.15
      port: 5432
      evidence:
        source: observed
        observations: 501
        confidence: 0.9
        processes:
          - myapp
```

DNS learning is intentionally **best effort**. The current backend enables a glibc `getaddrinfo()` uprobe only when the probe is available. Static resolvers and non-glibc runtimes may therefore not expose DNS names yet.

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

## Verify

AliveSpec can re-check learned or declared conditions later:

```text
╭─ AliveSpec · Contract Verification ────────────────────────────────╮
│  ● VERIFYING                                                       │
│                                                                    │
│  Journey       login                                               │
│  Contract      login.alivespec.yaml                                │
╰────────────────────────────────────────────────────────────────────╯

  CHECKS

  ✓  PROCESS     myapp
                 └─ running

  ✓  CONNECTION  10.10.20.15:5432
                 └─ TCP reachable

  ✕  FILE        /etc/myapp/app.yaml
                 └─ SHA-256 changed
```

A degraded verification returns a non-zero exit code, making AliveSpec usable in CI/CD pipelines.

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

## Contributing

Issues and pull requests are welcome while the project is in alpha. For runtime-observer changes, please document what metadata is captured and what permissions are required.

See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

MIT
