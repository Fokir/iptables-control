"use client";

import { useRouter } from "next/navigation";
import { Button, Input } from "@nextui-org/react";
import { addNetworkNode } from "@/network-nodes/network-nodes-manager";
import { useState } from "react";
import { validateNode } from "@/network-nodes/functions/validate-node";
import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";

export default function NetworkNodeForm() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [ip, setIp] = useState("");

  async function saveNode() {
    const node: NetworkNodeItemDatabaseInterface = {
      name,
      ip,
    };

    if (validateNode(node)) {
      await addNetworkNode(node);

      router.refresh();
    }
  }

  return (
    <div className="w-full flex gap-4 items-center">
      <Input
        type="text"
        label="Имя узла"
        size="sm"
        onValueChange={setName}
        value={name}
      />

      <Input
        type="text"
        label="IP адрес"
        size="sm"
        onValueChange={setIp}
        value={ip}
      />

      <Button
        className="w-[250px]"
        color="primary"
        variant="bordered"
        onClick={saveNode}
        size="lg"
      >
        Сохранить
      </Button>
    </div>
  );
}
