import { IptablesStrategyInterface } from "@/iptables/interfaces/iptables-strategy.interface";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

export class WindowsStrategy implements IptablesStrategyInterface {
  async addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    console.log('add group rule', group);
  }

  async removeGroupRules(
    group: IptablesDatabaseGroupInterface,
  ): Promise<void> {
    console.log('remove group rule', group);
  }

  async syncGroupRules(group: IptablesDatabaseGroupInterface): Promise<void> {
    console.log('sync group rule', group);
  }

  async scanSystemRules(): Promise<IptablesDatabaseGroupInterface[]> {
    console.log('scan system rules (not implemented for Windows)');
    return [];
  }

  async saveRules(): Promise<void> {
    console.log('save rules (not implemented for Windows)');
  }
}
