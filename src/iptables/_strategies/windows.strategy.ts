import { IptablesStrategyInterface } from "@/iptables/interfaces/iptables-strategy.interface";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

export class WindowsStrategy implements IptablesStrategyInterface {
  async addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {}

  async removeGroupRules(
    group: IptablesDatabaseGroupInterface,
  ): Promise<void> {}
}
