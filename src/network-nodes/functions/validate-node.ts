import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";

const IPV4_REGEX = /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

export function validateIp(ip: string): boolean {
  return IPV4_REGEX.test(ip);
}

export function validateNode(node: NetworkNodeItemDatabaseInterface): boolean {
  return !!node.name && !!node.ip && validateIp(node.ip);
}
