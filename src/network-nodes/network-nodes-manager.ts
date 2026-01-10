"use server";

import {JSONFilePreset} from "lowdb/node";
import {Low} from "lowdb";
import {resolve} from "path";
import {NetworkNodesDatabaseInterface} from "@/network-nodes/interfaces/network-nodes-database.interface";
import {NetworkNodeItemDatabaseInterface} from "@/network-nodes/interfaces/network-node-item-database.interface";

async function getDatabaseConnection(): Promise<
    Low<NetworkNodesDatabaseInterface>
> {
    const connection = await JSONFilePreset<NetworkNodesDatabaseInterface>(
        resolve(process.cwd(), 'database/network-nodes.json'),
        {
            nodes: [],
        },
    );

    await connection.read();

    return connection;
}

export async function getNetworkNodes(): Promise<
    NetworkNodesDatabaseInterface["nodes"]
> {
    const connection = await getDatabaseConnection();

    await connection.read();

    return connection.data.nodes;
}

export async function addNetworkNode(
    node: NetworkNodeItemDatabaseInterface,
): Promise<void> {
    const connection = await getDatabaseConnection();

    connection.data.nodes = connection.data.nodes.filter(
        (item) => item.ip !== node.ip,
    );
    connection.data.nodes.push(node);

    await connection.write();
}

export async function deleteNetworkNode(
    node: NetworkNodeItemDatabaseInterface,
): Promise<void> {
    const connection = await getDatabaseConnection();

    connection.data.nodes = connection.data.nodes.filter(
        (item) => node.ip !== item.ip,
    );

    await connection.write();
}
