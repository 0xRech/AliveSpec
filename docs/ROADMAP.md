# Roadmap

## v0.1 — Contract foundation

- [x] YAML operational contract
- [x] `learn`
- [x] `verify`
- [x] `diff`
- [x] systemd service checks
- [x] TCP listener checks
- [x] DNS checks
- [x] TLS trust and certificate-lifetime checks
- [x] file fingerprint checks

## v0.2 — Observe real runtime journeys

### v0.2-alpha.1

- [x] observer abstraction
- [x] Linux eBPF recording through `bpftrace`
- [x] bounded `alivespec record <journey>` window
- [x] process execution observation
- [x] outgoing IPv4 TCP connection observation
- [x] process ↔ TCP dependency attribution
- [x] file-open observation with config-file noise filtering
- [x] evidence provenance
- [x] observation counts and initial confidence values
- [x] verification of learned processes and TCP dependencies
- [x] terminal UI layer for recording output

### Next v0.2 slices

- [ ] DNS observation without retaining query payloads
- [ ] IPv6 TCP observation
- [ ] process-tree / descendant tracking
- [ ] ignore / confirm workflow
- [ ] confidence scoring based on repeated successful journeys
- [ ] native `cilium/ebpf` backend
- [ ] graceful backend capability detection
- [ ] automated eBPF integration test environment

## v0.3 — Better operational proofs

- [ ] HTTP journey assertions
- [ ] dependency graph output
- [ ] JSON output for CI
- [ ] JUnit output
- [ ] baseline history
- [ ] known-good snapshots

## v0.4 — Change impact

- [ ] `alivespec whatif`
- [ ] reverse dependency lookup
- [ ] certificate blast-radius analysis
- [ ] DNS change impact
- [ ] service shutdown impact

## v1.0

AliveSpec should be able to observe a successful Linux application journey and generate a useful executable contract with minimal manual hints.
