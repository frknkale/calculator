# Test coverage

## Backend

| Package | Statement coverage |
| --- | --- |
| `internal/calculator` | 100.0% |
| `internal/httpapi` | 100.0% |
| `cmd/server` | 0.0% |
| Total | 83.7% |

Server startup is not covered by unit tests. It is included in the total.

From `backend/`:

```sh
go test -race -coverprofile=../reports/backend.out ./...
go tool cover -func=../reports/backend.out
go tool cover -html=../reports/backend.out -o ../reports/backend.html
```

Reports: [HTML](backend.html), [Go coverage profile](backend.out).

## Frontend

108 tests pass across 4 test files.

| File | Statements | Branches | Functions | Lines |
| --- | --- | --- | --- | --- |
| `App.tsx` | 100% | 100% | 100% | 100% |
| `api.ts` | 100% | 96.29% | 100% | 100% |
| `operations.ts` | 100% | 100% | 100% | 100% |
| `formatResult.ts` | 100% | 100% | 100% | 100% |
| Total | 100% | 99% | 100% | 100% |

Coverage measures these four source files. It excludes the browser entry point, CSS, configuration, and tests. Minimum coverage is 90% for statements, functions, and lines, and 85% for branches.

From `frontend/`:

```sh
pnpm test:coverage
```

Reports: [JSON summary](frontend/coverage-summary.json), [LCOV](frontend/lcov.info). The command also generates a local HTML report at `reports/frontend/index.html`.
