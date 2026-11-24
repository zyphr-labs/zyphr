# Zyphr

A local security record for coding agents, whichever harness runs them.

Zyphr records what a coding agent did and where its traffic went, joins both
into sessions, and explains security findings with the evidence behind them.
One event model, one collector, one console, and one policy file serve every
harness. Each harness plugs in through a thin adapter.

**Status: under active development. There is no release yet.** The pieces
arrive through pull requests in this order: core, adapter framework, harness
adapters, egress proxy, findings and console, release.

- [Product contract](PRODUCT.md): what Zyphr is, its scope, and its invariants.
- [Architecture](docs/architecture.md): data flow, packages, and the event model.
- [Adapter contract](docs/adapters.md): how a harness plugs in.

## Build and verify

Use Linux or macOS and the Go version in `go.mod`.

```bash
go build -o bin/zyphr ./cmd/zyphr
./bin/zyphr version
./verify
```

## License

MIT. See [LICENSE](LICENSE).
