# Adapter contract

An adapter is the only place where Zyphr knows about a specific harness. The
core never imports an adapter package directly. It sees adapters only through
the registry and the interface below.

The goal: a new harness costs one small package, its fixtures, and one registry
line. It needs no change to the collector, store, policy, findings, or console.

## The interface

The Go shape below is the contract. Implementers can refine names and helper
types, but not the split of responsibilities.

```go
package adapter

type Adapter interface {
    // Info is static: ID, display name, mechanism, vendor docs URL,
    // capabilities, supported scopes, and support level.
    Info() Info

    // Detect reports whether the harness is present, its version, and the
    // config locations for each scope. It reads. It never writes.
    Detect(env Env) Detection

    // Plan returns the file changes that install this adapter for one scope,
    // as data. The adapter never writes files. The install engine applies
    // the plan.
    Plan(env Env, scope Scope) (install.Plan, error)

    // Decode turns one native hook payload into normalized events.
    Decode(hook string, payload []byte) ([]event.Event, error)

    // Reply renders a decision in the harness's native reply format:
    // stdout bytes, stderr text, and exit code.
    Reply(hook string, d Decision) Reply
}
```

### What each part owns

| Part | Owner | Rule |
| --- | --- | --- |
| Native payload shapes and tool names | Adapter | Mapped from vendor docs or observed payloads only |
| Native reply format | Adapter | The zero-value decision renders the permissive reply |
| Which files and entries to add | Adapter, as an `install.Plan` | Declarative. No file I/O in the adapter |
| Atomic writes, backups, ownership manifest, symlink checks, uninstall | Install engine | Shared. Written once |
| Socket I/O, timeouts, fail-open | Hook entry point | Shared. Written once |
| Run ID, timestamps, event IDs, capture status defaults | Hook entry point | Shared. Adapters do not invent identity |
| Policy decisions | Policy package | Adapters only render them |

## Mechanisms

Two mechanisms cover the harnesses known today. Both end at the same entry
point: `zyphr hook <harness> <hook-name>`, with the native payload on stdin and
the native reply on stdout.

1. **Command hooks.** The harness runs a command for each hook event and
   passes JSON on stdin. Claude Code, Codex, Cursor, Gemini CLI, Copilot CLI,
   Windsurf, and Droid work this way. The plan adds owned entries to the
   harness's hook configuration file.
2. **Shims.** The harness loads an in-process plugin or extension. OpenCode and
   Pi work this way. The plan writes one small generated file that forwards
   each event to `zyphr hook` and applies the reply. A shim holds no logic
   beyond forwarding. It must never throw into the harness and never return a
   blocking result when Zyphr fails.

A harness with neither mechanism is `wrap-only`. It still gets an `Info` entry
in the registry so that coverage output names it and states the gap.

## Capabilities

Each adapter declares what it can see. Coverage output, the console, and the
coverage docs are generated from these declarations. Nobody maintains a
coverage table by hand.

A capability set states, for this harness:

- which normalized actions it can report (`file_read`, `file_write`,
  `command_exec`, `mcp_call`, `web_fetch`, `prompt`, `response`,
  `session_start`, `session_end`, `subagent`, and so on),
- for each action, whether intent, result, or both are available,
- whether a pre-tool deny is honored by the harness (`can_block`),
- which scopes are supported (`user`, `project`),
- known gaps, as short plain statements.

## Decode rules

- Map a native tool to a specific action only when the mapping is certain.
- Any other tool becomes `tool_use` with `mapped: false`, and it keeps the
  native tool name. Never guess a file or shell action.
- Keep the native hook name and native tool name on the event.
- Send metadata, not content. No file bodies. No auth headers. Prompt and
  response text goes in the body fields only, and the collector drops it unless
  capture is enabled.
- Bound the payload before parsing. A malformed or oversized payload yields one
  gap event when possible, never a crash.
- Omit unknown values. Do not fill in zero, success, or an invented timestamp.

## Reply rules

- A permissive reply must be exactly what the harness documents as "continue".
  Some harnesses need a JSON body on stdout for some hooks. Others need an
  empty stdout and exit code 0.
- A deny reply is rendered only for a decision that the policy package made,
  only at intent stage, and only when the adapter declares `can_block`.
- On any internal error, reply permissively.

## Install rules

The install engine enforces these. An adapter cannot opt out.

- The hook command must fail open when the `zyphr` binary is missing or moved.
- Every added entry is recorded in the ownership manifest. Uninstall removes
  exactly those entries and leaves everything else as it is, including edits
  the user made after install.
- The first backup of a file is kept across reinstalls.
- A scope that the adapter does not support is rejected before any write.
- Project scope never touches user configuration.
- Config paths honor the harness's own relocation variables, such as
  `CLAUDE_CONFIG_DIR`, `CODEX_HOME`, and `CURSOR_CONFIG_DIR`.
- `--dry-run` prints the plan and writes nothing.

## Conformance suite

One shared test suite runs against every adapter in the registry. An adapter
that fails it does not ship. The suite checks at least:

1. `Info` is complete, and capabilities are consistent with the fixtures.
2. Every fixture in `testdata/` decodes to its expected normalized events.
3. Unknown tools decode to `tool_use` with `mapped: false`.
4. Malformed, empty, and oversized payloads do not panic and do not block.
5. The permissive reply for every declared hook matches the recorded expected
   reply.
6. A deny reply exists for every hook the adapter claims it can block.
7. Install into an empty config, and into a config with unrelated entries.
   Then install again, edit the config, and uninstall. Unrelated entries and
   later edits survive. No Zyphr entry is left.
8. Project scope writes nothing outside the project directory.
9. With the collector down, the hook entry point exits permissively inside the
   time budget.

## Adding a harness

1. Read the vendor's hook documentation. Record the URL and the harness version
   in the adapter's `Info`.
2. Capture real payloads from a disposable project when the harness is
   available locally. Scrub them. Store them as fixtures. Mark each fixture as
   `observed` or `documented`.
3. Write the package: `Info`, `Detect`, `Plan`, `Decode`, `Reply`.
4. Add the registry line.
5. Run the conformance suite.
6. Run a live smoke test when the harness is available: install at project
   scope in a disposable project, run one trivial prompt, confirm the expected
   events, uninstall, and confirm the config is back to its original state.
   Record the result in [validation.md](validation.md). Only then set the
   support level to `verified`.
