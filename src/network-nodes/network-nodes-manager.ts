"use server";

import { JSONFilePreset } from "lowdb/node";
import { Low } from "lowdb";
import { NetworkNodesDatabaseInterface } from "@/network-nodes/interfaces/network-nodes-database.interface";
import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";

async function getDatabaseConnection(): Promise<
  Low<NetworkNodesDatabaseInterface>
> {
  return await JSONFilePreset<NetworkNodesDatabaseInterface>(
    "database/network-nodes.json",
    {
      nodes: [],
    },
  );
}

const connectionPromise = getDatabaseConnection();

export async function getNetworkNodes(): Promise<
  NetworkNodesDatabaseInterface["nodes"]
> {
  const connection = await connectionPromise;

  await connection.read();

  return connection.data.nodes;
}

export async function addNetworkNode(
  node: NetworkNodeItemDatabaseInterface,
): Promise<void> {
  const connection = await connectionPromise;

  await deleteNetworkNode(node);

  connection.data.nodes.push(node);

  await connection.write();
}

export async function deleteNetworkNode(
  node: NetworkNodeItemDatabaseInterface,
): Promise<void> {
  const connection = await connectionPromise;

  connection.data.nodes = connection.data.nodes.filter(
    (item) => node.ip !== item.ip,
  );

  await connection.write();
}
