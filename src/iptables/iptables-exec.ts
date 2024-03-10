import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";
import { IptablesStrategyInterface } from "@/iptables/interfaces/iptables-strategy.interface";
import { WindowsStrategy } from "@/iptables/_strategies/windows.strategy";
import { UnixStrategy } from "@/iptables/_strategies/unix.strategy";

const strategyIptables: IptablesStrategyInterface =
  process.platform === "win32" ? new WindowsStrategy() : new UnixStrategy();

export async function execAppendGroup(group: IptablesDatabaseGroupInterface) {
  await strategyIptables.addGroupRules(group);
}

export async function execRemoveGroup(group: IptablesDatabaseGroupInterface) {
  await strategyIptables.removeGroupRules(group);
}
