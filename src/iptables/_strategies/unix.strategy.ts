import {IptablesStrategyInterface} from "@/iptables/interfaces/iptables-strategy.interface";
import {IptablesDatabaseGroupInterface} from "@/iptables/interfaces/iptables-database-group.interface";
import {exec} from "child_process";
import {promisify} from "util";

const execPromise = promisify(exec);

export class UnixStrategy implements IptablesStrategyInterface {
    async addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
        // iptables -A PREROUTING -t nat -d 212.193.55.91 -p tcp --dport $3 -j DNAT --to-dest 10.7.0.3:$3;
        // iptables -A POSTROUTING -t nat -d 10.7.0.3 -p tcp --dport $3 -j SNAT --to-source 10.7.0.1;

        for (const port of group.ports) {
            for (const protocol of port.protocols) {
                await this.execPreRouting(
                    group.targetIp,
                    group.destinationIp,
                    protocol,
                    port.value,
                    true,
                );

                await this.execPostRouting(
                    group.destinationIp,
                    group.targetReverseIp,
                    protocol,
                    port.value,
                    true,
                );
            }
        }
    }

    async removeGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
        // iptables -D PREROUTING -t nat -d 212.193.55.91 -p tcp --dport $3 -j DNAT --to-dest 10.7.0.3:$3;
        // iptables -D POSTROUTING -t nat -d 10.7.0.3 -p tcp --dport $3 -j SNAT --to-source 10.7.0.1;

        for (const port of group.ports) {
            for (const protocol of port.protocols) {
                await this.execPreRouting(
                    group.targetIp,
                    group.destinationIp,
                    protocol,
                    port.value,
                    false,
                );

                await this.execPostRouting(
                    group.destinationIp,
                    group.targetReverseIp,
                    protocol,
                    port.value,
                    false,
                );
            }
        }
    }

    private async execPreRouting(
        targetIp: string,
        destinationIp: string,
        protocol: "udp" | "tcp",
        port: number,
        append: boolean,
    ): Promise<void> {
        await execPromise(
            `iptables -${append ? 'A' : 'D'} PREROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${port} -j DNAT --to-dest ${destinationIp}:${port}`,
        );
    }

    private async execPostRouting(
        targetIp: string,
        destinationIp: string,
        protocol: "udp" | "tcp",
        port: number,
        append: boolean,
    ): Promise<void> {
        await execPromise(
            `iptables -${append ? 'A' : 'D'} POSTROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${port} -j SNAT --to-source ${destinationIp}`,
        );
    }
}
