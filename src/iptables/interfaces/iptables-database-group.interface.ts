import { IptablesDatabasePortInterface } from "@/iptables/interfaces/iptables-database-port.interface";

export interface IptablesDatabaseGroupInterface {
  id: string;
  name: string;
  enabled: boolean;
  targetIp: string;
  targetReverseIp: string;
  destinationIp: string;
  ports: IptablesDatabasePortInterface[];
}
