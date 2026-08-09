# graphql

**n/a** when no graphql server dependency

## detected imports

| library | import path |
|---------|-------------|
| gqlgen | `github.com/99designs/gqlgen` |
| graphql-go | `github.com/graph-gophers/graphql-go` |
| graphql (code-first) | `github.com/graphql-go/graphql` |

## checks

| id | status | when | why | fix | severity |
|----|--------|------|-----|-----|----------|
| gql-01 | implemented | sql/db driver imported in resolver package | resolver does storage | delegate to service/repo | fail |
| gql-02 | implemented | field resolver calls db per parent row (heuristic: query in resolver without batch) | slow lists; query in a loop | dataloader, join in repo, or sql batch | warn |
| gql-03 | implemented | hand-edited `generated.go` / `models_gen.go` | regen overwrite | change schema + `gqlgen generate` only | fail |
| gql-04 | implemented | no complexity or depth limit in server setup | expensive query dos | gqlgen extension or custom middleware | warn |
| gql-05 | implemented | introspection enabled without env gate | schema dump in prod | disable or auth-gate introspection | warn |
| gql-06 | implemented | graphiql/playground route registered unconditionally | public schema UI | build tag or dev-only route | fail |
| gql-07 | implemented | `map[string]interface{}` in hand-written resolvers (gqlgen) | type safety lost | generated models | warn |
| gql-08 | implemented | resolver file >500 lines without split | hard review | split by domain or thin delegates | info |
| gql-09 | implemented | get-based graphql queries without size limit | cache poisoning / oversized urls | post-only or limit query string | info |

## overlap checks

| id | status | overlap | notes |
|----|--------|---------|-------|
| sec-g01 | overlap → gql-05 | introspection without env gate | security gate |
| sec-g02 | overlap → gql-04 | no query depth limit | security gate |
| sec-g03 | overlap → gql-04 | no complexity/cost limit | security gate |

## gqlgen-specific

| id | status | when | fix |
|----|--------|------|-----|
| gqlgen-01 | implemented | `gqlgen.yml` missing | add config; pin generate in ci |
| gqlgen-02 | implemented | federation `entity.resolvers` without auth on id | check actor owns entity |

## graphql-go / code-first

| id | status | when | fix |
|----|--------|------|-----|
| ggl-01 | implemented | no panic recovery at execute boundary | wrap executor |
| ggl-02 | implemented | legacy `graphql-go/graphql` only | consider gqlgen for new work (info) |

## security

see [security-owasp.md](security-owasp.md) — sec-g01…sec-g10
