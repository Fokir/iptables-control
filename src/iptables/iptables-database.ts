import { JSONFilePreset } from "lowdb/node";
import { Low } from "lowdb";
import { IptablesDatabaseInterface } from "@/iptables/interfaces/iptables-database.interface";

export async function getDatabaseConnection(): Promise<
  Low<IptablesDatabaseInterface>
> {
  return await JSONFilePreset<IptablesDatabaseInterface>(
    "database/iptables.json",
    {
      groups: [],
    },
  );
}
