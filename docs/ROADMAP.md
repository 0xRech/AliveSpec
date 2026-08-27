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

- [ ] Linux eBPF observer
- [ ] PID ↔ TCP connection correlation
- [ ] DNS observation
- [ ] file-open observation
- [ ] bounded `alivespec learn <journey>` recording window
- [ ] evidence provenance
- [ ] dependency confidence scoring
- [ ] ignore / confirm workflow

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
