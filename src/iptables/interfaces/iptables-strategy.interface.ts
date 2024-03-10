import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

export interface IptablesStrategyInterface {
  addGroupRules(group: IptablesDatabaseGroupInterface): Promise<void>;
  removeGroupRules(group: IptablesDatabaseGroupInterface): Promise<void>;
}
