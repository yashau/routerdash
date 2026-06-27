package routerdash

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Collector struct {
	runner Runner
	now    func() time.Time
	fake   bool
	mu     sync.Mutex
	prev   *metricSnapshot
	fakeIx int
}

type metricSnapshot struct {
	at         time.Time
	cpuIdle    uint64
	cpuTotal   uint64
	interfaces map[string]InterfaceIO
}

func NewCollector(runner Runner, now func() time.Time) *Collector {
	_, fake := runner.(FakeRunner)
	return &Collector{runner: runner, now: now, fake: fake}
}

func (c *Collector) Summary(ctx context.Context) Summary {
	return Summary{
		Hostname:     c.Hostname(),
		Uptime:       c.Uptime(ctx),
		WANIP:        c.WANIP(ctx),
		LAN:          c.LAN(ctx),
		Connectivity: c.Connectivity(ctx),
		Tailscale:    c.Tailscale(ctx),
		Rathole:      c.Rathole(ctx),
	}
}

func (c *Collector) Hostname() string {
	if c.fake {
		return "routerdash-lab"
	}
	name, err := os.Hostname()
	if err != nil || strings.TrimSpace(name) == "" {
		return "router"
	}
	return strings.TrimSpace(name)
}

func (c *Collector) Uptime(_ context.Context) AvailabilityValue {
	if c.fake {
		return AvailabilityValue{Availability: Availability{Available: true}, Value: "3d 4h 12m"}
	}
	raw, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return AvailabilityValue{Availability: errUnavailable("uptime", err)}
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return unavailableValue("uptime output was empty")
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return unavailableValue(err.Error())
	}
	return AvailabilityValue{Availability: Availability{Available: true}, Value: humanDuration(time.Duration(seconds) * time.Second)}
}

func (c *Collector) WANIP(ctx context.Context) AvailabilityValue {
	out, err := c.runner.Run(ctx, "curl", "-fsS", "--max-time", "3", "https://api.ipify.org")
	if err != nil {
		return AvailabilityValue{Availability: errUnavailable("curl", err)}
	}
	return AvailabilityValue{Availability: Availability{Available: true}, Value: strings.TrimSpace(out)}
}

func (c *Collector) Metrics(_ context.Context) Metrics {
	if c.fake {
		return c.fakeMetrics()
	}
	stat, mem, netdev, err := readProcMetrics()
	if err != nil {
		return Metrics{Availability: errUnavailable("procfs", err)}
	}
	now := c.now()
	snap := metricSnapshot{at: now, cpuIdle: stat.idle, cpuTotal: stat.total, interfaces: netdev}

	c.mu.Lock()
	defer c.mu.Unlock()

	var cpuPct float64
	if c.prev != nil && stat.total > c.prev.cpuTotal {
		totalDelta := stat.total - c.prev.cpuTotal
		idleDelta := stat.idle - c.prev.cpuIdle
		cpuPct = 100 * float64(totalDelta-idleDelta) / float64(totalDelta)
	}

	interfaces := make([]InterfaceIO, 0, len(netdev))
	for name, current := range netdev {
		if c.prev != nil {
			if old, ok := c.prev.interfaces[name]; ok {
				elapsed := now.Sub(c.prev.at).Seconds()
				if elapsed > 0 {
					current.RxBps = float64(current.RxBytes-old.RxBytes) / elapsed
					current.TxBps = float64(current.TxBytes-old.TxBytes) / elapsed
				}
			}
		}
		interfaces = append(interfaces, current)
	}
	sort.Slice(interfaces, func(i, j int) bool { return interfaces[i].Name < interfaces[j].Name })
	c.prev = &snap

	return Metrics{
		Availability: Availability{Available: true},
		CPUPercent:   cpuPct,
		Memory:       mem,
		Interfaces:   interfaces,
	}
}

func (c *Collector) fakeMetrics() Metrics {
	c.mu.Lock()
	sample := fakeMetricSamples[c.fakeIx%len(fakeMetricSamples)]
	c.fakeIx++
	index := uint64(c.fakeIx)
	c.mu.Unlock()

	return Metrics{
		Availability: Availability{Available: true},
		CPUPercent:   sample.cpu,
		Memory: MemoryMetrics{
			UsedBytes:  9261023232,
			TotalBytes: 17179869184,
			UsedPct:    53.9,
		},
		Interfaces: []InterfaceIO{
			{Name: "br-lan", RxBytes: 3920010221 + index*15000, TxBytes: 882000122 + index*9000, RxBps: sample.lanRx, TxBps: sample.lanTx, OperState: "up", AddressCIDR: "192.168.88.1/24"},
			{Name: "tailscale0", RxBytes: 229001212 + index*4000, TxBytes: 102004002 + index*2500, RxBps: sample.tsRx, TxBps: sample.tsTx, OperState: "up", AddressCIDR: "100.101.102.103/32"},
			{Name: "wan0", RxBytes: 19220010221 + index*22000, TxBytes: 7829910021 + index*19000, RxBps: sample.wanRx, TxBps: sample.wanTx, OperState: "up", AddressCIDR: "100.64.12.9/24"},
		},
	}
}

var fakeMetricSamples = []struct {
	cpu          float64
	lanRx, lanTx float64
	tsRx, tsTx   float64
	wanRx, wanTx float64
}{
	{cpu: 28, lanRx: 420000, lanTx: 180000, tsRx: 42000, tsTx: 32000, wanRx: 780000, wanTx: 520000},
	{cpu: 36, lanRx: 760000, lanTx: 260000, tsRx: 61000, tsTx: 44000, wanRx: 1260000, wanTx: 780000},
	{cpu: 31, lanRx: 540000, lanTx: 150000, tsRx: 39000, tsTx: 28000, wanRx: 960000, wanTx: 470000},
	{cpu: 42, lanRx: 890000, lanTx: 310000, tsRx: 72000, tsTx: 53000, wanRx: 1480000, wanTx: 910000},
	{cpu: 34, lanRx: 610000, lanTx: 220000, tsRx: 50000, tsTx: 37000, wanRx: 1040000, wanTx: 660000},
}

func (c *Collector) LAN(ctx context.Context) LANInfo {
	out, err := c.runner.Run(ctx, "ip", "-j", "addr", "show", "scope", "global")
	if err != nil {
		return LANInfo{Availability: errUnavailable("ip", err)}
	}
	addresses, err := parseIPJSON(out)
	if err != nil {
		return LANInfo{Availability: errUnavailable("ip", err)}
	}
	return LANInfo{Availability: Availability{Available: true}, Addresses: addresses}
}

func (c *Collector) Connectivity(ctx context.Context) []Probe {
	targets := []Probe{
		{Name: "Google DNS", Host: "8.8.8.8"},
		{Name: "Cloudflare DNS", Host: "1.1.1.1"},
	}
	for i := range targets {
		start := c.now()
		out, err := c.runner.Run(ctx, "ping", "-c", "1", "-W", "2", targets[i].Host)
		targets[i].OK = err == nil
		if err == nil {
			if ms, ok := parsePingMS(out); ok {
				targets[i].MS = ms
			} else if elapsed := c.now().Sub(start).Milliseconds(); elapsed > 0 {
				targets[i].MS = elapsed
			}
		}
		if err != nil {
			targets[i].Err = err.Error()
		}
	}
	return targets
}

func (c *Collector) Tailscale(ctx context.Context) TailscaleStatus {
	return c.TailscalePage(ctx, PageRequest{Page: 1, PageSize: 10})
}

func (c *Collector) TailscalePage(ctx context.Context, pageRequest PageRequest) TailscaleStatus {
	out, err := c.runner.Run(ctx, "tailscale", "status", "--json")
	if err != nil {
		return TailscaleStatus{Availability: errUnavailable("tailscale", err)}
	}
	status, err := parseTailscale(out, c.now())
	if err != nil {
		return TailscaleStatus{Availability: errUnavailable("tailscale", err)}
	}
	if status.Uptime == "" {
		status.Uptime = c.tailscaleServiceUptime(ctx)
	}
	status.Peers, status.Page = paginate(status.Peers, pageRequest)
	status.Availability = Availability{Available: true}
	return status
}

func (c *Collector) tailscaleServiceUptime(ctx context.Context) string {
	out, err := c.runner.Run(ctx, "systemctl", "show", "tailscaled.service", "--property=ActiveEnterTimestamp", "--value")
	if err != nil {
		return ""
	}
	started, ok := parseSystemdTimestamp(strings.TrimSpace(out))
	if !ok {
		return ""
	}
	return humanDuration(c.now().Sub(started))
}

func (c *Collector) Rathole(ctx context.Context) RatholeStatus {
	out, err := c.runner.Run(
		ctx,
		"systemctl",
		"list-units",
		"--type=service",
		"--all",
		"--no-legend",
		"--plain",
		"ratholec@*.service",
		"rathole@*.service",
		"rathole.service",
	)
	if err != nil {
		return RatholeStatus{Availability: errUnavailable("systemctl", err)}
	}
	units := parseRatholeUnits(out)
	if len(units) == 0 {
		return RatholeStatus{Availability: Availability{Available: false, Error: "no rathole systemd units found"}}
	}
	status := RatholeStatus{
		Availability: Availability{Available: true},
		Units:        units,
		State:        ratholeStateSummary(units),
	}
	var outputs []string
	for _, unit := range units {
		if unit.Active == "active" {
			status.Active = true
		}
		detail, statusErr := c.runner.Run(ctx, "systemctl", "status", "--no-pager", unit.Name)
		if statusErr == nil && strings.TrimSpace(detail) != "" {
			outputs = append(outputs, strings.TrimSpace(detail))
		}
	}
	status.Output = strings.Join(outputs, "\n\n")
	return status
}

func (c *Collector) Firewall(ctx context.Context) FirewallStatus {
	nft, nftErr := c.runner.Run(ctx, "nft", "list", "ruleset")
	ipt, iptErr := c.runner.Run(ctx, "iptables-save")
	return FirewallStatus{
		NFTables: AvailabilityValue{Availability: errUnavailable("nft", nftErr), Value: strings.TrimSpace(nft)},
		IPTables: AvailabilityValue{Availability: errUnavailable("iptables-save", iptErr), Value: strings.TrimSpace(ipt)},
	}
}

func (c *Collector) Routes(ctx context.Context) RouteTables {
	return c.RoutesPage(ctx, PageRequest{Page: 1, PageSize: 50})
}

func (c *Collector) RoutesPage(ctx context.Context, pageRequest PageRequest) RouteTables {
	out, err := c.runner.Run(ctx, "ip", "route", "show", "table", "all")
	lines := nonEmptyLines(out)
	pageLines, page := paginate(lines, pageRequest)
	return RouteTables{
		Availability: errUnavailable("ip", err),
		Output:       strings.Join(pageLines, "\n"),
		Lines:        pageLines,
		Page:         page,
	}
}

func (c *Collector) DHCPLeasesPage(ctx context.Context, pageRequest PageRequest) DHCPLeases {
	if c.fake {
		leases, err := parseDHCPLeases(fakeDHCPLeases, c.now())
		if err != nil {
			return DHCPLeases{Availability: Availability{Available: false, Error: err.Error()}}
		}
		pageLeases, page := paginate(leases, pageRequest)
		return DHCPLeases{
			Availability: Availability{Available: true},
			Path:         "fake:/dnsmasq.leases",
			Leases:       pageLeases,
			Page:         page,
		}
	}
	raw, path, err := readDHCPLeasesFile(ctx)
	if err != nil {
		return DHCPLeases{Availability: Availability{Available: false, Error: err.Error()}}
	}
	leases, err := parseDHCPLeases(raw, c.now())
	if err != nil {
		return DHCPLeases{Availability: Availability{Available: false, Error: err.Error()}, Path: path}
	}
	pageLeases, page := paginate(leases, pageRequest)
	return DHCPLeases{
		Availability: Availability{Available: true},
		Path:         path,
		Leases:       pageLeases,
		Page:         page,
	}
}

func (c *Collector) FRR(ctx context.Context) FRRStatus {
	ospf, ospfErr := c.runner.Run(ctx, "vtysh", "-c", "show ip ospf neighbor")
	bgp, bgpErr := c.runner.Run(ctx, "vtysh", "-c", "show bgp summary")
	config, configErr := c.runner.Run(ctx, "vtysh", "-c", "show running-config")
	return FRRStatus{
		OSPF:          AvailabilityValue{Availability: errUnavailable("vtysh ospf", ospfErr), Value: strings.TrimSpace(ospf)},
		BGP:           AvailabilityValue{Availability: errUnavailable("vtysh bgp", bgpErr), Value: strings.TrimSpace(bgp)},
		RunningConfig: AvailabilityValue{Availability: errUnavailable("vtysh config", configErr), Value: strings.TrimSpace(config)},
	}
}

func (c *Collector) Diagnostic(ctx context.Context, req DiagnosticRequest) DiagnosticResult {
	target := strings.TrimSpace(req.Target)
	if !validTarget(target) {
		return DiagnosticResult{Availability: Availability{Available: false, Error: "target must be a hostname or IP address"}, Tool: req.Tool, Target: target}
	}
	switch req.Tool {
	case "ping":
		out, err := c.runWithTimeout(ctx, 12*time.Second, "ping", "-c", "4", "-W", "2", target)
		return DiagnosticResult{Availability: errUnavailable("ping", err), Tool: req.Tool, Target: target, Output: strings.TrimSpace(out)}
	case "mtr":
		out, err := c.runWithTimeout(ctx, 35*time.Second, "mtr", "-r", "-b", "-w", "-c10", target)
		return DiagnosticResult{Availability: errUnavailable("mtr", err), Tool: req.Tool, Target: target, Output: strings.TrimSpace(out)}
	default:
		return DiagnosticResult{Availability: Availability{Available: false, Error: "tool must be ping or mtr"}, Tool: req.Tool, Target: target}
	}
}

func (c *Collector) runWithTimeout(ctx context.Context, timeout time.Duration, name string, args ...string) (string, error) {
	if runner, ok := c.runner.(TimeoutRunner); ok {
		return runner.RunWithTimeout(ctx, timeout, name, args...)
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return c.runner.Run(ctx, name, args...)
}

type cpuStat struct {
	idle  uint64
	total uint64
}

func readProcMetrics() (cpuStat, MemoryMetrics, map[string]InterfaceIO, error) {
	statRaw, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuStat{}, MemoryMetrics{}, nil, err
	}
	memRaw, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return cpuStat{}, MemoryMetrics{}, nil, err
	}
	netRaw, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return cpuStat{}, MemoryMetrics{}, nil, err
	}
	stat, err := parseCPUStat(string(statRaw))
	if err != nil {
		return cpuStat{}, MemoryMetrics{}, nil, err
	}
	mem, err := parseMeminfo(string(memRaw))
	if err != nil {
		return cpuStat{}, MemoryMetrics{}, nil, err
	}
	return stat, mem, parseNetDev(string(netRaw)), nil
}

func parseCPUStat(raw string) (cpuStat, error) {
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 || fields[0] != "cpu" {
			continue
		}
		var values []uint64
		for _, field := range fields[1:] {
			value, err := strconv.ParseUint(field, 10, 64)
			if err != nil {
				return cpuStat{}, err
			}
			values = append(values, value)
		}
		var total uint64
		for _, value := range values {
			total += value
		}
		return cpuStat{idle: values[3], total: total}, nil
	}
	return cpuStat{}, fmt.Errorf("cpu line not found")
}

func parseMeminfo(raw string) (MemoryMetrics, error) {
	values := map[string]uint64{}
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			values[strings.TrimSuffix(fields[0], ":")] = value * 1024
		}
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 {
		return MemoryMetrics{}, fmt.Errorf("MemTotal missing")
	}
	used := total - available
	return MemoryMetrics{UsedBytes: used, TotalBytes: total, UsedPct: 100 * float64(used) / float64(total)}, nil
}

func parseNetDev(raw string) map[string]InterfaceIO {
	result := map[string]InterfaceIO{}
	for _, line := range strings.Split(raw, "\n") {
		parts := strings.Split(line, ":")
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 16 || name == "lo" {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		result[name] = InterfaceIO{Name: name, RxBytes: rx, TxBytes: tx}
	}
	return result
}

type ipAddrJSON struct {
	IfName   string `json:"ifname"`
	AddrInfo []struct {
		Family    string `json:"family"`
		Local     string `json:"local"`
		PrefixLen int    `json:"prefixlen"`
	} `json:"addr_info"`
}

func parseIPJSON(raw string) ([]AddressInfo, error) {
	var devices []ipAddrJSON
	if err := json.Unmarshal([]byte(raw), &devices); err != nil {
		return nil, err
	}
	var addresses []AddressInfo
	for _, device := range devices {
		for _, addr := range device.AddrInfo {
			if addr.Family == "inet" {
				addresses = append(addresses, AddressInfo{Interface: device.IfName, CIDR: fmt.Sprintf("%s/%d", addr.Local, addr.PrefixLen)})
			}
		}
	}
	return addresses, nil
}

func parseRatholeUnits(raw string) []RatholeUnit {
	var units []RatholeUnit
	for _, line := range strings.Split(raw, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		description := ""
		if len(fields) > 4 {
			description = strings.Join(fields[4:], " ")
		}
		units = append(units, RatholeUnit{
			Name:        fields[0],
			Load:        fields[1],
			Active:      fields[2],
			Sub:         fields[3],
			Description: description,
		})
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Name < units[j].Name })
	return units
}

func ratholeStateSummary(units []RatholeUnit) string {
	active := 0
	failed := 0
	for _, unit := range units {
		switch unit.Active {
		case "active":
			active++
		case "failed":
			failed++
		}
	}
	if len(units) == 1 {
		state := units[0].Active
		if units[0].Sub != "" && units[0].Sub != state {
			state += " (" + units[0].Sub + ")"
		}
		return state
	}
	parts := []string{fmt.Sprintf("%d units", len(units)), fmt.Sprintf("%d active", active)}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	return strings.Join(parts, ", ")
}

var pingTimePattern = regexp.MustCompile(`time[=<]\s*([0-9]+(?:\.[0-9]+)?)\s*ms`)

func parsePingMS(raw string) (int64, bool) {
	match := pingTimePattern.FindStringSubmatch(raw)
	if len(match) != 2 {
		return 0, false
	}
	value, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, false
	}
	return int64(math.Max(1, math.Ceil(value))), true
}

type tailscaleJSON struct {
	BackendState string `json:"BackendState"`
	Self         struct {
		TailscaleIPs     []string `json:"TailscaleIPs"`
		AllowedIPs       []string `json:"AllowedIPs"`
		AdvertisedRoutes []string `json:"AdvertisedRoutes"`
		PrimaryRoutes    []string `json:"PrimaryRoutes"`
		Started          string   `json:"Started"`
	} `json:"Self"`
	Peer map[string]struct {
		HostName      string   `json:"HostName"`
		DNSName       string   `json:"DNSName"`
		Online        bool     `json:"Online"`
		AllowedIPs    []string `json:"AllowedIPs"`
		PrimaryRoutes []string `json:"PrimaryRoutes"`
		LastSeen      string   `json:"LastSeen"`
	} `json:"Peer"`
}

func parseTailscale(raw string, now time.Time) (TailscaleStatus, error) {
	var source tailscaleJSON
	if err := json.Unmarshal([]byte(raw), &source); err != nil {
		return TailscaleStatus{}, err
	}
	status := TailscaleStatus{
		BackendState:    source.BackendState,
		SelfIPs:         source.Self.TailscaleIPs,
		AcceptingRoutes: len(source.Self.AllowedIPs) > len(source.Self.TailscaleIPs),
	}
	status.AdvertisedRoutes = source.Self.AdvertisedRoutes
	if len(status.AdvertisedRoutes) == 0 {
		status.AdvertisedRoutes = source.Self.PrimaryRoutes
	}
	if started, err := time.Parse(time.RFC3339, source.Self.Started); err == nil {
		status.Uptime = humanDuration(now.Sub(started))
	}
	for key, peer := range source.Peer {
		name := peer.HostName
		if name == "" {
			name = strings.TrimSuffix(peer.DNSName, ".")
		}
		if name == "" {
			name = key
		}
		status.Peers = append(status.Peers, TailscalePeer{
			Name:           name,
			Online:         peer.Online,
			AllowedIPs:     peer.AllowedIPs,
			ReceivedRoutes: peer.PrimaryRoutes,
			LastSeen:       peer.LastSeen,
		})
	}
	sort.Slice(status.Peers, func(i, j int) bool { return status.Peers[i].Name < status.Peers[j].Name })
	return status, nil
}

func parseSystemdTimestamp(raw string) (time.Time, bool) {
	if raw == "" || raw == "n/a" {
		return time.Time{}, false
	}
	layouts := []string{
		"Mon 2006-01-02 15:04:05 -07",
		"Mon 2006-01-02 15:04:05 MST",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func nonEmptyLines(raw string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func readDHCPLeasesFile(ctx context.Context) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	candidates := []string{
		os.Getenv("ROUTERDASH_DHCP_LEASES_FILE"),
		"/tmp/dhcp.leases",
		"/var/lib/misc/dnsmasq.leases",
		"/var/lib/dnsmasq/dnsmasq.leases",
	}
	var tried []string
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		raw, err := os.ReadFile(candidate)
		if err == nil {
			return string(raw), candidate, nil
		}
		tried = append(tried, candidate)
	}
	return "", "", fmt.Errorf("dnsmasq leases file not found; tried %s", strings.Join(tried, ", "))
}

func parseDHCPLeases(raw string, now time.Time) ([]DHCPLease, error) {
	var leases []DHCPLease
	for lineNumber, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) < 5 {
			return nil, fmt.Errorf("invalid dnsmasq lease line %d", lineNumber+1)
		}
		expiresUnix, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid dnsmasq lease expiry on line %d: %w", lineNumber+1, err)
		}
		lease := DHCPLease{
			MAC:      fields[1],
			IP:       fields[2],
			Hostname: emptyStar(fields[3]),
			ClientID: emptyStar(strings.Join(fields[4:], " ")),
		}
		if expiresUnix == 0 {
			lease.ExpiresAt = "never"
			lease.Remaining = "never"
		} else {
			expiresAt := time.Unix(expiresUnix, 0)
			lease.ExpiresAt = expiresAt.Format(time.RFC3339)
			lease.Expired = !expiresAt.After(now)
			if lease.Expired {
				lease.Remaining = "expired"
			} else {
				lease.Remaining = humanDuration(expiresAt.Sub(now))
			}
		}
		leases = append(leases, lease)
	}
	return leases, nil
}

func emptyStar(value string) string {
	if value == "*" {
		return ""
	}
	return value
}

func paginate[T any](items []T, request PageRequest) ([]T, *PageInfo) {
	pageSize := request.PageSize
	if pageSize <= 0 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}
	total := len(items)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	page := request.Page
	if page < 1 {
		page = 1
	}
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return items[start:end], &PageInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasPrev:    page > 1,
		HasNext:    totalPages > 0 && page < totalPages,
	}
}

func validTarget(target string) bool {
	if target == "" || len(target) > 253 {
		return false
	}
	if ip := net.ParseIP(target); ip != nil {
		return true
	}
	matched, _ := regexp.MatchString(`^[A-Za-z0-9][A-Za-z0-9.-]*[A-Za-z0-9]$`, target)
	return matched && !strings.Contains(target, "..")
}

func unavailableValue(message string) AvailabilityValue {
	return AvailabilityValue{Availability: Availability{Available: false, Error: message}}
}

func humanDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
