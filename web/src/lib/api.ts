export type Availability = {
  available: boolean;
  error?: string;
};

export type AvailabilityValue = Availability & {
  value?: string;
};

export type Summary = {
  hostname: string;
  uptime: AvailabilityValue;
  wanIp: AvailabilityValue;
  lan: LANInfo;
  connectivity: Probe[];
  tailscale: TailscaleStatus;
  rathole: RatholeStatus;
};

export type Metrics = Availability & {
  cpuPercent: number;
  memory: {
    usedBytes: number;
    totalBytes: number;
    usedPct: number;
  };
  interfaces: InterfaceIO[];
};

export type InterfaceIO = {
  name: string;
  rxBytes: number;
  txBytes: number;
  rxBps: number;
  txBps: number;
  operState?: string;
  addressCidr?: string;
};

export type LANInfo = Availability & {
  addresses: { interface: string; cidr: string }[];
};

export type Probe = {
  name: string;
  host: string;
  ok: boolean;
  ms?: number;
  error?: string;
};

export type TailscaleStatus = Availability & {
  backendState?: string;
  acceptingRoutes: boolean;
  advertisedRoutes?: string[];
  uptime?: string;
  selfIps?: string[];
  peers?: TailscalePeer[];
  page?: PageInfo;
};

export type TailscalePeer = {
  name: string;
  online: boolean;
  receivedRoutes?: string[];
  allowedIps?: string[];
  lastSeen?: string;
};

export type RatholeStatus = Availability & {
  active: boolean;
  state?: string;
  output?: string;
  units?: RatholeUnit[];
};

export type RatholeUnit = {
  name: string;
  load?: string;
  active?: string;
  sub?: string;
  description?: string;
};

export type FirewallStatus = {
  nftables: AvailabilityValue;
  iptables: AvailabilityValue;
};

export type RouteTables = Availability & {
  output?: string;
  lines?: string[];
  page?: PageInfo;
};

export type DHCPLeases = Availability & {
  path?: string;
  leases?: DHCPLease[];
  page?: PageInfo;
};

export type DHCPLease = {
  expiresAt?: string;
  remaining?: string;
  mac: string;
  ip: string;
  hostname?: string;
  clientId?: string;
  expired: boolean;
};

export type PageInfo = {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
  hasPrev: boolean;
  hasNext: boolean;
};

export type FRRStatus = {
  ospf: AvailabilityValue;
  bgp: AvailabilityValue;
  runningConfig: AvailabilityValue;
};

export type DiagnosticResult = Availability & {
  tool: string;
  target: string;
  output?: string;
};

export type VersionInfo = {
  version: string;
};

export async function getJSON<T>(path: string, signal?: AbortSignal): Promise<T> {
  const response = await fetch(path, { signal });
  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<T>;
}

export async function postDiagnostic(tool: string, target: string): Promise<DiagnosticResult> {
  const response = await fetch("/api/diagnostics", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ tool, target }),
  });
  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText}`);
  }
  return response.json() as Promise<DiagnosticResult>;
}
