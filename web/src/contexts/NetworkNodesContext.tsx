import { createContext, useContext, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { NetworkNode } from '../types'

interface NetworkNodesContextValue {
  nodes: NetworkNode[]
  getNodeName: (ip: string) => string | undefined
}

const NetworkNodesContext = createContext<NetworkNodesContextValue>({
  nodes: [],
  getNodeName: () => undefined,
})

export function NetworkNodesProvider({ children }: { children: React.ReactNode }) {
  const { data: nodes = [] } = useQuery({
    queryKey: ['network-nodes'],
    queryFn: () => api.get<NetworkNode[]>('/api/network-nodes'),
  })

  const value = useMemo(() => {
    const ipMap = new Map(nodes.map(n => [n.ip, n.name]))
    return {
      nodes,
      getNodeName: (ip: string) => ipMap.get(ip),
    }
  }, [nodes])

  return (
    <NetworkNodesContext.Provider value={value}>
      {children}
    </NetworkNodesContext.Provider>
  )
}

export function useNetworkNodes() {
  return useContext(NetworkNodesContext)
}
