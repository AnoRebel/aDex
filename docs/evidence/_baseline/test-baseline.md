# Test Baseline

Captured: 2026-05-04

## Go

```
$ go build ./...   → clean (no output)
$ go vet  ./...    → clean
$ go test ./...    → all packages "no test files" except aDex-UI/tests/backend (ok, 0.010s)
```

The `tests/backend` package contains stubs for security tests (`security_auth_test.go`, `security_session_test.go`) but the package compiles without any actual test functions running because the `security` package internals aren't wired through. Counting them as 0 tests for baseline purposes.

## Frontend

`bun test` is not yet configured (Vitest is in deps but no `vitest.config.ts` exists). Section 0 will add one. Treat current frontend test count as 0.

## Targets after this change

| Layer       | Baseline | Target after Section 8 |
|-------------|----------|------------------------|
| Go tests    | 0        | ≥ 6 boundary tests (terminal, system, network, filesystem, theme, settings, coordinator-cwd) |
| Vitest      | 0        | ≥ 7 store/composable tests + ≥ 3 component tests |
| Evidence    | 0        | One `docs/evidence/<feature>/` per shipped feature, each with screenshot + notes.md |
