import {JSONFilePreset} from "lowdb/node";
import {Low} from "lowdb";
import {IptablesDatabaseInterface} from "@/iptables/interfaces/iptables-database.interface";
import {resolve} from "path";

export async function getDatabaseConnection(): Promise<
    Low<IptablesDatabaseInterface>
> {
    const connection = await JSONFilePreset<IptablesDatabaseInterface>(
        resolve(process.cwd(), 'database/iptables.json'),
        {
            groups: [],
        },
    );

    await connection.read();

    return connection;
}
