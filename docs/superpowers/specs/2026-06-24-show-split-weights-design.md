# Design: `testtrack show <split>`

## Problem

There is no way to read the current variant weights of an ongoing split / feature
gate directly from the CLI. Today the only options are:

- `testtrack sync` (mutates the local schema with remote weights), then open the file
- the TestTrack admin UI
- a raw `curl` against `api/v*/split_registry`

A read-only command fills this gap and reuses the CLI's existing server-read plumbing.

## Command surface

- `testtrack show <split-name>` — required single arg (`cobra.ExactArgs(1)`).
  The arg is the fully-qualified split name (e.g.
  `retail.cash_in_portfolios_q2_2026_enabled`) and is matched verbatim against
  registry keys. No auto-prefixing — predictable, and matches how `sync` keys off
  full names.
- `--json` flag — emit the raw weights map for scripting.
- Uses `TESTTRACK_CLI_URL` via `servers.New()`, exactly like `sync` (errors clearly
  if unset).

## Implementation

New file `cmds/show.go`, registered in its `init()` via
`rootCmd.AddCommand(showCommand)`, mirroring `cmds/sync.go`.

```go
func Show(name string, asJSON bool) error {
    server, err := servers.New()
    if err != nil {
        return err
    }

    var registry serializers.RemoteRegistry
    if err := server.Get("api/v2/split_registry.json", &registry); err != nil {
        return err
    }

    split, ok := registry.Splits[name]
    if !ok {
        return fmt.Errorf("split %q not found in remote registry; check the name or run `testtrack sync`", name)
    }
    // print split.Weights — table by default, JSON if asJSON
}
```

Reuses `serializers.RemoteRegistry` / `RemoteRegistrySplit` as-is (they already
deserialize `api/v2/split_registry.json`). No struct changes, no new dependencies.

## Output

Default — readable, one variant per line, sorted by variant name for deterministic
output:

```
retail.cash_in_portfolios_q2_2026_enabled
  false  100%
  true     0%
```

`--json`:

```
{"false":100,"true":0}
```

## Testing

Follows the repo pattern: the `fakeserver` package + `servers.IServer` interface
inject a fake registry response. Tests assert:

1. weights found → correct table / JSON output
2. split absent → error including the hint
3. `TESTTRACK_CLI_URL` unset → error from `servers.New()`

## Scope / YAGNI

Out of scope: `--local`, listing all splits, assignment counts, showing the
`feature_gate` flag (the v2 struct doesn't carry it today; a separate enhancement).

## Docs

Add a short `testtrack show` entry to `README.md` near the `sync` section.
