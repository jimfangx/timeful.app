// Cloudflare Worker that relays requests to Mailgun's API.
//
// Why this exists: Mailgun's API (api.mailgun.net) has no IPv6 address, so a
// host with no outbound IPv4 connectivity can't reach it directly. Cloudflare's
// edge has normal IPv4 egress, so this Worker sits in between: the server
// sends its Mailgun API request here instead, and this forwards it on.
//
// This only ever forwards to Mailgun's own API host (hardcoded below, not
// taken from the request), and requires a shared secret header so it can't
// be used as an open proxy by anyone who finds the Worker's URL.
//
// Deploy with `wrangler deploy`, then set the shared secret:
//   wrangler secret put RELAY_SECRET
// (or add it under the Worker's Settings > Variables and Secrets in the
// Cloudflare dashboard). Point the server at this Worker's URL via
// MAILGUN_API_BASE_URL, and set the same value as MAILGUN_RELAY_SECRET in
// the server's own .env - see server/.env.template.

const MAILGUN_API_HOST = "https://api.mailgun.net"

export default {
  async fetch(request, env) {
    const secret = request.headers.get("X-Relay-Secret")
    if (!env.RELAY_SECRET || secret !== env.RELAY_SECRET) {
      return new Response("Forbidden", { status: 403 })
    }

    const url = new URL(request.url)
    const target = MAILGUN_API_HOST + url.pathname + url.search

    const headers = new Headers(request.headers)
    headers.delete("x-relay-secret")
    headers.delete("host")

    const response = await fetch(target, {
      method: request.method,
      headers,
      body: request.body,
      duplex: "half",
    })

    return new Response(response.body, {
      status: response.status,
      headers: response.headers,
    })
  },
}
