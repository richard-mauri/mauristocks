# CURRENT WORK — shared UI palette and development workflow (2026-10-23)

This section is authoritative over historical handoff statements below. GitHub `main` is the repository's authoritative accepted baseline. Develop each scoped change on a short-lived branch created from the latest `main`, with a **draft pull request to main**. Do not commit development experiments straight to main. The owner alone authorizes merge, tag, release, deployment, or service restart. The current local checkout is `~/projects/mauristocks` as reported by the owner; inspect `git status -sb`, `git log -3 --oneline`, current branch, upstream and diff before modifying local files. Never overwrite uncommitted local work.

## Working sequence

1. Read `AGENTS.md` and `README.md` fully, then examine the actual implementation, tests, call sites and affected scripts. Confirm the branch starts from current GitHub main; confirm local checkout state separately.
2. Make one cohesive, reviewable change on a task branch. Keep quote APIs, pricing calculations, browser localStorage keys, flags, sorting, watchlists, refresh behavior, endpoints, launch configuration and secrets unchanged unless explicitly in scope.
3. Run `gofmt` on modified Go sources and `go test -count=1 ./...` and `go build ./...` (or the repository's verified build command). Check changes with `git diff --check`, review actual staged inventory and exclude generated executables and user data. Report unavailable CI/test access as *unverified*.
4. Open or update a **draft PR targeting main**, report exact commit and checked evidence. Do not confuse successful compilation with manual Mac/Safari acceptance.
5. After successful tests, obtain owner Mac/Safari acceptance of visual behavior, stock data rendering and persistence. The owner must explicitly authorize merge. Do not force-push shared history, tag, release or restart automatically.
6. Following an owner-approved merge, update local source with `git checkout main` and `git pull --ff-only origin main`, and use the project's confirmed build/launch workflow. Do not assume a GitHub-only change is already installed locally.

## Current UI task and invariants

Match the semantic light/dark design language of ScraperGo (`internal/ui/static/reader.css`) and NewsGo: warm near-white light background, blue-black dark background, white/slate panels, teal accents, muted secondary text, understated borders, and accessible input/button focus. Centralized CSS tokens should be the source of truth. **Keep green/red financial gains and losses semantically distinct from theme accents**, and preserve gold starred-row highlighting, sorted columns, compact table layout, range bars, session-dependent columns, mobile horizontal scrolling, keyboard access and readable contrast in both modes. No backend/API changes for a theme-only patch.

## Shared UI specification — source references and implementation rules

Before UI work, inspect **current** `richard-mauri/scrapergo/internal/ui/static/reader.css` and `richard-mauri/newsgo/internal/dashboard/dashboard.go` (embedded stylesheet). Both currently use `color-scheme: light dark` with `light-dark()`, and shared palette tokens: background `#f5f5f0/#0e141a`, panel `#fff/#19232b`, primary text `#182631/#eef2f4`, muted `#60707b/#a5b3bb`, border `#dce2e1/#34434a`, accent `#126877/#69c6d3`, soft `#e6f3f3/#203c43` (light/dark respectively).

Match their zero-margin, fully painted page background; bounded, padded containers; theme-driven surface, focus and hover colors; accessible keyboard controls; and responsive scrolling rather than page-level overflow. NewsGo's `main` content container and ScraperGo's `.shell` are references, not a reason to copy their article/card structure into a compact stock dashboard.

For MauriStocks tables, the legacy HTML `<colgroup>` with fixed inline pixel widths was removed after Safari screenshots showed a blank right strip and overlapping Change headings. The visible header cells now own the column percentages. Do not reintroduce competing column-width rules. The regular session suppresses the extended column. A table `width:100%` **alone does not guarantee that the visible columns fill the table**; inspect column constraints as well. In regular mode, displayed columns' allocated percentages must total 100%, excluding the hidden extended column. In extended mode, all nine column percentages must total 100%. Keep stable numeric alignment, range-marker labels, minimum table widths, and narrow-screen horizontal scrolling. Verify in Safari at normal and narrow widths, during both regular and extended market-session states, using deterministic data if necessary.

The Name column is now measured from the rendered watchlist names using canvas text metrics, clamped by `COMPANY_NAME_MIN_PX` and `COMPANY_NAME_MAX_PX` in `main.go`. Keep those adjustable limits documented and ensure recomputation on watchlist/quote refresh. Its width is exposed through `--company-name-width`; long labels ellipsize.

Avoid layering unexamined CSS patches indefinitely: consolidate superseded declarations when refactoring is in scope, and validate the browser's computed geometry rather than inferring behavior from source alone.

## Column controls candidate (2026-10-23)

The table now offers a Columns dropdown with per-column visibility, a reset control, pointer-draggable header separators and keyboard Left/Right resizing. Browser preferences are stored in `mauristocks_columns_v1` localStorage. Symbol stays visible; the existing regular-session suppression of extended price takes precedence. Day and 52-week range columns start with the **same** 205px base width; free remaining table width is divided equally between visible range columns, with horizontal scroll when needed. Verify browser behavior including drag, keyboard resize, hidden-column sorting, manual refresh, reset, reload persistence, mobile overflow, Safari and both session states before approval. This is a candidate, not verified by CI or Mac acceptance.

## Status and acceptance as of 2026-10-23

GitHub main was inspected, including `AGENTS.md`, `README.md`, and `main.go`: source reports appVersion 1.0.0 and buildVersion v39. The theme change is a development candidate, not an accepted release. GitHub connector edits alone cannot establish a passing Go build, browser layout validation, or local checkout status. Before merge, run Go checks and validate light/dark appearance, gain/loss text, flagged rows, sorting, refresh and watchlist persistence in Safari on the owner's Mac. Update this section after verified acceptance, removing superseded state rather than appending contradictory handoffs.

---

# MauriStocks — AI Development Process and Chat Handoff

**This file is for coding agents and future development chats.** `README.md` is the end-user guide. Source basis: GitHub `richard-mauri/mauristocks` README and `main.go` consulted October 20, 2026; the user's current local files and deployment have **not** been inspected.

## Identity and architecture

- Repository: https://github.com/richard-mauri/mauristocks
- Application: Mauri's Stock Quotes, a compact self-hosted Go dashboard.
- Owner's stated local working directory: `~/projects/mauristocks` (inspect actual checkout before use); default `http://localhost:8082`.
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
- GitHub accepted `main` is the authoritative baseline; reconcile the user's **local source tree** and any uncommitted changes before editing and never discard them. Obtain current affected files and reconcile local changes; never overwrite them from older remote sources.
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

