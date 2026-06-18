# backtest_e2e

`backtest_e2e` is the end-to-end smoke harness for the Backtest engine.
It glues three things together so a single command proves the whole
data path is healthy:

```
Tushare ingestion  ->  Market Data Postgres  ->  Backtest engine
```

The harness is intentionally a CLI, not a service: it is meant to run in
CI / local pre-release as a regression check, not to serve traffic.

## What it does

For each scenario it walks three optional stages:

1. **ingest** - drives `market.IngestionService` against Tushare to fill
   the security master, trading calendar, daily bars, financials, and
   factors needed by the scenario. Skipped when `--skip-ingest` is set
   or when the scenario sets `ingest.skip: true`.
2. **run** - constructs the same dependency graph
   `app/backtest/internal/svc.NewServiceContext` would, then drives a
   single `appBacktest.Service.Run` over the scenario's range / formula
   / universe and prints the resulting `PerformanceReport` summary.
3. **baseline** - compares the new metrics against the JSON baseline
   stored beside the scenario file. Missing baselines are written on
   first run; subsequent runs diff against them and exit non-zero when
   any metric drifts past the configured tolerance.

## Quick start

```
# 1. one-time: TUSHARE_TOKEN must be set (see .env)
# 2. start docker compose so quantlab-postgresql is up
# 3. run the bundled scenario
go run ./cmd/backtest_e2e -scenario cmd/backtest_e2e/scenarios/cn_smoke.yaml
```

Use `--update-baseline` to overwrite the saved metrics. Use
`--skip-ingest` once the database is warm so the iteration loop does
not hit Tushare every time.

## Scenarios

Scenarios are YAML files. See [scenarios/cn_smoke.yaml](scenarios/cn_smoke.yaml)
for the canonical example. The schema mirrors the application
`CreateBacktestRequest` plus an `ingest` block describing what to pull
from Tushare. Scenarios are not part of the production binary; they are
checked into the repo so reviewers can reproduce runs deterministically.

## A note on Tushare permissions

The bundled `cn_smoke` scenario only requires the free `daily` and
`trade_cal` endpoints. Financial-statement / daily_basic ingestion is
left commented out because it requires Tushare premium points; flip the
relevant blocks back on once your token has access.

`stock_basic` is rate limited to one call per hour on free accounts. If
the harness hits 40203 during a fresh ingest, wait an hour or run with
`-skip-ingest` against a previously populated database.

## Stock-code convention

Bars in the database keep the bare code (e.g. `000001`); the Tushare
adapter strips the suffix on the way in. Scenario universes follow the
same convention: write `000001`, not `000001.SZ`.
