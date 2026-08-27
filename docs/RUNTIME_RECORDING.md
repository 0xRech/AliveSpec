# Runtime recording

`alivespec record` is the first automatic learning mode in AliveSpec.

## Goal

A recording represents a **known-good successful journey**, for example:

```text
login
upload-document
generate-pdf
search-customer
```

During a short recording window AliveSpec observes runtime metadata and compiles candidate requirements into an operational contract.

## Current Linux backend

The first backend uses `bpftrace` and Linux tracepoints. This deliberately keeps the AliveSpec Go build independent from generated BPF object files while the event model is still evolving.

The backend currently observes:

- `sched:sched_process_exec` for process execution
- `syscalls:sys_enter_connect` for outgoing IPv4 TCP endpoints
- `syscalls:sys_enter_openat` for file-open metadata

AliveSpec does **not** capture packet payloads or file contents.

## Recommended recording

Always filter to the processes involved in the journey when possible:

```bash
sudo alivespec record login \
  --comm nginx \
  --comm myapp \
  --duration 20s
```

Without a `--comm` filter the observer records system-wide metadata. AliveSpec warns about this because unrelated background activity can become false candidate dependencies.

## File noise filtering

By default the contract compiler keeps likely configuration files and drops common runtime noise from `/proc`, `/sys`, `/dev`, `/tmp`, shared-library paths and similar locations.

Use `--all-files` only when debugging the learning engine:

```bash
sudo alivespec record login --comm myapp --all-files
```

File contents are never stored. Observed files are existence requirements only; manually declared files can still be pinned by SHA-256 using `alivespec learn`.

## Evidence

Automatically learned requirements carry evidence metadata:

```yaml
evidence:
  source: observed
  observations: 6
  confidence: 0.9
  processes:
    - myapp
```

Current alpha confidence values are intentionally simple constants. A later release will calculate confidence across repeated successful journeys and allow users to confirm, ignore or promote observations.

## Current limitations

- Linux only
- requires `bpftrace` and BPF privileges
- IPv4 TCP connect observation only
- DNS names are not learned yet; connections are stored as observed IP endpoints
- process descendants are not automatically followed when their `comm` changes
- an observed dependency is correlation, not proof of necessity

These limitations are why the feature is marked alpha.

## Planned native backend

The observer interface is backend-independent. A native Go backend based on `cilium/ebpf` is planned once the event schema stabilizes. The CLI and generated contract format should not need to change.
