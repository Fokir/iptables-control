"use server";

import { IptablesDatabaseInterface } from "@/iptables/interfaces/iptables-database.interface";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";
import { getDatabaseConnection } from "@/iptables/iptables-database";
import { execAppendGroup, execRemoveGroup } from "@/iptables/iptables-exec";

const connectionPromise = getDatabaseConnection();

export async function getIptablesGroups(): Promise<
  IptablesDatabaseInterface["groups"]
> {
  const connection = await connectionPromise;

  await connection.read();

  return connection.data.groups;
}

export async function addIptablesGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await connectionPromise;

  connection.data.groups.push(group);

  await connection.write();

  if (group.enabled) await execAppendGroup(group);
}

export async function updateIptablesGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await connectionPromise;
  const index = connection.data.groups.findIndex(
    (item) => item.id === group.id,
  );

  if (index > -1) {
    await execRemoveGroup(connection.data.groups[index]);

    connection.data.groups[index] = group;
  }

  await connection.write();

  if (group.enabled) await execAppendGroup(group);
}

export async function updateOrCreateGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await connectionPromise;

  const index = connection.data.groups.findIndex(
    (item) => item.id === group.id,
  );

  if (index > -1) {
    await updateIptablesGroup(group);
  } else {
    await addIptablesGroup(group);
  }
}

export async function deleteGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await connectionPromise;

  const index = connection.data.groups.findIndex(
    (item) => group.id !== item.id,
  );

  await execRemoveGroup(connection.data.groups[index]);

  connection.data.groups = connection.data.groups.filter(
    (item) => group.id !== item.id,
  );

  await connection.write();
}
