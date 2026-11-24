# Zyphr product contract

This file is the source of truth for what Zyphr is and what it must never do.
Code, docs, and the marketing site follow it. Change this file first when the
product changes.

## What Zyphr is

Zyphr is an open-source, local security record for coding agents. It records
what an agent did and where its traffic went, joins both into sessions, and
explains security findings with the evidence behind them.

Zyphr is harness-agnostic. A developer who runs Claude Code today and Codex,
Cursor, OpenCode, or Pi tomorrow gets the same record, the same findings, and
the same rules. No harness is special in the core.

## Who it is for

1. First: one engineer who runs several coding agents on their own machine and
   wants to know what those agents did.
2. Later: security teams that want to let engineers use coding agents without
   losing sight of actions and data.

Version 0.1 serves the first group. It must be useful in daily work before
anyone else is asked to install it.

## Success criteria

- An engineer can leave Zyphr on during normal work. It does not slow, break,
  or block the agent.
- An investigation gets evidence that is correct, and Zyphr states where its
  visibility ends.
- Adding a harness is small, routine work: one adapter package, its fixtures,
  and one registry line.

## Principles

1. **Harness-agnostic core.** The event model, collector, store, policy,
   findings, and console contain no harness-specific code. A harness exists
   only as an adapter.
2. **One adapter interface.** Every harness plugs into the same contract. See
   [docs/adapters.md](docs/adapters.md).
3. **Engineering work comes first.** Zyphr observes by default. It denies only
   when a rule that the user wrote says deny. A Zyphr failure never blocks a
   tool call.
4. **Evidence, not claims.** Observed facts, policy decisions, and inferences
   stay distinct. Missing data is unknown. It is never zero and never "safe".
5. **Coverage is part of the result.** Each adapter declares what it can see.
   What Zyphr cannot see is shown as a gap.
6. **Local.** No cloud service, no account, no telemetry, no remote upload.

## Scope of version 0.1

### The two planes

- **Tool plane.** Adapters record native agent events: sessions, tool intent
  and result, file reads and writes, commands, MCP calls, web fetches, and
  subagents. Prompt and response text is opt-in.
- **Egress plane.** `zyphr run -- <agent>` starts the agent behind a
  process-scoped HTTP(S) proxy. It records each request and applies explicit
  egress rules. This plane is the same for every harness, including ones with
  no adapter.

### The core

- One Go binary with no CGO, no database server, and no cloud account.
- A local collector on a Unix socket with SQLite storage.
- One normalized event schema with run identity, native session identity,
  capture status, and explicit gap events.
- One policy file for both planes. The default decision is observe.
- Evidence-linked findings.
- A local browser console for sessions, events, findings, policy, and coverage.
- NDJSON export.
- A synthetic demo that needs no agent and no API key.

### Harness adapters

Each adapter has a support level:

- `verified`: a live smoke test on a named harness version passed, and
  [docs/validation.md](docs/validation.md) records it.
- `fixtures`: built from the vendor's documented hook contract and tested with
  fixtures only. No live run yet.
- `wrap-only`: the harness has no usable hook surface. Zyphr records egress
  through `zyphr run`, and the tool plane is a declared gap.

Targets for 0.1:

| Harness | Integration mechanism | Target level |
| --- | --- | --- |
| Claude Code | Command hooks in `settings.json` | verified |
| Codex CLI | Command hooks in `hooks.json` | verified |
| Cursor (IDE and `cursor-agent`) | Command hooks in `hooks.json` | verified |
| OpenCode | Plugin shim | verified |
| Pi | Extension shim | verified |
| Antigravity CLI (`agy`) | Hook or plugin surface, if one exists | verified, else wrap-only |
| Gemini CLI | Command hooks in `settings.json` | fixtures |
| GitHub Copilot CLI | Command hooks | fixtures |
| Windsurf | Command hooks | fixtures |
| Factory Droid | Command hooks | fixtures |
| Amp, Kiro, Cline, Grok Build, others | Add when the vendor documents a hook surface | fixtures or wrap-only |
| Any other agent | None needed | wrap-only |

An adapter ships only when its hook contract comes from vendor documentation or
from an observed live payload. Guessed payload shapes do not ship.

## Invariants

These are product rules, not implementation notes. A change that breaks one is
a bug.

### Do not disturb the agent

- A hook, adapter, or collector failure never denies or delays a tool call
  beyond the hook time budget. This includes a missing binary, a dead
  collector, a full disk, a malformed payload, and a timeout.
- The hook path has a hard time budget. The target is under 25 ms at p95 on a
  warm machine. On timeout it gives up and replies permissively.
- The proxy never changes application bytes. An inspection budget limits the
  copy that Zyphr keeps, not the traffic that Zyphr forwards.
- The proxy never retries an uncertain write.

### Deny only on an explicit rule

- With no policy file, nothing is denied.
- A deny on the tool plane applies only to intent-stage events, and only for
  adapters that declare they can block.
- Every deny is recorded with the rule that caused it.
- Proxy failure is different from hook failure and is documented as such: a
  setup failure stops `zyphr run` before the agent starts, and a transport
  failure can break a request.

### Identity is exact

- Every `zyphr run` owns its proxy listener and its run ID. Zyphr never infers
  identity from timestamps or from a shared process environment.
- A run and a native agent session are different things. A query for one never
  returns events that belong only to the other.

### Evidence is honest

- Each event states its source and its capture status: `complete`, `partial`,
  `disabled`, or `unavailable`.
- An unknown native tool becomes an explicit unmapped event. Zyphr never
  guesses that it was a file or shell action.
- A pre-tool event is intent, not proof that the action ran.
- A sequence such as "read a secret file, then connect to a host" is a lead.
  It is not proof that the data left the machine.
- Unknown counts stay unknown. Zyphr never shows them as zero.

### Installation coexists

- Install and uninstall change only entries that Zyphr owns, tracked in an
  ownership manifest. Unrelated hooks and settings survive, including a mixed
  hook group.
- The first backup survives a reinstall.
- An unsupported scope is rejected before any file is written. A project-scoped
  operation never writes user configuration.
- Zyphr rejects a symlinked target and writes files atomically.

### Privacy

- Prompt and response capture is off by default. When enabled, text is
  redacted, then encrypted at rest.
- Metadata is the default evidence. Zyphr does not store file contents.
- Exports can contain private paths and commands. Zyphr says so at export time.

## Security boundary

Zyphr 0.1 gives visibility within one user account. It does not claim more.

- Code that runs as the same user can change Zyphr hooks, binaries, keys, and
  data.
- The proxy sees only traffic from the process tree that honors the proxy and
  CA environment. Raw sockets, QUIC, certificate pinning, and an agent backend
  hosted by the vendor are outside it.
- An adapter sees only what its harness exposes. A renamed or new native tool
  is unmapped until the adapter learns it.
- Local encryption does not protect data from a process that can read the key.

## Out of scope for 0.1

Each item below is a later stage. None of them is implied, faked, or claimed.

- OS-level process and connection sensors (eBPF, Endpoint Security, Network
  Extension).
- Fleet management, organization policy distribution, and a hosted backend.
- Tamper protection against the same user.
- A model-based judge on the tool path or the egress path.
- Token and cost accounting.
- SIEM connectors beyond NDJSON export.
- Windows support.

## Development and verification rules

- Tests and development never install into live user configuration, never
  change the operating system trust store, and never capture real credentials.
- Tests use isolated `ZYPHR_CONFIG_DIR`, `ZYPHR_DATA_DIR`, and `ZYPHR_SOCKET`
  values and loopback servers. They need no model API.
- A live smoke test of a harness uses a disposable project and project scope,
  or a config directory that the harness lets us relocate.
- `./verify` is the single verification entry point. It stops at the first
  failure.
- A support claim needs a validation record. Source code that exists is not
  evidence that it works.

## Project identity

- Source: `github.com/zyphr-labs/zyphr`. License: MIT.
- CLI: `zyphr`. Environment prefix: `ZYPHR_`.
- Configuration: `~/.config/zyphr/`. Data: `~/.local/share/zyphr/`.
- Lantern is the predecessor experiment. Zyphr 0.1 is a fresh design and does
  not read Lantern data or hooks.
- The marketing site lives in the separate private repository `zyphr-site`.
