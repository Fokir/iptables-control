import {IptablesDatabaseGroupInterface} from "@/iptables/interfaces/iptables-database-group.interface";

const IPV4_REGEX = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

function validateIp(ip: string): boolean {
  return IPV4_REGEX.test(ip);
}

function validatePort(port: number): boolean {
  return Number.isInteger(port) && port >= 1 && port <= 65535;
}

export function validateGroup(group: IptablesDatabaseGroupInterface): boolean {
  return (
      !!group.name &&
      !!group.id &&
      !!group.destinationIp &&
      !!group.targetReverseIp &&
      !!group.targetIp &&
      validateIp(group.destinationIp) &&
      validateIp(group.targetReverseIp) &&
      validateIp(group.targetIp) &&
      group.ports.length > 0 &&
      group.ports.every((port) =>
          validatePort(port.value) &&
          validatePort(port.valueTarget) &&
          port.protocols.length > 0
      )
  );
}
