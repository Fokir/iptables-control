import { useNetworkNodes } from '../../contexts/NetworkNodesContext'

interface Props {
  ip: string
  className?: string
}

export function IpWithNodeName({ ip, className }: Props) {
  const { getNodeName } = useNetworkNodes()
  const nodeName = getNodeName(ip)

  return (
    <span className={className}>
      <span className="font-mono">{ip}</span>
      {nodeName && (
        <span className="block text-[11px] text-slate-500 leading-tight">{nodeName}</span>
      )}
    </span>
  )
}
