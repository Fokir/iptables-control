import { IptablesStrategyInterface } from "@/iptables/interfaces/iptables-strategy.interface";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

export class DarwinStrategy implements IptablesStrategyInterface {
  async addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    console.log('add group rule', group);
  }

  async removeGroupRules(
      group: IptablesDatabaseGroupInterface,
  ): Promise<void> {
    console.log('remove group rule', group);
  }
}
