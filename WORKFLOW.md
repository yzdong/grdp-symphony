---
tracker:
  kind: linear
  project_slug: "rdp-bench-experiment-e2d36e4dca0e"
  active_states:
    - Todo
    - In Progress
  terminal_states:
    - Done
    - Cancelled
    - Canceled
    - Duplicate
polling:
  interval_ms: 30000
workspace:
  root: /home/bench/symphony-workspaces
hooks:
  after_create: |
    set -e
    git clone --depth 1 https://github.com/yzdong/grdp-symphony .
    git config user.email "bench@example.com"
    git config user.name "bench"
agent:
  max_concurrent_agents: 4
  max_turns: 30
codex:
  command: codex --config 'shell_environment_policy.inherit=all' --config 'model="gpt-5"' app-server
  approval_policy: never
  thread_sandbox: workspace-write
  turn_sandbox_policy:
    type: workspaceWrite
---

You are working on Linear issue `{{ issue.identifier }}` for the rdp-bench experiment.

Issue context:
Identifier: {{ issue.identifier }}
Title: {{ issue.title }}
Current status: {{ issue.state }}
Labels: {{ issue.labels }}
URL: {{ issue.url }}

Description:
{% if issue.description %}
{{ issue.description }}
{% else %}
No description provided.
{% endif %}

## Goal

This is the symphony stack of an experiment comparing multi-agent orchestrators on
making `nakagami/grdp` (a pure-Go RDP client) closer to feature parity with `xfreerdp`.

The repo at `https://github.com/yzdong/grdp-symphony` is a working copy you may
push branches to and open PRs against. Your job: implement the Linear issue's
acceptance criteria.

## Constraints

- Single language: Go.
- Reference impl: you may consult `https://github.com/FreeRDP/FreeRDP` (Apache-2.0).
  Idiomatic Go ports are fine; PRs that lift code must say so in the description.
- License: GPL-3.0 (matches upstream grdp).
- Do NOT push to upstream `nakagami/grdp`. Only this fork.
- Do NOT modify `WORKFLOW.md` or any file under `.codex/` — those are workflow infra.
- Hard cap: $500 OpenAI spend (enforced at the LiteLLM proxy at OPENAI_BASE_URL).
- Hard cap: 24 hours wall-clock per the systemd unit.

## Workflow

When you start on an issue:

1. Read the issue description carefully — it contains acceptance criteria,
   scope, and links to MS-RDP* spec sections + xfreerdp source paths.
2. If the issue depends on other issues (`blockedBy`/`relatedTo`), check
   whether those have merged PRs; if not, comment on the issue and stop.
3. Implement the acceptance criteria in a feature branch. Commit early and often.
4. `go build ./...` and `go test ./...` must pass.
5. Open a PR against `master` of this fork. Add a clear description.
6. If CI fails, iterate.

Move the issue to `In Progress` on entry, leave a `## Codex Workpad` comment
with your plan and progress, and update it as you work. Move to `Done` only
when the PR is merged.

## Definition of done

- All acceptance criteria from the issue checked off.
- Tests added for the new code.
- `go build ./...` and `go test ./...` pass on CI.
- PR description is clear (what / why / any FreeRDP attribution).
- The bench scenario suite (`/opt/bench/scenarios/lib/score.sh`) doesn't
  regress on previously-passing scenarios.

## When stuck

If blocked by a true external blocker (missing tool/auth that a human must
resolve), comment on the issue and move it to `Cancelled`. Do not loop.
