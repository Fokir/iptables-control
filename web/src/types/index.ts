export interface NatGroup {
  id: number
  name: string
  enabled: boolean
  targetIp: string
  targetReverseIp: string
  destinationIp: string
  ports: NatPort[]
  createdAt: string
  updatedAt: string
}

export interface NatPort {
  id: number
  groupId: number
  externalPort: number
  internalPort: number
  protocolTcp: boolean
  protocolUdp: boolean
  createdAt: string
}

export interface CreateGroupRequest {
  name: string
  enabled: boolean
  targetIp: string
  targetReverseIp?: string
  destinationIp?: string
  ports: CreatePortRequest[]
}

export interface CreatePortRequest {
  externalPort: number
  internalPort: number
  protocolTcp: boolean
  protocolUdp: boolean
}

export interface NetworkNode {
  id: number
  name: string
  ip: string
  createdAt: string
}

export interface NodeStatus {
  nodeId: number
  isOnline: boolean
  latencyMs: number
  firstOnline: string | null
  lastSeen: string | null
}

export interface NginxDomain {
  id: number
  domain: string
  upstreamIp: string
  upstreamPort: number
  upstreamScheme: string
  upstreamSslVerify: boolean
  sslEnabled: boolean
  enabled: boolean
  basicAuthUser: string
  basicAuthPassword: string
  createdAt: string
  updatedAt: string
}

export interface ExternalNginxDomain {
  filename: string
  domain: string
  upstreamIp: string
  upstreamPort: number
  sslEnabled: boolean
  hasBasicAuth: boolean
}

export interface AuditLogEntry {
  id: number
  userId: number | null
  username: string
  action: string
  entityType: string
  entityId: number | null
  details: string
  ipAddress: string
  createdAt: string
}

export interface AuditLogResult {
  entries: AuditLogEntry[]
  total: number
  page: number
  limit: number
}

export interface User {
  id: number
  username: string
}

export interface TrafficPoint {
  t: string
  in: number
  out: number
}

export interface TrafficSeries {
  nodeId?: number
  nodeName?: string
  interfaceName?: string
  points: TrafficPoint[]
}

export interface WireguardStatus {
  installed: boolean
  running: boolean
  publicKey?: string
  listenPort?: number
  address?: string
  endpoint?: string
  peerCount: number
}

export interface WireguardPeer {
  id: number
  name: string
  publicKey: string
  allowedIps: string
  address: string
  dns: string
  endpoint?: string
  latestHandshake?: string
  transferRx?: number
  transferTx?: number
  createdAt: string
}

export interface WireguardPeerConfig {
  config: string
  qrCode?: string
}

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}
