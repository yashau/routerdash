package routerdash

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFakeHarnessCollectsWithoutSystemBinaries(t *testing.T) {
	collector := NewCollector(FakeRunner{}, func() time.Time {
		return time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	})
	ctx := context.Background()

	if summary := collector.Summary(ctx); summary.Hostname != "routerdash-lab" || !summary.WANIP.Available || summary.WANIP.Value != "203.0.113.42" {
		t.Fatalf("fake WAN IP was not collected: %+v", summary.WANIP)
	}
	if routes := collector.Routes(ctx); !routes.Available || !strings.Contains(routes.Output, "table 52") {
		t.Fatalf("fake routes were not collected: %+v", routes)
	}
	if rathole := collector.Rathole(ctx); !rathole.Available || !rathole.Active || len(rathole.Units) != 2 || !strings.Contains(rathole.Output, "ratholec@edge.service") {
		t.Fatalf("fake rathole units were not collected: %+v", rathole)
	}
	if tailscale := collector.Tailscale(ctx); !tailscale.Available || tailscale.Uptime != "4h 0m" || !contains(tailscale.AdvertisedRoutes, "192.168.88.0/24") {
		t.Fatalf("fake tailscale was not collected: %+v", tailscale)
	}
	if diag := collector.Diagnostic(ctx, DiagnosticRequest{Tool: "mtr", Target: "1.1.1.1"}); !diag.Available || !strings.Contains(diag.Output, "HOST") {
		t.Fatalf("fake mtr was not collected: %+v", diag)
	}
	if leases := collector.DHCPLeasesPage(ctx, PageRequest{Page: 1, PageSize: 10}); !leases.Available || len(leases.Leases) != 4 || leases.Leases[0].Hostname != "workstation" {
		t.Fatalf("fake DHCP leases were not collected: %+v", leases)
	}
}

func TestParseMeminfo(t *testing.T) {
	mem, err := parseMeminfo("MemTotal: 1000 kB\nMemAvailable: 250 kB\n")
	if err != nil {
		t.Fatal(err)
	}
	if mem.TotalBytes != 1024000 || mem.UsedBytes != 768000 {
		t.Fatalf("unexpected memory metrics: %+v", mem)
	}
}

func TestParsePingMS(t *testing.T) {
	ms, ok := parsePingMS("64 bytes from host: icmp_seq=1 ttl=57 time=12.8 ms\n")
	if !ok || ms != 13 {
		t.Fatalf("unexpected parsed latency: %d %v", ms, ok)
	}
}

func TestParseRatholeUnits(t *testing.T) {
	units := parseRatholeUnits("ratholec@edge.service loaded active running Rathole client edge tunnel\n")
	if len(units) != 1 || units[0].Name != "ratholec@edge.service" || units[0].Active != "active" || units[0].Sub != "running" {
		t.Fatalf("unexpected units: %+v", units)
	}
	if summary := ratholeStateSummary(units); summary != "active (running)" {
		t.Fatalf("unexpected summary: %s", summary)
	}
}

func TestParseTailscaleUsesPrimaryRoutesWhenAdvertisedRoutesMissing(t *testing.T) {
	status, err := parseTailscale(`{
		"BackendState": "Running",
		"Self": {
			"TailscaleIPs": ["100.64.0.6", "fd7a:115c:a1e0::6"],
			"AllowedIPs": ["10.10.31.0/24", "100.64.0.6/32", "fd7a:115c:a1e0::6/128"],
			"PrimaryRoutes": ["10.10.31.0/24"]
		},
		"Peer": {}
	}`, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !contains(status.AdvertisedRoutes, "10.10.31.0/24") {
		t.Fatalf("expected primary routes to be advertised routes: %+v", status)
	}
}

func TestParseSystemdTimestamp(t *testing.T) {
	started, ok := parseSystemdTimestamp("Thu 2026-06-25 08:00:00 +00")
	if !ok || started.IsZero() {
		t.Fatalf("expected systemd timestamp to parse: %v %v", started, ok)
	}
}

func TestRoutesPagePaginatesServerSide(t *testing.T) {
	collector := NewCollector(FakeRunner{}, time.Now)
	routes := collector.RoutesPage(context.Background(), PageRequest{Page: 2, PageSize: 2})
	if routes.Page == nil || routes.Page.Page != 2 || routes.Page.Total != 4 || routes.Page.TotalPages != 2 {
		t.Fatalf("unexpected route page metadata: %+v", routes.Page)
	}
	if len(routes.Lines) != 2 || strings.Contains(routes.Output, "default via") {
		t.Fatalf("unexpected paged route output: %+v", routes)
	}
}

func TestParseDHCPLeases(t *testing.T) {
	now := time.Unix(1782403200, 0)
	leases, err := parseDHCPLeases("1782406800 aa:bb:cc:dd:ee:ff 192.168.1.10 laptop 01:aa\n0 11:22:33:44:55:66 192.168.1.2 * *", now)
	if err != nil {
		t.Fatal(err)
	}
	if len(leases) != 2 {
		t.Fatalf("unexpected lease count: %d", len(leases))
	}
	if leases[0].Remaining != "1h 0m" || leases[0].Hostname != "laptop" || leases[0].ClientID != "01:aa" {
		t.Fatalf("unexpected parsed lease: %+v", leases[0])
	}
	if leases[1].ExpiresAt != "never" || leases[1].Hostname != "" || leases[1].ClientID != "" {
		t.Fatalf("unexpected static lease: %+v", leases[1])
	}
}

func TestDHCPLeasesPagePaginatesServerSide(t *testing.T) {
	collector := NewCollector(FakeRunner{}, func() time.Time {
		return time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	})
	leases := collector.DHCPLeasesPage(context.Background(), PageRequest{Page: 2, PageSize: 2})
	if leases.Page == nil || leases.Page.Page != 2 || leases.Page.Total != 4 || leases.Page.TotalPages != 2 {
		t.Fatalf("unexpected DHCP page metadata: %+v", leases.Page)
	}
	if len(leases.Leases) != 2 || leases.Leases[0].Hostname != "core-switch" {
		t.Fatalf("unexpected paged leases: %+v", leases.Leases)
	}
}

func TestTailscalePagePaginatesPeersServerSide(t *testing.T) {
	collector := NewCollector(FakeRunner{}, func() time.Time {
		return time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC)
	})
	status := collector.TailscalePage(context.Background(), PageRequest{Page: 1, PageSize: 1})
	if status.Page == nil || status.Page.Total != 2 || !status.Page.HasNext || len(status.Peers) != 1 {
		t.Fatalf("unexpected tailscale page: %+v", status)
	}
}

func TestDiagnosticRejectsShellMetacharacters(t *testing.T) {
	collector := NewCollector(FakeRunner{}, time.Now)
	result := collector.Diagnostic(context.Background(), DiagnosticRequest{Tool: "ping", Target: "host;reboot"})
	if result.Available {
		t.Fatalf("unsafe target should be rejected: %+v", result)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
