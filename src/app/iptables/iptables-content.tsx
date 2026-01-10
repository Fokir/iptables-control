"use client";

import { IptablesDatabaseInterface } from "@/iptables/interfaces/iptables-database.interface";
import {
  Button,
  Table,
  TableBody,
  TableCell,
  TableColumn,
  TableHeader,
  TableRow,
} from "@nextui-org/react";
import { useState } from "react";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";
import IptablesEditModal from "@/app/iptables/iptables-edit-modal";
import { NetworkNodesDatabaseInterface } from "@/network-nodes/interfaces/network-nodes-database.interface";
import { TrashIconComponent } from "@/app/components/trash-icon.component";
import { EditIconComponent } from "@/app/components/edit-icon.component";
import { deleteGroup } from "@/iptables/iptables-manager";
import { useRouter } from "next/navigation";
import { generateId } from "@/utils/generate-id";

export type IptablesCardProps = {
  groups: IptablesDatabaseInterface["groups"];
  nodes: NetworkNodesDatabaseInterface["nodes"];
};

const getDefaultGroupValue: () => IptablesDatabaseGroupInterface = () => ({
  id: generateId(),
  name: "",
  enabled: true,
  targetIp: "",
  targetReverseIp: "",
  destinationIp: "",
  ports: [],
});

export default function IptablesContent({ groups, nodes }: IptablesCardProps) {
  const router = useRouter();

  const [isOpen, setIsOpen] = useState(false);

  const [editedGroup, setEditedGroup] = useState(getDefaultGroupValue());

  function addGroup(): void {
    setEditedGroup(getDefaultGroupValue());
    setIsOpen(true);
  }

  async function removeGroup(group: IptablesDatabaseGroupInterface) {
    await deleteGroup(group);

    router.refresh();
  }

  function onEditGroup(group: IptablesDatabaseGroupInterface): void {
    setEditedGroup({ ...group });
    setIsOpen(true);
  }

  const disabledGroupKeys = groups
    .filter((group) => !group.enabled)
    .map((group) => group.id);

  return (
    <div>
      <div className="w-full flex items-center">
        <h1 className="text-3xl font-bold">Правила маршуризации</h1>

        <Button
          className="ml-auto"
          color="primary"
          variant="flat"
          onClick={addGroup}
        >
          Добавить группу
        </Button>
      </div>

      <div className="w-full mt-12">
        <Table
          aria-label="Example static collection table"
          disabledKeys={disabledGroupKeys}
        >
          <TableHeader>
            <TableColumn>Название</TableColumn>
            <TableColumn>Маршруты</TableColumn>
            <TableColumn>Порты</TableColumn>
            <TableColumn width={100}>Действия</TableColumn>
          </TableHeader>

          <TableBody emptyContent={"Правил маршрутизации не найдено"}>
            {groups.map((group) => (
              <TableRow key={group.id}>
                <TableCell>{group.name}</TableCell>

                <TableCell>
                  <div className="text-xs">
                    Входной интерфейс: {group.targetIp}
                  </div>
                  <div className="text-xs">
                    Выходной интерфейс: {group.targetReverseIp}
                  </div>
                  <div className="text-xs">
                    Интерфейс назначения: {group.destinationIp}
                  </div>
                </TableCell>

                <TableCell>
                  {group.ports.map((port) => (
                    <div key={port.id} className="text-xs">
                      {port.value}:{port.valueTarget ?? port.value} <span className="text-yellow-300">[{port.protocols.join(", ")}]</span>
                    </div>
                  ))}
                </TableCell>

                <TableCell>
                  <div className="flex gap-2">
                    <Button
                      isIconOnly
                      color="primary"
                      size="lg"
                      aria-label="Изменить"
                      onClick={() => onEditGroup(group)}
                    >
                      <EditIconComponent maxWidth={16} />
                    </Button>

                    <Button
                      isIconOnly
                      color="danger"
                      size="lg"
                      aria-label="Удалить"
                      onClick={() => removeGroup(group)}
                    >
                      <TrashIconComponent maxWidth={16} />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <IptablesEditModal
        group={editedGroup}
        nodes={nodes}
        isOpen={isOpen}
        onClose={() => setIsOpen(false)}
      ></IptablesEditModal>
    </div>
  );
}
