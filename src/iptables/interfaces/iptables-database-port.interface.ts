export interface IptablesDatabasePortInterface {
  id: string;
  value: number;
  valueTarget: number;
  protocols: Array<"udp" | "tcp">;
}
