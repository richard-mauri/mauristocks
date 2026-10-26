# MauriStocks development and validation

GitHub [main](https://github.com/richard-mauri/mauristocks/tree/main) is the authoritative merged baseline. Read [AGENTS.md](AGENTS.md), [CURRENT_WORK.md](CURRENT_WORK.md), and [ARCHITECTURE.md](ARCHITECTURE.md) before modifying source.

## Branch / review lifecycle

Start a short-lived scope-specific branch from the latest main. Use a draft pull request targeting main, with rationale, risks, exact changed files and CI/acceptance status. Track reproducible independent bugs with GitHub Issues; keep unrelated fixes in separate PRs. Never commit experiments directly to main. Only the owner approves merge, tag, release, deployment, or restart. Do not bump the semantic version during ordinary development. After an owner-approved merge, verify merge status before deleting a branch; preserve unmerged work.

## CI and local validation

[Go CI](.github/workflows/go-ci.yml) follows the ScraperGo workflow: pushes to `dev`, PRs targeting `main`, and manual dispatch; non-blocking gofmt report, then `go test -count=1 ./...` and `go build ./...`. Check the latest PR-head SHA and its actual workflow outcome. Formatting debt output is informational; tests and build are required gates. For documentation-only changes review Markdown links/path/commands plus PR CI; Mac restart/Safari testing is unnecessary.

For code/UI changes, on the owner's Mac inspect local changes first and use:

```sh
cd ~/projects/mauristocks
git status -sb
git fetch origin
git checkout ui/content-fit-columns-and-docs
git pull --ff-only origin ui/content-fit-columns-and-docs
bash ./devbuild.sh
./mauristocks -version
./stocksgo.sh
```

`devbuild.sh` runs tests, increments a Git-private local build counter and creates the native binary with one identifier shared by the CLI and dashboard. It starts at 000041 for a fresh clone after historical v40, but independently existing checkout counters are retained. CI checks the injected build label without modifying this private counter. `stocksgo.sh` is a machine-local launcher **not** included in GitHub; inspect its behavior and stop the previous instance using the owner's usual procedure. Do not overwrite local modifications or assume the launcher uses the newly built binary. In Safari at http://127.0.0.1:8082 verify compact automatic content widths, long names, range labels, hidden/shown columns, table overflow, regular/extended session transitions, sorting, stars, watchlists, refresh and light/dark presentation. GitHub's Linux CI cannot establish Mac/Safari acceptance. Report pass/fail/not-run accurately.

Do not force-push shared work or expose tokens, and do not silently modify any other local service.