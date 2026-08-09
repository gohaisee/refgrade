# graphql

**n/a** when no graphql server dependency

## detected imports

| library | import path |
|---------|-------------|
| gqlgen | `github.com/99designs/gqlgen` |
| graphql-go | `github.com/graph-gophers/graphql-go` |
| graphql (code-first) | `github.com/graphql-go/graphql` |

## checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| gql-01 | sql/db driver imported in resolver package | resolver does storage | delegate to service/repo | fail |
| gql-02 | field resolver calls db per parent row (heuristic: query in resolver without batch) | slow lists; query in a loop | dataloader, join in repo, or sql batch | warn |
| gql-03 | hand-edited `generated.go` / `models_gen.go` | regen overwrite | change schema + `gqlgen generate` only | fail |
| gql-04 | no complexity or depth limit in server setup | expensive query dos | gqlgen extension or custom middleware | warn |
| gql-05 | introspection enabled without env gate | schema dump in prod | disable or auth-gate introspection | warn |
| gql-06 | graphiql/playground route registered unconditionally | public schema UI | build tag or dev-only route | fail |
| gql-07 | `map[string]interface{}` in hand-written resolvers (gqlgen) | type safety lost | generated models | warn |
| gql-08 | resolver file >500 lines without split | hard review | split by domain or thin delegates | info |
| gql-09 | get-based graphql queries without size limit | cache poisoning / oversized urls | post-only or limit query string | info |

## security (catalog overlap — v1.0.0)

| id | status | implemented as |
|----|--------|----------------|
| sec-g01 | covered-by | [gql-05](#checks) introspection without env gate |
| sec-g02 | covered-by | [gql-04](#checks) no query depth limit |
| sec-g03 | covered-by | [gql-04](#checks) no complexity/cost limit |

## gqlgen-specific

| id | when | fix |
|----|------|-----|
| gqlgen-01 | `gqlgen.yml` missing | add config; pin generate in ci |
| gqlgen-02 | federation `entity.resolvers` without auth on id | check actor owns entity |

## graphql-go / code-first

| id | when | fix |
|----|------|-----|
| ggl-01 | no panic recovery at execute boundary | wrap executor |
| ggl-02 | legacy `graphql-go/graphql` only | consider gqlgen for new work (info) |

## security overlap

see [security-owasp.md](security-owasp.md) — sec-g01…sec-g10
