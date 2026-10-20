# MauriStocks — AI Development Process and Chat Handoff

**This file is for coding agents and future development chats.** `README.md` is the end-user guide. Source basis: GitHub `richard-mauri/mauristocks` README and `main.go` consulted October 20, 2026; the user's current local files and deployment have **not** been inspected.

## Identity and architecture

- Repository: https://github.com/richard-mauri/mauristocks
- Application: Mauri's Stock Quotes, a compact self-hosted Go dashboard.
- Local directory documented in source README: `~/mauristocks` (confirm actual path before use); default `http://localhost:8082`.
- Main executable source: `main.go`; `go.mod`; Go **standard library only**.
- Data provider: **unofficial Yahoo Finance chart/screener** endpoints, subject to outages, throttling and provider terms. Optional server-side `FINNHUB_API_KEY` for company websites.
- Browser persists user watchlists, flags, sort, link destination and refresh choices in `localStorage` (browser/origin-specific, not server-side).
- Default list of 20 symbols, named editable watchlists and an upstream-dependent dynamic Most Active Top 20 list.
- Local flag and default bind: `127.0.0.1:8082`; a hosted `PORT` may change default bind to `0.0.0.0`. Public tunnel arrangements must be verified before changing them.

## Current development handoff — October 20, 2026

The GitHub README describes **semantic version 1.0.0**, displays **Build v38** in its opening while also documenting a newer **Build v39 sort indicator**. **Do not assume the header identifies the latest running code or next build.** Check actual `main.go` version/build constants, local source, and browser version first. There is no user-confirmed local Git state in this handoff.

Current features documented include premarket/after-hours column with session-specific headings, extended price changes relative to regular Price, price momentum coloring, persistent starred-symbol row highlighting, sorting, year/day ranges, named watchlists, a collapsible About/Data Notes area, and optional corporate URL discovery.

**Known documentation inconsistency:** README's v38 headline and v39 feature entry conflict; reconcile against current source before assigning a next build number or announcing a release. Do not rewrite past versions as verified without evidence.

## Engineering and regression rules

- Preserve the distinction between regular **Price**, premarket/after-hours price and the two **Change** columns. Extended-session changes compare to regular Price; regular-session changes compare to prior close. Avoid mixing timestamp sources or claiming delayed one-minute bars are live trades.
- The extended-hours column is hidden completely (heading + data) during regular session; sort fallback moves to descending Price if the hidden column had been selected.
- Preserve star/flag persistence and color distinction from independent price-momentum text colors; do not accidentally reset flags during resort or refresh.
- Preserve named list management and the read-only Default and Most Active lists. Do not expose optional Finnhub API keys in browser source, logs or Git.
- Preserve numeric sorting, range-marker semantics, Pacific-time display and no-overlapping-refresh behavior. Confirm API JSON contracts (`/api/quotes`, `/api/most-active`, `/api/company`, `/company`, `/healthz`).
- Test safely with recorded/fixture data when necessary; do not make passing tests depend on Yahoo/Finnhub availability or guarantee upstream results.
- Base code modifications on current `main.go`. When delivering replacements, include all affected files and both updated documents; retain existing executable/build workflow, rather than imposing ScraperGo's ZIP requirement.

## Local build and verification

The documented local workflow is:

```sh
cd ~/mauristocks
go test ./...
go build -o mauristocks .
./mauristocks
```

If the project path or existing launch procedure differs, confirm it. `go run main.go` is an alternative documented source-only run. Check browser at `http://localhost:8082`, `/healthz`, manual refresh, different session display states (with deterministic test data where appropriate), sorting, watchlists and flagged rows. Do not claim the build or live feed was verified unless actually tested.

## Non-negotiable agent conduct

- Read this file and `README.md` in full before changing code, advising Git operations, or claiming completion.
- The user's verified **local source tree** is authoritative whenever it differs from GitHub. Obtain current affected files and reconcile local changes; never overwrite them from older remote sources.
- Trace **both definitions and actual call sites**, handlers, UI uses, tests, and configuration dependencies. A new function that is not called is **not** a finished fix.
- Protect unrelated working behavior. Keep changes minimal and test existing regression areas.
- Distinguish **drafted/generated**, **verified file exists**, **formatted**, **tested in full project**, **tested in isolated harness**, **installed**, **user-verified live**, **committed**, **pushed**, **tagged**, and **released**. Never imply one means another.
- If you cannot run tests or produce a real artifact, say so. Never invent a download link, test result, deployment, or Git action.
- Do not commit, push, tag, release, restart services, deploy, or modify tunnel configuration without the user's explicit authorization.

## Mandatory Git pre-commit audit — learned from ScraperGo failure

**Never compose a `git add` filename list from memory, an earlier ZIP manifest, or an assumption about changed files.** A previous agent omitted a crucial `readability.go` integration change in ScraperGo; the user found it before pushing. This is a cross-project safeguard.

For a commit request, **separately from normal patch installation**:

1. Obtain fresh `git status`, including staged, unstaged and untracked files, and ahead/behind state.
2. Examine `git diff --check`, `git diff`, `git diff --cached`, and relevant untracked files. If commits already exist locally, review `git log origin/main..HEAD` and commit file inventories.
3. Check all changed source files **and integration points**, tests, both documentation files, and build/configuration artifacts as applicable. Exclude secrets, local tokens, binaries, logs and unrelated temporary files.
4. Only after that evidence recommend exact staging commands. Before committing, verify `git diff --cached --name-status` and `git diff --cached` against the intended scope.
5. Do not infer that a clean working tree means the correct commits were made, or that ahead-of-origin commits were pushed. Verify publication independently.

## Cross-chat handoff obligations

At the start of a new chat, use `AGENTS.md` for the **current engineering handoff** and `README.md` for the public-facing project context. Confirm current local source and Git state before applying edits. Maintain this file with a short, dated snapshot of the most recent verified baseline, pending work, confirmed tests, known regressions and next action; **replace obsolete current-state claims**, rather than accumulating contradictory updates. Keep the README focused on users and general architecture, not chat history or agent rules. Preserve the existing project's patch/install convention; do not transplant another project's conventions blindly.

