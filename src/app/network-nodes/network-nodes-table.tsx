"use client";

import {
  Button,
  Table,
  TableBody,
  TableCell,
  TableColumn,
  TableHeader,
  TableRow,
} from "@nextui-org/react";
import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";
import { NetworkNodesDatabaseInterface } from "@/network-nodes/interfaces/network-nodes-database.interface";
import { deleteNetworkNode } from "@/network-nodes/network-nodes-manager";
import { useRouter } from "next/navigation";
import { TrashIconComponent } from "@/app/components/trash-icon.component";

export type NetworkNodesTableProps = {
  nodes: NetworkNodesDatabaseInterface["nodes"];
};

export default function NetworkNodesTable({ nodes }: NetworkNodesTableProps) {
  const router = useRouter();

  async function removeNode(node: NetworkNodeItemDatabaseInterface) {
    await deleteNetworkNode(node);

    router.refresh();
  }

  return (
    <Table aria-label="Example static collection table">
      <TableHeader>
        <TableColumn>Название</TableColumn>
        <TableColumn>IP адрес</TableColumn>
        <TableColumn width={100}>Действия</TableColumn>
      </TableHeader>

      <TableBody emptyContent={"Узлов маршрутизации не найдено"}>
        {nodes.map((node) => (
          <TableRow key={node.ip}>
            <TableCell>{node.name}</TableCell>
            <TableCell>{node.ip}</TableCell>
            <TableCell>
              <Button
                isIconOnly
                color="danger"
                size="lg"
                aria-label="Удалить"
                onClick={() => removeNode(node)}
              >
                <TrashIconComponent maxWidth={16} />
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
