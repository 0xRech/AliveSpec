# Changelog

All notable changes to AliveSpec will be documented here.

## v0.2.0-alpha.1

First public alpha of the runtime-learning workflow.

### Added

- `alivespec record [journey]`
- Linux eBPF observation through a `bpftrace` backend
- process execution observation
- outgoing IPv4 and IPv6 TCP dependency observation
- best-effort DNS observation through glibc `getaddrinfo()` uprobes
- relevant configuration-file observation
- evidence provenance, observation counts and confidence metadata
- compact duplicate-free live discovery output
- Modern Ops terminal UI for `record`, `verify` and `diff`
- verification of learned processes and TCP dependencies
- diff support for learned runtime dependencies

### Notes

- Runtime recording currently requires Linux and elevated BPF permissions.
- DNS observation is best effort and may not work for static resolvers or non-glibc runtimes.
- The current backend uses `bpftrace`; a native Go/eBPF backend is planned.
