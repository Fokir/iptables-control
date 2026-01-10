"use server";

import { IptablesDatabaseInterface } from "@/iptables/interfaces/iptables-database.interface";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";
import { getDatabaseConnection } from "@/iptables/iptables-database";
import { execAppendGroup, execRemoveGroup, saveRules } from "@/iptables/iptables-exec";

export async function getIptablesGroups(): Promise<
  IptablesDatabaseInterface["groups"]
> {
  const connection = await getDatabaseConnection();

  await connection.read();

  return connection.data.groups;
}

export async function addIptablesGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await getDatabaseConnection();

  connection.data.groups.push(group);

  await connection.write();

  if (group.enabled) {
    await execAppendGroup(group);
    await saveRules();
  }
}

export async function updateIptablesGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await getDatabaseConnection();
  const index = connection.data.groups.findIndex(
    (item) => item.id === group.id,
  );

  if (index > -1) {
    const groupInDb = connection.data.groups[index];

    if(groupInDb.enabled) {
      try {
          await execRemoveGroup(connection.data.groups[index]);
      } catch (error) {
          console.error(error);
      }
    }

    connection.data.groups[index] = group;
  }

  await connection.write();

  if (group.enabled) await execAppendGroup(group);

  await saveRules();
}

export async function updateOrCreateGroup(
  group: IptablesDatabaseGroupInterface,
): Promise<void> {
  const connection = await getDatabaseConnection();

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
  const connection = await getDatabaseConnection();

  const index = connection.data.groups.findIndex(
    (item) => item.id === group.id,
  );

  if (index > -1) {
    const groupInDb = connection.data.groups[index];
    if (groupInDb.enabled) {
      try {
          await execRemoveGroup(connection.data.groups[index]);
      } catch (error) {
          console.error(error);
      }
    }

    connection.data.groups = connection.data.groups.filter(
      (item) => item.id !== group.id,
    );

    await connection.write();
    await saveRules();
  }
}
