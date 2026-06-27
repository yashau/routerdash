package routerdash

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type FakeRunner struct{}

func (FakeRunner) Run(_ context.Context, name string, args ...string) (string, error) {
	key := name + " " + strings.Join(args, " ")
	switch {
	case name == "curl":
		return "203.0.113.42\n", nil
	case key == "ip -j addr show scope global":
		return fakeIPAddr, nil
	case key == "ip route show table all":
		return fakeRoutes, nil
	case name == "ping":
		return "PING host: 56 data bytes\n64 bytes from host: icmp_seq=1 ttl=57 time=12.8 ms\n", nil
	case key == "tailscale status --json":
		return fakeTailscale, nil
	case key == "tailscale debug prefs":
		return fakeTailscalePrefs, nil
	case key == "systemctl show tailscaled.service --property=ActiveEnterTimestamp --value":
		return "Thu 2026-06-25 08:00:00 +00\n", nil
	case strings.HasPrefix(key, "systemctl list-units"):
		return fakeRatholeUnits, nil
	case strings.HasPrefix(key, "systemctl status"):
		return fakeRatholeStatus(key), nil
	case key == "nft list ruleset":
		return fakeNFT, nil
	case key == "iptables-save":
		return fakeIPTables, nil
	case strings.Contains(key, "show ip ospf"):
		return fakeOSPF, nil
	case strings.Contains(key, "show bgp"):
		return fakeBGP, nil
	case strings.Contains(key, "show running-config"):
		return fakeFRRConfig, nil
	case name == "mtr":
		return fakeMTR, nil
	default:
		return "", fmt.Errorf("fake output missing for %s", key)
	}
}

func (runner FakeRunner) RunWithTimeout(ctx context.Context, _ time.Duration, name string, args ...string) (string, error) {
	return runner.Run(ctx, name, args...)
}

const fakeIPAddr = `[
	{"ifname":"br-lan","addr_info":[{"family":"inet","local":"192.168.88.1","prefixlen":24}]},
	{"ifname":"wan0","addr_info":[{"family":"inet","local":"100.64.12.9","prefixlen":24}]}
]`

const fakeRoutes = `default via 100.64.12.1 dev wan0 table main proto dhcp src 100.64.12.9
192.168.88.0/24 dev br-lan table main proto kernel scope link src 192.168.88.1
10.42.0.0/16 dev tailscale0 table 52
local 192.168.88.1 dev br-lan table local proto kernel scope host src 192.168.88.1`

const fakeDHCPLeases = `1782943200 84:2a:fd:11:22:33 192.168.88.24 workstation 01:84:2a:fd:11:22:33
1782945000 dc:a6:32:44:55:66 192.168.88.31 printer *
0 3c:06:30:77:88:99 192.168.88.2 core-switch *
1782939600 aa:bb:cc:dd:ee:ff 192.168.88.52 phone 01:aa:bb:cc:dd:ee:ff`

const fakeTailscale = `{
	"BackendState": "Running",
	"Self": {
		"TailscaleIPs": ["100.101.102.103"],
		"AllowedIPs": ["100.101.102.103/32", "10.10.0.0/16"],
		"PrimaryRoutes": ["192.168.88.0/24"]
	},
	"Peer": {
		"node-a": {
			"HostName": "nas",
			"Online": true,
			"AllowedIPs": ["100.111.0.1/32", "10.10.0.0/16"],
			"PrimaryRoutes": ["10.10.0.0/16"]
		},
		"node-b": {
			"HostName": "phone",
			"Online": false,
			"AllowedIPs": ["100.111.0.2/32"],
			"LastSeen": "2026-06-25T10:02:00Z"
		}
	},
	"CurrentTailnet": {"MagicDNSSuffix": "tailnet.example.ts.net"}
}`

const fakeTailscalePrefs = `{
	"AdvertiseRoutes": ["192.168.88.0/24", "192.168.99.0/24"]
}`

const fakeRatholeUnits = `ratholec@edge.service loaded active running Rathole client edge tunnel
ratholec@backup.service loaded inactive dead Rathole client backup tunnel`

func fakeRatholeStatus(key string) string {
	if strings.Contains(key, "ratholec@backup.service") {
		return "ratholec@backup.service - Rathole client backup tunnel\n   Loaded: loaded (/etc/systemd/system/ratholec@.service; disabled)\n   Active: inactive (dead) since Thu 2026-06-25 11:45:00 UTC\n"
	}
	return "ratholec@edge.service - Rathole client edge tunnel\n   Loaded: loaded (/etc/systemd/system/ratholec@.service; enabled)\n   Active: active (running) since Thu 2026-06-25 08:00:00 UTC\n"
}

const fakeNFT = `table inet filter {
	chain input {
		type filter hook input priority filter; policy drop;
		iifname "br-lan" accept
		ct state established,related accept
	}
}`

const fakeIPTables = `*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT ACCEPT [0:0]
-A INPUT -i br-lan -j ACCEPT
COMMIT`

const fakeOSPF = `Neighbor ID     Pri State           Dead Time Address         Interface
10.0.0.2          1 Full/DR           32.221s 192.168.88.2    br-lan:192.168.88.1`

const fakeBGP = `IPv4 Unicast Summary:
BGP router identifier 192.168.88.1, local AS number 64512 vrf-id 0
Neighbor        V         AS   MsgRcvd   MsgSent   TblVer  InQ OutQ  Up/Down State/PfxRcd
100.64.12.2     4      64513      4801      4812        0    0    0 01:12:44            42`

const fakeFRRConfig = `frr version 10.0
hostname routerdash-lab
router ospf
 network 192.168.88.0/24 area 0
router bgp 64512
 neighbor 100.64.12.2 remote-as 64513
 address-family ipv4 unicast
  network 192.168.88.0/24
 exit-address-family`

const fakeMTR = `Start: 2026-06-25T12:00:00+0000
HOST: routerdash-lab       Loss%   Snt   Last   Avg  Best  Wrst StDev
  1. AS??? 192.168.88.1     0.0%    10    0.3   0.4   0.2   0.8   0.2
  2. AS13335 1.1.1.1        0.0%    10   12.1  12.4  11.8  13.1   0.4`
