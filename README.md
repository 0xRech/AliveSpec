# AliveSpec

> **Learn how your system works while it works.**

AliveSpec is an experimental **runtime-to-spec compiler** for operations.

Instead of manually writing health checks, AliveSpec captures a known-good runtime state and turns it into an **executable operational contract**. Later, `alivespec verify` checks whether the conditions that made the system work are still true.

## Why?

Traditional monitoring can tell you that a server is up while the actual user journey is broken.

AliveSpec asks a different question:

> **What conditions were true when this function worked — and are they still true now?**

A contract can describe:

- required systemd services
- local TCP listeners
- DNS resolution
- TLS trust and certificate lifetime
- required files and optional SHA-256 fingerprints

Future versions will learn runtime dependencies automatically using eBPF and correlate them with named user journeys.

## Status

**v0.1 prototype / foundation.**

The current `learn` command captures a Linux known-good baseline with explicit hints. The planned v0.2 observer will replace those hints with runtime observation.

## Quick start

```bash
go build -o alivespec ./cmd/alivespec

sudo ./alivespec learn \
  --name login \
  --service nginx.service \
  --dns ldap.example.internal \
  --tls ldap.example.internal:636 \
  --file /etc/myapp/config.yaml \
  --out login.alivespec.yaml

./alivespec verify login.alivespec.yaml
```

Compare two contracts:

```bash
./alivespec diff before.yaml after.yaml
```

## Contract example

```yaml
apiVersion: alivespec.dev/v1alpha1
kind: OperationalContract
metadata:
  name: login
requires:
  services:
    - name: nginx.service
      active: true
  listeners:
    - protocol: tcp
      port: 443
  dns:
    - name: ldap.example.internal
      resolves: true
  tls:
    - host: ldap.example.internal
      port: 636
      minValidityDays: 14
  files:
    - path: /etc/myapp/config.yaml
      exists: true
```

## The end goal

```text
alivespec learn "login"
        │
        ▼
Observe a successful runtime journey
        │
        ▼
Processes · DNS · TCP · TLS · files · services
        │
        ▼
Executable Operational Contract
        │
        ├── alivespec verify
        ├── alivespec diff
        └── alivespec whatif   (planned)
```

The long-term goal is to answer questions like:

```text
What breaks if this certificate changes?
Which successful journeys depend on db01:5432?
What changed between the last known-good run and now?
```

## Design principles

1. **Observed beats assumed.** Runtime evidence should be first-class.
2. **No black box.** Generated contracts stay readable YAML.
3. **Confidence matters.** Future dependencies will be tagged as observed, confirmed, inferred, or ignored.
4. **Local-first.** No cloud service is required.
5. **Useful before AI.** Core verification remains deterministic.

## Roadmap

See [`docs/ROADMAP.md`](docs/ROADMAP.md).

The major milestone is eBPF-based runtime learning: process execution, TCP connections, DNS activity, file opens and process-to-socket correlation.

## Security

AliveSpec observes operational metadata. It should **never capture application payloads, credentials, private keys, or secret values**.

See [`SECURITY.md`](SECURITY.md).

## License

MIT
