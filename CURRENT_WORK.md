# MauriStocks — current work

**As of 2026-11-02.** Authoritative merged baseline: GitHub `main`. Active new integration branch: `feat/homepage-storage-sync`, [draft PR #3](https://github.com/richard-mauri/mauristocks/pull/3). Previous independent UI draft [PR #2](https://github.com/richard-mauri/mauristocks/pull/2) must not be overwritten or silently merged.

## Objective

Add explicit optional Homepage-account persistence for MauriStocks watchlists and user preferences without removing browser-only operation. The Homepage server owns authentication and per-user isolated data; MauriStocks has no token administration.

## Implemented in draft PR #3

- Added Load from account control beside watchlist controls. Save As New, Save Changes, Delete List and changes to refresh interval, visible columns (including Reset columns), sorting, flagged stars, symbol link destination and selected watchlist schedule account saves when signed in; signed-out use is browser-only. Removed redundant Save to account control.
- The app always writes browser `localStorage` first. After signed-in user preference/watchlist changes, a debounced save pushes the latest snapshot to server using revision checks; signed-out changes remain browser-only. Explicit Load from account remains available.
- At `https://stocks.themauris.org`, a read-only credentialed Homepage profile request supplies the username; credentialed GET to `https://home.themauris.org/api/storage/mauristocks` detects availability (200 record exists, 404 authenticated but empty, 401 signed out). The cross-origin `/api/session` endpoint is intentionally not used.
- First write to an existing account record asks for confirmation, subsequent watchlist edits use the server revision; HTTP 409 refuses concurrent overwrites. Unsupported stored data is not overwritten. Load validates a versioned data envelope, asks for confirmation, snapshots old local values and tries to restore them if applying new values throws.
- No raw token is accepted or stored in this app. Signed-out and localhost use existing browser settings, without server sync.
- App version remains 1.0.0; provider, quote, refresh and watchlist logic otherwise unchanged.

## Pending acceptance

Check latest CI at exact head (not yet verified); run `go test -count=1 ./...` and `bash ./devbuild.sh` on owner's Mac; inspect Safari UX at `https://stocks.themauris.org` across signed-out/user/admin states, including save, load, conflicting revisions, unsupported records, refresh, stars, columns, and local-only mode. Verify no errors in Web Inspector, and ensure server side persistence is backed up. The first draft is **not production accepted**.

Do not merge, release, restart either service, modify tunnels or DNS, or migrate data without explicit owner approval.
