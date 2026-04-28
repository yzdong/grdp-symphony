# Workflow contract for Symphony agents

This file lives at the root of the grdp fork and is read by Symphony for every
issue. It defines the policy for autonomous coding-agent runs.

## Goal of every run

Implement the issue exactly as scoped. The issue title and body are
self-contained — do not expand scope beyond what the acceptance criteria list.

## Allowed actions

- Read any file in this repo.
- Browse `xfreerdp` source on GitHub (`https://github.com/FreeRDP/FreeRDP`).
- Browse Microsoft Open Specifications.
- Build (`go build ./...`) and run tests (`go test ./...`).
- Run the scenario suite at `/opt/bench/scenarios/lib/score.sh` to verify your
  changes pass the relevant rubric checks.

## Disallowed actions

- Do not modify `/opt/bench/scenarios/` — that is the scoring rubric.
- Do not push to upstream `nakagami/grdp`. Use this fork only.
- Do not bypass the LiteLLM proxy (it is set as `OPENAI_BASE_URL`).

## Definition of done

A PR that:
1. Closes the issue's acceptance criteria.
2. Adds tests for the new code.
3. Passes `go build ./...` and `go test ./...`.
4. Does not regress any previously-passing scenario.
5. Has a clear PR description: what changed, why, and (if ported from
   xfreerdp) attribution.

## When stuck

If the issue is blocked by a missing dependency that should be a separate
issue, write a short PR comment naming the missing dependency and stop. Do not
write the dependency yourself in this run — Symphony will route it.
