# Architecture

Read [PRODUCT.md](../PRODUCT.md) first. This file describes how the code is
arranged to keep that contract.

## Data flow

```
 harness A   harness B   harness C        any agent process
     |           |           |                    |
  command      shim       command           HTTP(S) through
   hook         |          hook            the scoped proxy
     \          |           /                     |
      zyphr hook <harness> <hook>            zyphr run
      (adapter.Decode / adapter.Reply)       (internal/proxy)
                 \                               /
                  \        normalized events    /
                   +-----> collector <---------+
                           (Unix socket)
                               |
                         policy  findings
                               |
                             SQLite
                               |
              CLI queries, NDJSON export, local console
```

Everything to the left of the collector is harness-specific and lives in an
adapter. Everything from the collector onward is the core and knows only the
normalized event.

## Packages

| Package | Responsibility |
| --- | --- |
| `cmd/zyphr` | Process entry. Calls `internal/cli` |
| `internal/cli` | Command parsing and output |
| `internal/event` | The normalized event, its validation, and the JSON schema |
| `internal/paths` | Config, data, and socket locations from `ZYPHR_*` and XDG defaults |
| `internal/store` | SQLite schema, writes, and queries for events, runs, sessions, findings |
| `internal/collector` | Unix-socket server, RPC, ingestion, retention |
| `internal/client` | RPC client used by the CLI, the hook entry point, and the proxy |
| `internal/adapter` | Adapter interface, capability model, conformance suite |
| `internal/harness` | The registry: the one place that lists every adapter |
| `internal/harness/<name>` | One package per harness. Imports only `adapter`, `event`, and `install` types |
| `internal/install` | Plan executor: atomic writes, backups, ownership manifest, uninstall |
| `internal/hook` | The `zyphr hook` entry point: decode, send, decide, reply, fail open |
| `internal/proxy` | Per-run HTTP(S) proxy, local CA, byte-exact forwarding, bounded inspection |
| `internal/policy` | Rule file parsing and evaluation for both planes |
| `internal/findings` | Detectors that read stored events and write evidence-linked findings |
| `internal/redact` | Secret redaction for retained text |
| `internal/console` | Local web UI and its JSON API, embedded static assets |

Dependency rule: nothing outside `internal/harness` imports
`internal/harness/<name>`. Only the registry does. A check in `./verify`
enforces this.

## The normalized event

One event shape serves both planes. Fields that do not apply are omitted.

| Group | Fields |
| --- | --- |
| Identity | `schema_version`, `id`, `ts`, `run_id`, `session_id` |
| Origin | `harness`, `harness_version`, `source` (`adapter`, `proxy`, `gap`, `system`), `native_hook`, `native_tool` |
| Action | `action`, `stage` (`intent`, `result`) |
| Tool detail | `target`, `command`, `cwd`, `mcp_server`, `mcp_tool`, `exit_code` |
| Network detail | `method`, `host`, `url`, `status`, `bytes_out`, `bytes_in` |
| Decision | `decision` (`observe`, `allow`, `deny`), `rule_id` |
| Evidence | `capture` (`complete`, `partial`, `disabled`, `unavailable`), `note` |
| Opt-in text | `prompt`, `response` (dropped unless capture is enabled) |

Actions: `session_start`, `session_end`, `prompt`, `response`, `file_read`,
`file_write`, `file_delete`, `command_exec`, `mcp_call`, `web_fetch`,
`web_search`, `subagent_start`, `subagent_stop`, `tool_use`, `net_request`,
`gap`. `tool_use` is the explicit marker for a native tool that the adapter
could not map with certainty.

The Go definition in `internal/event` is the reference.

`run_id` comes from `ZYPHR_RUN_ID`, which `zyphr run` sets for its child
process tree. `session_id` is the harness's own session identity. Either can be
absent. An event needs at least one.

## The hook path

`zyphr hook <harness> <hook-name>` runs once per native hook event. It is the
hot path, so it stays small:

1. Read stdin with a size bound.
2. `adapter.Decode` to normalized events.
3. Send to the collector with a short deadline. The collector evaluates policy
   for intent-stage events and returns a decision.
4. `adapter.Reply` and exit.

Any error, at any step, produces the permissive reply. The entry point never
waits beyond the time budget.

## Policy

One JSON file, `policy.json` in the config directory, holds rules for both
planes. A rule has an ID, a plane (`tool` or `egress`), match fields, and a
decision (`observe`, `allow`, or `deny`). Rules are evaluated in order. The
first match wins. With no match, the decision is `observe`.

The collector is the only decision point for the tool plane. The proxy
evaluates egress rules itself, because it must decide before it forwards.

## The proxy

`zyphr run` creates a run ID, starts a proxy on its own loopback listener,
sets the proxy and CA environment for the child only, starts the child, and
removes the listener when the child exits. It does not touch the operating
system trust store.

The proxy forwards every byte unchanged. It tees a bounded copy for inspection.
When the copy is cut short, the event says `capture: partial`.

## Storage and privacy

SQLite in the data directory holds events, runs, sessions, and findings.
Retention defaults to 90 days and is configurable. Prompt and response text is
stored only when capture is enabled, after redaction, and encrypted with a
local key.

## The console

`zyphr ui` serves embedded static files and a small JSON API on loopback, with
a per-launch token. It has no build step and makes no external requests. It is
read-only: it reads through the collector and changes nothing.
