export interface IptablesDatabasePortInterface {
  id: string;
  value: number;
  protocols: Array<"udp" | "tcp">;
}
