import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

export function validateGroup(group: IptablesDatabaseGroupInterface): boolean {
  return (
    !!group.name &&
    !!group.id &&
    !!group.destinationIp &&
    !!group.targetReverseIp &&
    !!group.targetIp &&
    group.ports.every((port) => !!port.value && port.protocols.length > 0)
  );
}
