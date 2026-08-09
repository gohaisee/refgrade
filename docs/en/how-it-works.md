# how it works

refgrade is a cli, not a linter replacement

## pipeline

1. load module (`go/packages`)
2. **detect** stack — gin, gqlgen, pgx, mongo, …
3. run **universal** checks (config, layers, errors, tests, dead code)
4. run **stack** checks only when the library is present
5. mark missing stack as **n/a**, not fail
6. print report: text, markdown, or json

## commands (planned)

| command | role |
|---------|------|
| `scan` | main report |
| `detect` | print detected stack only |
| `init` | write `.refgrade.yaml` template |
| `explain <id>` | one check in human words |

## exit codes

- `0` — no fail-level findings
- `1` — at least one fail
- `2` — scan error (broken module, bad path)
