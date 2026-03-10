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

export interface NginxDomain {
  id: number
  domain: string
  upstreamIp: string
  upstreamPort: number
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

export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
}
