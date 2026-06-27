package routerdash

type Availability struct {
	Available bool   `json:"available"`
	Error     string `json:"error,omitempty"`
}

type Summary struct {
	Hostname     string            `json:"hostname"`
	Uptime       AvailabilityValue `json:"uptime"`
	WANIP        AvailabilityValue `json:"wanIp"`
	LAN          LANInfo           `json:"lan"`
	Connectivity []Probe           `json:"connectivity"`
	Tailscale    TailscaleStatus   `json:"tailscale"`
	Rathole      RatholeStatus     `json:"rathole"`
}

type AvailabilityValue struct {
	Availability
	Value string `json:"value,omitempty"`
}

type Metrics struct {
	Availability
	CPUPercent float64       `json:"cpuPercent"`
	Memory     MemoryMetrics `json:"memory"`
	Interfaces []InterfaceIO `json:"interfaces"`
}

type MemoryMetrics struct {
	UsedBytes  uint64  `json:"usedBytes"`
	TotalBytes uint64  `json:"totalBytes"`
	UsedPct    float64 `json:"usedPct"`
}

type InterfaceIO struct {
	Name        string  `json:"name"`
	RxBytes     uint64  `json:"rxBytes"`
	TxBytes     uint64  `json:"txBytes"`
	RxBps       float64 `json:"rxBps"`
	TxBps       float64 `json:"txBps"`
	OperState   string  `json:"operState,omitempty"`
	AddressCIDR string  `json:"addressCidr,omitempty"`
}

type LANInfo struct {
	Availability
	Addresses []AddressInfo `json:"addresses"`
}

type AddressInfo struct {
	Interface string `json:"interface"`
	CIDR      string `json:"cidr"`
}

type Probe struct {
	Name string `json:"name"`
	Host string `json:"host"`
	OK   bool   `json:"ok"`
	MS   int64  `json:"ms,omitempty"`
	Err  string `json:"error,omitempty"`
}

type TailscaleStatus struct {
	Availability
	BackendState          string           `json:"backendState,omitempty"`
	AcceptingRoutes       bool             `json:"acceptingRoutes"`
	AdvertisedRoutes      []string         `json:"advertisedRoutes,omitempty"`
	AdvertisedRouteStates []TailscaleRoute `json:"advertisedRouteStates,omitempty"`
	Uptime                string           `json:"uptime,omitempty"`
	SelfIPs               []string         `json:"selfIps,omitempty"`
	Peers                 []TailscalePeer  `json:"peers,omitempty"`
	Page                  *PageInfo        `json:"page,omitempty"`
}

type TailscaleRoute struct {
	Route    string `json:"route"`
	Approved bool   `json:"approved"`
	Status   string `json:"status"`
}

type TailscalePeer struct {
	Name           string   `json:"name"`
	Online         bool     `json:"online"`
	ReceivedRoutes []string `json:"receivedRoutes,omitempty"`
	AllowedIPs     []string `json:"allowedIps,omitempty"`
	LastSeen       string   `json:"lastSeen,omitempty"`
}

type RatholeStatus struct {
	Availability
	Active bool          `json:"active"`
	State  string        `json:"state,omitempty"`
	Output string        `json:"output,omitempty"`
	Units  []RatholeUnit `json:"units,omitempty"`
}

type RatholeUnit struct {
	Name        string `json:"name"`
	Load        string `json:"load,omitempty"`
	Active      string `json:"active,omitempty"`
	Sub         string `json:"sub,omitempty"`
	Description string `json:"description,omitempty"`
}

type FirewallStatus struct {
	NFTables AvailabilityValue `json:"nftables"`
	IPTables AvailabilityValue `json:"iptables"`
}

type RouteTables struct {
	Availability
	Output string    `json:"output,omitempty"`
	Lines  []string  `json:"lines,omitempty"`
	Page   *PageInfo `json:"page,omitempty"`
}

type DHCPLeases struct {
	Availability
	Path   string      `json:"path,omitempty"`
	Leases []DHCPLease `json:"leases,omitempty"`
	Page   *PageInfo   `json:"page,omitempty"`
}

type DHCPLease struct {
	ExpiresAt string `json:"expiresAt,omitempty"`
	Remaining string `json:"remaining,omitempty"`
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname,omitempty"`
	ClientID  string `json:"clientId,omitempty"`
	Expired   bool   `json:"expired"`
}

type PageInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasPrev    bool `json:"hasPrev"`
	HasNext    bool `json:"hasNext"`
}

type PageRequest struct {
	Page     int
	PageSize int
}

type VersionInfo struct {
	Version string `json:"version"`
}

type FRRStatus struct {
	OSPF          AvailabilityValue `json:"ospf"`
	BGP           AvailabilityValue `json:"bgp"`
	RunningConfig AvailabilityValue `json:"runningConfig"`
}

type DiagnosticRequest struct {
	Tool   string `json:"tool"`
	Target string `json:"target"`
}

type DiagnosticResult struct {
	Availability
	Tool   string `json:"tool"`
	Target string `json:"target"`
	Output string `json:"output,omitempty"`
}
