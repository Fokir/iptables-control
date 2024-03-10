import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";

export function validateNode(node: NetworkNodeItemDatabaseInterface): boolean {
  return !!node.name && !!node.ip;
}
