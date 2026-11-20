# MauriStocks — app-specific development and validation

Read [AGENTS.md](AGENTS.md) and [Homepage's authoritative suite development policy](https://github.com/richard-mauri/homepage/blob/main/SUITE_DEVELOPMENT_POLICY.md). Homepage owns shared workflow, checklist, FINISH, promotion and branch cleanup. This file retains app-specific CI and Mac/Safari acceptance.

## CI and local validation

For Issue #31, compare the stock masthead against Homepage, NewsGo and ScraperGo at identical desktop, tablet and mobile viewport widths. The shared banner owns all horizontal dimensions; Stock content below uses `.stocks-content` independently. Confirm Tanner is served by `/tanner.jpg`, shared CSS is inline in the no-store dashboard (no Safari stale CSS URL), and `suite-auth-status` follows verified Homepage identity or browser-only state. Confirm the original two blurred Homepage ambient shapes without the previous radial approximation. Keep quote sorting, columns, heatmap, watchlists, dialogs and red/green financial data semantics unchanged.

For Issue #33, verify a deliberately divergent browser/account pair in Safari: remove `DRAM` from a named watchlist, click Save Changes, and observe **Saved in this browser** followed by a separate **Settings saved to your account** result or an explicit declined/failed/browser-only status. A local success alone is not a remote success. In a browser profile with the older account copy, inspect the vertically stacked comparison: `DRAM` should be identified as absent/present in the appropriate side, and the legacy selected-symbol mirror should not repeat the Candidates change. Confirm no horizontal overflow at desktop/mobile widths; scroll many changed settings vertically. Exercise Keep Browser Settings, Restore Account Settings, Escape and keyboard Enter/Tab. Verify no account overwrite on Keep; Restore replaces the browser snapshot and reloads. Repeat with a missing field and malformed legacy watchlist serialization; ensure displayed saved values remain plain text. Confirm 409 revision-conflict behavior is unchanged.

For the latest Issue #33 Safari review, verify a short-height Safari window with multiple differences: the dialog must remain within the viewport; only the differences scroll, and Keep Browser Settings / Restore Account Settings remain visible and keyboard accessible at the bottom. When browser Default and account Candidates are selected and each stored symbols mirror matches its respective selected list, the dialogue must show Selected watchlist but omit misleading Last saved symbols inventories. When Candidates differs by only DRAM, show **Added in this browser (1): DRAM** (or **Removed from this browser (1): DRAM** in the opposite direction); never show the entire watchlist twice. Run `node scripts/test-account-diff.cjs`, which is included in Go CI.

For Issue #33's follow-up, compare a stored `{\"hidden\":[\"exchange\",\"name\"]}` browser preference with an account storing `{\"hidden\":[\"exchange\"]}`. The dialog must display `Name: hidden` versus `Name: shown`, not `hidden: shown`. After keeping browser settings, altering Auto refresh or Columns should show a pending Homepage save followed by persistent `Settings saved to your account.` after a successful PUT; focusing/revalidating the session must not erase it. Confirm HTTP 200 PUT in Safari Network without exposing authorization tokens. Test failed PUT and browser-only status, and confirm Save Changes still applies only to edited watchlist symbols. Run `node scripts/test-account-diff.cjs` as a focused local regression; GitHub CI runs it too.

[Go CI](.github/workflows/go-ci.yml) follows the ScraperGo workflow: pushes to `dev`, PRs targeting `main`, and manual dispatch; non-blocking gofmt report, then `go test -count=1 ./...`, Node account-comparison regression checks and `go build ./...`. Check the latest PR-head SHA and its actual workflow outcome. Formatting debt output is informational; tests and build are required gates. For documentation-only changes review Markdown links/path/commands plus PR CI; Mac restart/Safari testing is unnecessary.

For code/UI changes, first inspect the owner's local tree and stop if dirty. Use the following with the **actual scoped PR branch**, not the historical example. `./devbuild.sh` is the only required build/test command; do not separately request `go test` or `go build`:

```sh
cd ~/projects/mauristocks
git status -sb
git fetch origin
git checkout ACTUAL-PR-BRANCH
git pull --ff-only origin ACTUAL-PR-BRANCH
./devbuild.sh
./mauristocks -version
```

Only after separate owner authorization, run `./stocksgo.sh` to restart the app and tunnel for Safari acceptance.

`devbuild.sh` runs tests, increments a Git-private local build counter and creates the native binary with one identifier shared by the CLI and dashboard. It starts at 000041 for a fresh clone after historical v40, but independently existing checkout counters are retained. CI checks the injected build label without modifying this private counter. `stocksgo.sh` is tracked in GitHub and is the owner's Mac launcher; inspect its current version and any local differences before relying on it. It restarts MauriStocks and its dedicated tunnel, so run it only after owner authorization. Do not overwrite local modifications or assume the launcher uses the newly built binary. In Safari at http://127.0.0.1:8082 verify compact automatic content widths, long names, range labels, hidden/shown columns, table overflow, regular/extended session transitions, sorting, stars, watchlists, refresh and light/dark presentation. GitHub's Linux CI cannot establish Mac/Safari acceptance. Report pass/fail/not-run accurately.

Do not force-push shared work or expose tokens, and do not silently modify any other local service.

## Correct watchlist identity and two-column review (Issue #33)

Safari acceptance must verify: browser selected Default versus account selected Candidates shows a **Selected watchlist** card but no `Last saved symbols` card; Default is never treated as an editable personal list. A Candidates list with browser DRAM and account TSLA (all other symbols equal) shows a Candidates card with `Added here: DRAM` / `Missing here: TSLA` on the browser side and the reverse on the account side. A list present only in one snapshot shows `Present (N symbols)` versus `Not present`, not an inferred deletion. Visible Columns uses the same **This browser** left and **Saved account** right layout. The footer buttons remain visible in a short Safari window while only the differences scroll. A bytewise-only legacy selected-symbol mirror or JSON key-order difference must not prompt for account replacement. Use `node scripts/test-account-diff.cjs` for the isolated JavaScript regression, which is also run in Go CI. Confirm Safari HTTP 200 authenticated PUT separately; no credentials should appear in test reports.

## Global comparison header Safari acceptance (Issue #33)

On a desktop-sized Safari window, a review containing Selected watchlist, Auto refresh and Visible Columns must display **This browser** and **Saved account** exactly once in a shared heading row, aligned over the respective left/right values. Scrolling the differences should leave this header and both decision buttons visible. At a width below 550px, the shared header must be hidden and each setting must stack values with its own visible **This browser** and **Saved account** labels; neither column should become ambiguous. Ordinary confirmation and New Watchlist dialogs must not display the comparison headings. Run the existing `./devbuild.sh` once for Go tests and the native build; CI additionally runs `node scripts/test-account-diff.cjs` for these contracts. The app/tunnel launcher remains a separate authorized action.
