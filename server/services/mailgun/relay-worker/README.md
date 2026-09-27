# Mailgun relay (Cloudflare Worker)

Mailgun's API (`api.mailgun.net`) has no IPv6 address, so a server with no
outbound IPv4 connectivity can't reach it directly. This is a tiny relay,
deployed on Cloudflare's free tier (Cloudflare's edge has normal IPv4 egress),
that the server talks to instead - it just forwards the request to Mailgun
and returns the response.

It's gated by a shared secret so it can't be used as an open proxy by anyone
who finds the deployed URL.

## Deploy

You need a free Cloudflare account (cloudflare.com - no credit card required
for the free tier used here).

**Option A - CLI (`wrangler`):**

```bash
cd server/services/mailgun/relay-worker
npx wrangler login          # opens a browser to authorize the CLI
npx wrangler deploy
npx wrangler secret put RELAY_SECRET
# paste a random value when prompted, e.g. from: openssl rand -hex 32
```

The `deploy` output prints the Worker's URL, something like:
`https://mailgun-relay.<your-subdomain>.workers.dev`

**Option B - dashboard (no CLI/install needed):**

1. Cloudflare dashboard -> Workers & Pages -> Create -> Create Worker.
2. Replace the default script with the contents of `worker.js`.
3. Deploy.
4. Settings -> Variables and Secrets -> add a secret named `RELAY_SECRET`
   with a random value (e.g. from: `openssl rand -hex 32`).
5. Note the Worker's URL, shown on its overview page.

## Point the server at it

Add to the server's `.env` (see `server/.env.template`):

```
MAILGUN_API_BASE_URL=https://mailgun-relay.<your-subdomain>.workers.dev
MAILGUN_RELAY_SECRET=<the same random value you set as RELAY_SECRET>
```

Restart the server so it picks up the new `.env` values. Leaving both of
these unset (the default) sends directly to Mailgun as before - only set
them if the server itself can't reach Mailgun's API directly.
