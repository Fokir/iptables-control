import { getIptablesGroups } from "@/iptables/iptables-manager";
import IptablesContent from "@/app/iptables/iptables-content";
import { getNetworkNodes } from "@/network-nodes/network-nodes-manager";

export default async function IpTablesPage() {
  const groups = await getIptablesGroups();
  const nodes = await getNetworkNodes();

  return <IptablesContent groups={groups} nodes={nodes}></IptablesContent>;
}
