# Agent Guide

This repo is `routerdash`: a single Go binary that runs as a systemd service on a Linux router and serves an embedded SvelteKit dashboard. The app is read-only except for bounded diagnostic commands.

## Non-Negotiables

- Use `mise` for project commands. Do not run raw `gofmt`, `go test`, `pnpm build`, etc. unless you are debugging a failing mise task.
- The main quality gate is `mise run check`.
- The formatter is `mise run format`.
- Keep files under 1000 LOC; `mise run check` enforces this.
- Keep line endings LF. `.gitattributes` enforces this.
- Assume the target router binaries exist, but local development must work without them through the fake harness.
- All external command execution in Go must have fixed timeouts. Do not add unbounded command calls.
- The web UI should load the page shell first and hydrate status/output later.
- Do not add the removed Network or Connectivity pages back. Important connectivity probes live on the dashboard.

## Toolchain

Configured in `.mise.toml`:

- Go `1.26`
- Node `26`
- pnpm `11`

Common commands:

```sh
mise run setup
mise run format
mise run check
mise run screenshots
mise run screenshots:dark
mise run screenshots:light
mise run build-local
mise run pack
mise run pack-arm64
mise run version
mise run version-bump
```

`mise run screenshots` starts the fake local binary and writes both light and dark screenshots to `docs/screenshots/`. Use `mise run screenshots:dark` or `mise run screenshots:light` for one color scheme. Filenames include the scheme, for example `dashboard-dark.png`.

## Versioning

The project version lives in `VERSION` and must match:

```text
YYYY.MM.DD-N
```

Builds stamp this version into the Go binary with ldflags and expose it at:

```text
GET /api/version
```

`mise run pack` creates:

- `dist/routerdash-linux-amd64.tar.gz`
- `dist/routerdash-${VERSION}-linux-amd64.tar.gz`

`mise run pack-arm64` creates the matching arm64 package names. GitHub Actions checks that the version does not already exist as a GitHub release before producing package artifacts.

## Architecture

Top-level entry:

- `main.go` embeds `web/build` and starts the HTTP server.

Backend:

- `internal/routerdash/server.go`: routes, JSON handlers, SPA fallback.
- `internal/routerdash/collect.go`: all router status collection and parsing.
- `internal/routerdash/runner.go`: command runner with fixed command timeouts.
- `internal/routerdash/fake.go`: fake command output for local tests/screenshots.
- `internal/routerdash/types.go`: shared API response models.

Frontend:

- `web/src/routes/+page.svelte`: main dashboard.
- `web/src/routes/dhcp/+page.svelte`: dnsmasq DHCP leases with server-side pagination.
- `web/src/routes/tailscale/+page.svelte`: Tailscale details with server-side peer pagination.
- `web/src/routes/rathole/+page.svelte`: rathole service/unit details.
- `web/src/routes/firewall/+page.svelte`: nftables or iptables output.
- `web/src/routes/routes/+page.svelte`: route table output with server-side pagination.
- `web/src/routes/frr/+page.svelte`: FRR OSPF/BGP/config output.
- `web/src/routes/diagnostics/+page.svelte`: bounded ping/MTR runner.
- `web/src/lib/components/Shell.svelte`: header, nav, title, version display.

Static frontend output is generated into `web/build` and embedded by Go.

## API Notes

Current API routes:

- `GET /healthz`
- `GET /api/version`
- `GET /api/summary`
- `GET /api/metrics`
- `GET /api/dhcp?page=1&pageSize=50`
- `GET /api/tailscale?page=1&pageSize=10`
- `GET /api/rathole`
- `GET /api/firewall`
- `GET /api/routes?page=1&pageSize=50`
- `GET /api/frr`
- `POST /api/diagnostics`

Unknown `/api/*` routes should return 404, not the SPA shell.

The dashboard uses `/api/summary` and `/api/metrics`. Network interface cards at the bottom include IPs from summary LAN data; do not add a separate LAN IP card.

## Router Command Behavior

Command calls go through `Runner`. `ExecRunner` has a default timeout of 8 seconds. Diagnostics use larger per-tool bounded timeouts:

- ping: 12 seconds
- mtr: 35 seconds
- diagnostics HTTP handler: 45 seconds

MTR must run as:

```sh
mtr -r -b -w -c10 <target>
```

Rathole status discovers common systemd unit names:

- `ratholec@*.service`
- `rathole@*.service`
- `rathole.service`

Tailscale parsing supports current `tailscale status --json` shapes:

- self IPs from `Self.TailscaleIPs`
- configured advertised routes from `tailscale debug prefs` field `AdvertiseRoutes`
- approved advertised routes from `Self.PrimaryRoutes`
- route acceptance inferred from `AllowedIPs` count vs self IP count
- uptime from `Self.Started` when present, otherwise from `systemctl show tailscaled.service --property=ActiveEnterTimestamp --value`
- received routes per peer from `Peer.*.PrimaryRoutes`

Routes output comes from:

```sh
ip route show table all
```

Routes are paginated server-side so the browser does not receive huge routing tables.

DHCP leases are read from `ROUTERDASH_DHCP_LEASES_FILE` when set, then common dnsmasq lease paths:

- `/tmp/dhcp.leases`
- `/var/lib/misc/dnsmasq.leases`
- `/var/lib/dnsmasq/dnsmasq.leases`

DHCP leases are paginated server-side so the browser does not receive huge lease files.

## Frontend Conventions

- SvelteKit is under `web/` and uses pnpm.
- Formatting/linting uses oxlint, oxlint-tailwindcss, and oxfmt.
- shadcn-svelte components live under `web/src/lib/components/ui/`; do not hand-roll replacements when a registry component exists.
- Use lucide icons when possible.
- The app supports automatic light/dark mode based on system theme.
- Avoid horizontal scrolling. Long route/IP text should wrap with `whitespace-normal`, `break-words`, and `[overflow-wrap:anywhere]` as needed.
- Header hostname is trimmed in both the header and the browser title.
- The nav intentionally contains only:
  - Dashboard
  - DHCP
  - Tailscale
  - Firewall label dynamically shown as `nftables` or `iptables`
  - Routes
  - FRR
  - Diagnostics

## Fake Harness

Set:

```sh
ROUTERDASH_FAKE=1
```

The fake runner provides deterministic outputs for command-dependent collectors. Use this for local builds, screenshots, and tests on machines without router binaries.

Local preview binary:

```sh
mise run build-local
$env:ROUTERDASH_FAKE="1"; $env:ROUTERDASH_ADDR="127.0.0.1:18082"; .\dist\routerdash.exe
```

On Linux/macOS, use the matching shell syntax and `./dist/routerdash`.

## Deployment

Systemd unit:

- `deploy/routerdash.service`
- binary path: `/usr/local/bin/routerdash`
- default listen address: `:8080`

Package files include:

- `routerdash`
- `routerdash.service`
- `VERSION`

The known router target used during development is:

```text
root@10.10.31.1
```

It is x86_64 and uses systemd. Deploy by building `mise run pack`, copying `dist/routerdash-linux-amd64.tar.gz`, extracting it, installing the binary and unit, then restarting `routerdash.service`.

## CI

GitHub Actions workflow:

- `.github/workflows/ci.yml`

It:

- validates `VERSION`
- checks GitHub releases for `VERSION` and `vVERSION`
- runs `mise run check`
- packages amd64 and arm64 only when the version is not already released
- verifies the stamped binary reports the expected `/api/version`
- uploads versioned tarballs as workflow artifacts

## Before Finishing Work

For code changes, run:

```sh
mise run format
mise run check
```

For UI changes, also run:

```sh
mise run screenshots
mise run screenshots:dark
mise run screenshots:light
```

When deployment is requested, run:

```sh
mise run pack
```

Then deploy and verify:

- `systemctl is-active routerdash.service`
- `GET /api/version`
- the relevant changed API/page
