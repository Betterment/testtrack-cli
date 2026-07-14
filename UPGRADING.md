# Upgrading testtrack-cli

## 1.x to 2.0.0

2.0.0 changes the on-disk schema format (`testtrack/schema.{yml,json}`).

### What changed

The schema's applied-migration high-water mark used to be a single scalar
field:

```yaml
serializer_version: 1
schema_version: "2020011543200"
```

Every new migration rewrote that one line, so any two branches that added
migrations conflicted on it. 2.0.0 replaces the scalar with a list of every
applied migration version (`serializer_version` bumps to `2`):

```yaml
serializer_version: 2
schema_versions:
- "2020060151730"
- "2020011543200"
```

The list is ordered by a hash of each version (not chronologically), so
concurrently added migrations scatter through the file instead of clustering
on one line. The ordering is otherwise meaningless — don't rely on it.

### This is a breaking format change

A 2.0.0 CLI **refuses to read** an old (`serializer_version: 1`) schema file.
Any command that reads the schema errors and tells you to run `testtrack schema
upgrade` — `create`, `migrate`, `sync`, `decide`, the `destroy` commands, and
`schema load` (essentially everything except `assign`/`unassign`, which use a
different read path). This is deliberate: the old format can't be losslessly
read into the new struct, so the CLI makes you convert it explicitly rather than
silently round-tripping a lossy upgrade.

Conversely, a pre-2.0.0 CLI does **not** understand `schema_versions`. If it
reads a v2 schema it ignores the field, and any command that rewrites the
schema produces a hybrid: the file keeps `serializer_version: 2` (the old CLI
round-trips whatever number it read) but is otherwise back in the v1 shape — a
scalar `schema_version`, no `schema_versions` list. A 2.0.0 CLI detects the
leftover `schema_version` field, refuses to read the hybrid, and tells you to
run `testtrack schema upgrade`, which repairs it by rebuilding the version list
from `testtrack/migrate`. So the whole team — and every CI/build environment —
has to be on 2.0.0 before you commit a v2 schema, or the format will flip-flop
between commits.

Because a 2.0.0 CLI rejects a v1 schema and a 1.x CLI mangles a v2 schema, the
switch is a flag day: the moment the v2 schema lands in the repo, every consumer
of it must already be on 2.0.0. Order the rollout so the CLI bump and the schema
commit line up, not in separate steps.

1. **Make 2.0.0 available, but don't switch anything to it yet.** Cut the
   `v2.0.0` release, update the brew formula, and make the pinned binary
   available to CI — without yet changing the CI pin or asking devs to upgrade.

2. **Land the schema conversion and the CI bump together, in one change.** In
   your app root:

   ```sh
   testtrack schema upgrade
   ```

   This converts `testtrack/schema.{yml,json}` to the v2 format **in place**: it
   keeps the materialized state already in the file (splits, decisions,
   retirements, etc.) and only rebuilds the `schema_versions` list from the
   migration filenames in `testtrack/migrate`. In the **same** PR, bump the
   CI/build pin to `v2.0.0`. They must be atomic: if CI runs 2.0.0 against a
   still-v1 schema, `testtrack migrate` errors; if the v2 schema lands while CI
   still runs 1.x, the old binary rewrites it into the hybrid described above,
   and every 2.0.0 user is blocked until someone re-runs `schema upgrade`.
   Expect a large
   one-time diff — the new `schema_versions` block lists every migration — but
   the splits body is left untouched.

3. **Developers upgrade as that change lands** (`brew upgrade testtrack-cli`). A
   developer must be on 2.0.0 before running any testtrack command against the
   upgraded schema — an older binary silently rewrites it into the hybrid,
   which 2.0.0 CLIs then refuse to read until it's repaired with
   `schema upgrade`.

### Notes

- **Use `schema upgrade`, not `schema generate`, to convert.** `generate`
  rebuilds the schema by replaying every migration from scratch, which fails on
  many long-lived apps — e.g. when `testtrack/migrate` predates some splits, or
  contains a decision/retirement for a split that was created out of band in the
  TestTrack admin (so there's no create migration to replay). `upgrade` doesn't
  replay; it preserves the already-correct materialized state and just restamps
  the format, so it works regardless.
- `schema upgrade` reads the old file directly (it's the one command that
  bypasses the version guard), so it's always the recovery path if you hit the
  "older schema format" error.
- No server-side change is required. The TestTrack server tracks applied
  migrations itself; `schema_version`/`schema_versions` is a CLI-only artifact.
