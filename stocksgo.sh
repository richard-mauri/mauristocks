#!/bin/bash

# start Mauri's Stock Quotes and its Cloudflare Tunnel
# ipconfig getifaddr en0

set -e

HOME=/Users/richardmauri
cd "$HOME/projects/mauristocks" || exit 1

# Stop only the existing stock application.
pkill -x mauristocks 2>/dev/null || true

# Load Finnhub credentials from a private environment file.
if [[ -f "$HOME/projects/mauristocks/.env" ]]; then
    set -a
    source "$HOME/projects/mauristocks/.env"
    set +a
fi

# BEGIN SUITE ACTIVITY ENV - managed by setup-suite-activity.sh
ACTIVITY_ENV="$HOME/projects/mauristocks/.activity.env"
if [[ ! -r "$ACTIVITY_ENV" ]]; then
  echo "Missing readable activity configuration: $ACTIVITY_ENV" >&2
  exit 1
fi
# Private file is generated with export-only assignments; never print it.
# shellcheck disable=SC1090
source "$ACTIVITY_ENV"
# END SUITE ACTIVITY ENV
# Stop only the STOCKS Cloudflare Tunnel.
STOCKS_TOKEN="$HOME/projects/mauristocks/.cloudflared/tunnel_token"

ps -axo pid=,command= | while read -r pid command; do
    if [[ "$command" == *"cloudflared tunnel run"* &&
          "$command" == *"$STOCKS_TOKEN"* ]]; then
        kill "$pid"
    fi
done

./mauristocks -version

# Start the stock application.
nohup "$HOME/projects/mauristocks/mauristocks" -host 0.0.0.0 -port 8082 \
    > "$HOME/projects/mauristocks/mauristocks.log" 2>&1 < /dev/null &

# Start only the STOCKS Cloudflare Tunnel.
nohup "$HOME/Applications/cloudflared" tunnel run \
    --token-file "$STOCKS_TOKEN" \
    > "$HOME/projects/mauristocks/.cloudflared/tunnel.log" 2>&1 < /dev/null &

disown -a

# Show running processes.
pgrep -fl 'mauristocks|cloudflared'

