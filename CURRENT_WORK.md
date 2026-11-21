# Current work — MauriStocks

**Verified baseline:** GitHub merged `main` at `f1599398a5ddbc563360a604b6629e2fd5b5ca37`, PR #34 merged and Issue #33 closed. New [Issue #35](https://github.com/richard-mauri/mauristocks/issues/35), scoped `auth/35-required-suite-access` branch from main, [draft PR #36](https://github.com/richard-mauri/mauristocks/pull/36) targeting `main` (verify exact head and CI on GitHub).

**Active scope:** Implement mandatory verified Homepage personal or guest access. The previous implementation exposed anonymous quote/market/company APIs and allowed direct Safari Private Browser to land in `suite_mode=local`. This branch adds the pinned Homepage-owned `shared/auth` module, server-side per-request Homepage bearer verification guarding quote, market and company endpoints, an initial inert browser shell with hidden stock content and bearer-authenticated data fetches after trusted profile verification, authenticated corporate-site click lookup in place of an unauthenticated `/company` navigation, and immediate UI hiding/redirection on a revoked API bearer, and redirects rejected legacy local mode to Homepage sign-in/guest gate. Guest can use the app but never the personal Homepage storage endpoint; personal Save Changes remains explicit and success status is simplified. No consumer-independent authentication authority is added.

**Integration dependency:** Homepage [Issue #60](https://github.com/richard-mauri/homepage/issues/60), draft [PR #61](https://github.com/richard-mauri/homepage/pull/61), introduces shared Go auth and a fail-closed recovery bridge. It remains unmerged; DO NOT deploy this MauriStocks candidate without coordinated Homepage integration/owner acceptance. Homepage separate CORS PR #59 is not part of this candidate.

**Data safety:** No owner Mac-local checkout, live accounts, user preferences/watchlists, credentials/tokens, browser state, persistent records, private build counters, services or Cloudflare tunnels accessed or modified. Local dirty/unpushed state inaccessible. No schema, pricing-provider, server session lifetime or migration changes.

**Docs:** Read merged Homepage policy v1.1.0 and MauriStocks AGENTS, README, ARCHITECTURE, DEVELOPMENT, CURRENT_WORK. Updated README, ARCHITECTURE, DEVELOPMENT and this handoff. AGENTS and shared policy unchanged.

**Verification:** Go module pin and tests/CI require exact-head checks; server tests exercise missing/spoofed/guest/personal/expired/outage, client static checks require hidden bootstrap and guarded fetch. Mac/Safari/live Cloudflare and real per-user settings acceptance unverified. Confirm CI result rather than relying on prior MauriStocks PR #34 CI.

**Owner handoff after CI:** `cd ~/projects/mauristocks`; `git status -sb` STOP if dirty; `git fetch origin`; `git checkout auth/35-required-suite-access`; `git pull --ff-only origin auth/35-required-suite-access`; `./devbuild.sh`; `./mauristocks -version`. Only separately authorized `./stocksgo.sh` restarts the dedicated application and tunnel. Deployment order/coordination with Homepage PR #61 is mandatory.

**Promotion:** Draft only. No FINISH, merge, Issue closure, branch deletion, restart, deployment, tag, release or version/builder-counter mutation authorized.
