# sql и реляционные данные

**n/a** когда в модуле нет sql driver или orm

## detected stack

| tool | import / signal |
|------|-----------------|
| pgx v5 | `github.com/jackc/pgx/v5` |
| gorm | `gorm.io/gorm` |
| sqlx | `github.com/jmoiron/sqlx` |
| sqlc | `github.com/sqlc-dev/sqlc` (tool / generated queries) |
| ent | `entgo.io/ent` |
| pq (legacy pg) | `github.com/lib/pq` |
| mysql | `github.com/go-sql-driver/mysql` |
| sqlite | `modernc.org/sqlite`, `github.com/mattn/go-sqlite3` |

## universal sql checks

| id | when | why | fix | severity |
|----|------|-----|-----|----------|
| sql-01 | `Query`/`Exec`/вызов orm внутри `for` | запрос в цикле по спискам | batch `WHERE id = ANY($1)` или join | warn |
| sql-02 | sql собран через `fmt.Sprintf` с пользовательским вводом | инъекция | только параметризованные запросы | fail |
| sql-03 | функция `List*` / `FindAll*` без limit в sql | полный скан таблицы | жёсткий `LIMIT`; keyset pagination | warn |
| sql-04 | `SELECT *` на горячем пути (эвристика) | лишняя выборка | явно перечислить колонки | info |
| sql-05 | новый pool/соединение на каждый запрос | исчерпание fd | один pool на процесс | fail |
| sql-06 | `float64` для денежных полей | ошибки округления | целые минорные единицы или decimal type | warn |
| sql-07 | offset pagination на больших таблицах (комментарий/эвристика) | медленные страницы | keyset (`WHERE id > $cursor`) | info |

## pgx

| id | when | fix | severity |
|----|------|-----|----------|
| pgx-01 | `pgx.Connect` вместо pool в приложении | использовать `pgxpool.New` | warn |
| pgx-02 | сравнение ошибок pg по строке | использовать коды `pgconn.PgError` | info |
| pgx-03 | `sslmode=disable` в dsn для удалённого хоста | включить tls | fail |

## gorm

| id | when | fix | severity |
|----|------|-----|----------|
| gorm-01 | `db.Raw("... "+userInput)` | bind parameters | fail |
| gorm-02 | глубокие цепочки `Preload` на list endpoints | join или batch | warn |
| gorm-03 | `AutoMigrate` в `main` для prod-приложений | goose/golang-migrate | warn |
| gorm-04 | `Debug()` включён вне dev | убрать debug | warn |
| gorm-05 | package-level `var DB *gorm.DB` | внедрять через конструктор | warn |

## sqlx

| id | when | fix |
|----|------|-----|
| sqlx-01 | `Get`/`Select` без context | передавать `ctx` |

## sqlc / ent

| id | when | fix |
|----|------|-----|
| sqlc-01 | ручное редактирование `*.sql.go` | регенерировать из sql |
| ent-01 | обход edge в цикле без eager load | batch query или eager edge |

## security

см. [security-owasp.md](security-owasp.md) — sec-db01…sec-db05
