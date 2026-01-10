import {IptablesStrategyInterface} from "@/iptables/interfaces/iptables-strategy.interface";
import {IptablesDatabaseGroupInterface} from "@/iptables/interfaces/iptables-database-group.interface";
import {IptablesDatabasePortInterface} from "@/iptables/interfaces/iptables-database-port.interface";
import {exec} from "child_process";
import {promisify} from "util";
import {randomUUID} from "crypto";

const execPromise = promisify(exec);

interface ParsedRule {
  targetIp: string;
  destinationIp: string;
  destinationPort: number;
  protocol: "tcp" | "udp";
  portTarget: number;
}

interface ParsedPostRule {
  targetIp: string;
  sourceIp: string;
  port: number;
  protocol: "tcp" | "udp";
}

export class UnixStrategy implements IptablesStrategyInterface {
  async addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    // iptables -A PREROUTING -t nat -d 212.193.55.91 -p tcp --dport $3 -j DNAT --to-dest 10.7.0.3:$3;
    // iptables -A POSTROUTING -t nat -d 10.7.0.3 -p tcp --dport $3 -j SNAT --to-source 10.7.0.1;

    if (!group) return;

    for (const port of group.ports) {
      for (const protocol of port.protocols) {
        await this.execPreRouting(
            group.targetIp,
            group.destinationIp,
            protocol,
            port.value,
            port.valueTarget ?? port.value,
            true,
        );

        await this.execPostRouting(
            group.destinationIp,
            group.targetReverseIp,
            protocol,
            port.value,
            port.valueTarget ?? port.value,
            true,
        );
      }
    }
  }

  async removeGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    // iptables -D PREROUTING -t nat -d 212.193.55.91 -p tcp --dport $3 -j DNAT --to-dest 10.7.0.3:$3;
    // iptables -D POSTROUTING -t nat -d 10.7.0.3 -p tcp --dport $3 -j SNAT --to-source 10.7.0.1;

    if (!group) return;

    for (const port of group.ports) {
      for (const protocol of port.protocols) {
        await this.execPreRouting(
            group.targetIp,
            group.destinationIp,
            protocol,
            port.value,
            port.valueTarget ?? port.value,
            false,
        );

        await this.execPostRouting(
            group.destinationIp,
            group.targetReverseIp,
            protocol,
            port.value,
            port.valueTarget ?? port.value,
            false,
        );
      }
    }
  }

  async syncGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    if (!group) return;

    for (const port of group.ports) {
      for (const protocol of port.protocols) {
        const portTarget = port.valueTarget ?? port.value;

        const preRoutingExists = await this.checkPreRoutingExists(
            group.targetIp,
            group.destinationIp,
            protocol,
            port.value,
            portTarget,
        );

        const postRoutingExists = await this.checkPostRoutingExists(
            group.destinationIp,
            group.targetReverseIp,
            protocol,
            portTarget,
        );

        if (group.enabled) {
          if (!preRoutingExists) {
            await this.execPreRouting(
                group.targetIp,
                group.destinationIp,
                protocol,
                port.value,
                portTarget,
                true,
            );
          }

          if (!postRoutingExists) {
            await this.execPostRouting(
                group.destinationIp,
                group.targetReverseIp,
                protocol,
                port.value,
                portTarget,
                true,
            );
          }
        } else {
          if (preRoutingExists) {
            await this.execPreRouting(
                group.targetIp,
                group.destinationIp,
                protocol,
                port.value,
                portTarget,
                false,
            );
          }

          if (postRoutingExists) {
            await this.execPostRouting(
                group.destinationIp,
                group.targetReverseIp,
                protocol,
                port.value,
                portTarget,
                false,
            );
          }
        }
      }
    }
  }

  private async checkPreRoutingExists(
      targetIp: string,
      destinationIp: string,
      protocol: "udp" | "tcp",
      port: number,
      portTarget: number,
  ): Promise<boolean> {
    try {
      await execPromise(
          `iptables -C PREROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${port} -j DNAT --to-dest ${destinationIp}:${portTarget}`,
      );
      return true;
    } catch {
      return false;
    }
  }

  private async checkPostRoutingExists(
      targetIp: string,
      destinationIp: string,
      protocol: "udp" | "tcp",
      portTarget: number,
  ): Promise<boolean> {
    try {
      await execPromise(
          `iptables -C POSTROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${portTarget} -j SNAT --to-source ${destinationIp}`,
      );
      return true;
    } catch {
      return false;
    }
  }

  private async execPreRouting(
      targetIp: string,
      destinationIp: string,
      protocol: "udp" | "tcp",
      port: number,
      portTarget: number,
      append: boolean,
  ): Promise<void> {
    const command = `iptables -${append ? "A" : "D"} PREROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${port} -j DNAT --to-dest ${destinationIp}:${portTarget}`;
    try {
      await execPromise(command);
    } catch (error) {
      console.error(`Failed to execute PREROUTING command: ${command}`, error);
      throw error;
    }
  }

  private async execPostRouting(
      targetIp: string,
      destinationIp: string,
      protocol: "udp" | "tcp",
      port: number,
      portTarget: number,
      append: boolean,
  ): Promise<void> {
    const command = `iptables -${append ? "A" : "D"} POSTROUTING -t nat -d ${targetIp} -p ${protocol} --dport ${portTarget} -j SNAT --to-source ${destinationIp}`;
    try {
      await execPromise(command);
    } catch (error) {
      console.error(`Failed to execute POSTROUTING command: ${command}`, error);
      throw error;
    }
  }

  async scanSystemRules(): Promise<IptablesDatabaseGroupInterface[]> {
    const preRules = await this.parsePreRoutingRules();
    const postRules = await this.parsePostRoutingRules();

    const groupsMap = new Map<string, {
      targetIp: string;
      destinationIp: string;
      targetReverseIp: string;
      ports: Map<string, IptablesDatabasePortInterface>;
    }>();

    for (const rule of preRules) {
      const postRule = postRules.find(
          (p) =>
              p.targetIp === rule.destinationIp &&
              p.port === rule.portTarget &&
              p.protocol === rule.protocol,
      );

      if (!postRule) continue;

      const groupKey = `${rule.targetIp}-${rule.destinationIp}-${postRule.sourceIp}`;

      if (!groupsMap.has(groupKey)) {
        groupsMap.set(groupKey, {
          targetIp: rule.targetIp,
          destinationIp: rule.destinationIp,
          targetReverseIp: postRule.sourceIp,
          ports: new Map(),
        });
      }

      const group = groupsMap.get(groupKey)!;
      const portKey = `${rule.destinationPort}-${rule.portTarget}`;

      if (!group.ports.has(portKey)) {
        group.ports.set(portKey, {
          id: randomUUID(),
          value: rule.destinationPort,
          valueTarget: rule.portTarget,
          protocols: [rule.protocol],
        });
      } else {
        const existingPort = group.ports.get(portKey)!;
        if (!existingPort.protocols.includes(rule.protocol)) {
          existingPort.protocols.push(rule.protocol);
        }
      }
    }

    const groups: IptablesDatabaseGroupInterface[] = [];

    Array.from(groupsMap.values()).forEach((groupData) => {
      const ports = Array.from(groupData.ports.values());
      const portsList = ports.map((p) => p.value).join(", ");

      groups.push({
        id: randomUUID(),
        name: `<SYSTEM> [${portsList}]`,
        enabled: true,
        targetIp: groupData.targetIp,
        destinationIp: groupData.destinationIp,
        targetReverseIp: groupData.targetReverseIp,
        ports,
      });
    });

    return groups;
  }

  async saveRules(): Promise<void> {
    try {
      await execPromise("netfilter-persistent save");
    } catch {
      try {
        await execPromise("iptables-save > /etc/iptables/rules.v4");
      } catch {
        console.error("Failed to save iptables rules");
      }
    }
  }

  private async parsePreRoutingRules(): Promise<ParsedRule[]> {
    const rules: ParsedRule[] = [];

    try {
      const { stdout } = await execPromise("iptables-save -t nat");
      const lines = stdout.split("\n");

      for (const line of lines) {
        if (!line.startsWith("-A PREROUTING")) continue;
        if (!line.includes("-j DNAT")) continue;

        const dMatch = line.match(/-d\s+(\d+\.\d+\.\d+\.\d+)/);
        const pMatch = line.match(/-p\s+(tcp|udp)/);
        const dportMatch = line.match(/--dport\s+(\d+)/);
        const toDestMatch = line.match(/--to-destination\s+(\d+\.\d+\.\d+\.\d+):(\d+)/);

        if (dMatch && pMatch && dportMatch && toDestMatch) {
          rules.push({
            targetIp: dMatch[1],
            protocol: pMatch[1] as "tcp" | "udp",
            destinationPort: parseInt(dportMatch[1], 10),
            destinationIp: toDestMatch[1],
            portTarget: parseInt(toDestMatch[2], 10),
          });
        }
      }
    } catch {
      console.error("Failed to parse PREROUTING rules");
    }

    return rules;
  }

  private async parsePostRoutingRules(): Promise<ParsedPostRule[]> {
    const rules: ParsedPostRule[] = [];

    try {
      const { stdout } = await execPromise("iptables-save -t nat");
      const lines = stdout.split("\n");

      for (const line of lines) {
        if (!line.startsWith("-A POSTROUTING")) continue;
        if (!line.includes("-j SNAT")) continue;

        const dMatch = line.match(/-d\s+(\d+\.\d+\.\d+\.\d+)/);
        const pMatch = line.match(/-p\s+(tcp|udp)/);
        const dportMatch = line.match(/--dport\s+(\d+)/);
        const toSourceMatch = line.match(/--to-source\s+(\d+\.\d+\.\d+\.\d+)/);

        if (dMatch && pMatch && dportMatch && toSourceMatch) {
          rules.push({
            targetIp: dMatch[1],
            protocol: pMatch[1] as "tcp" | "udp",
            port: parseInt(dportMatch[1], 10),
            sourceIp: toSourceMatch[1],
          });
        }
      }
    } catch {
      console.error("Failed to parse POSTROUTING rules");
    }

    return rules;
  }
}
