# Zyphr

Zyphr is a local, harness-agnostic security record for coding agents.

Read `PRODUCT.md` before any change. It holds the product boundary and the
invariants. Read `docs/architecture.md` before you change the core, and
`docs/adapters.md` before you add or change a harness adapter.

## Rules

- The core never imports a harness package. A harness exists only under
  `internal/harness/<name>` and in the registry at `internal/harness`.
- Never install hooks into live user configuration, never change the operating
  system trust store, and never capture real credentials. Tests use isolated
  `ZYPHR_CONFIG_DIR`, `ZYPHR_DATA_DIR`, and `ZYPHR_SOCKET` values.
- A hook, adapter, or collector failure must reply permissively. Test this for
  every change on the hook path.
- The proxy forwards application bytes unchanged. Bound the inspected copy,
  not the forwarded traffic.
- Adapter payload shapes come from vendor documentation or observed payloads.
  Do not guess. Mark each fixture `observed` or `documented`.
- Keep observed evidence, policy decisions, inference, and unavailable data
  distinct in code, output, and docs.
- A support claim needs a record in `docs/validation.md`. Do not raise an
  adapter to `verified` without a live smoke test.
- Keep secrets, private prompts, generated keys, databases, and build output
  out of Git.
- Write plain prose with no em dashes.

## Verify

Run `./verify` before you push. It stops at the first failure. It needs the Go
version in `go.mod` and a C compiler for the race detector. Release binaries
stay CGO-free.

## Workflow

- Branch from `main`, open a pull request, and merge only with green checks.
- Extra checkouts go in `worktrees/<branch>` inside this repository.
- Record unfinished work and real verification evidence in `NEXT.md`.
