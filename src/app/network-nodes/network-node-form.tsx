"use client";

import { useRouter } from "next/navigation";
import { Button, Input } from "@nextui-org/react";
import { addNetworkNode } from "@/network-nodes/network-nodes-manager";
import { useState } from "react";
import { validateNode, validateIp } from "@/network-nodes/functions/validate-node";
import { NetworkNodeItemDatabaseInterface } from "@/network-nodes/interfaces/network-node-item-database.interface";

export default function NetworkNodeForm() {
  const router = useRouter();

  const [name, setName] = useState("");
  const [ip, setIp] = useState("");
  const [isPending, setIsPending] = useState(false);
  const [error, setError] = useState("");

  async function saveNode() {
    const node: NetworkNodeItemDatabaseInterface = {
      name,
      ip,
    };

    if (!name.trim()) {
      setError("Введите имя узла");
      return;
    }

    if (!ip.trim()) {
      setError("Введите IP адрес");
      return;
    }

    if (!validateIp(ip)) {
      setError("Неверный формат IP адреса");
      return;
    }

    if (validateNode(node)) {
      setError("");
      setIsPending(true);

      await addNetworkNode(node);

      setName("");
      setIp("");
      setIsPending(false);

      router.refresh();
    }
  }

  return (
    <div className="w-full">
      <div className="w-full flex gap-4 items-center">
        <Input
          type="text"
          label="Имя узла"
          size="sm"
          onValueChange={(value) => { setName(value); setError(""); }}
          value={name}
          isInvalid={!!error && !name.trim()}
        />

        <Input
          type="text"
          label="IP адрес"
          size="sm"
          onValueChange={(value) => { setIp(value); setError(""); }}
          value={ip}
          isInvalid={!!error && (!ip.trim() || !validateIp(ip))}
        />

        <Button
          className="w-[250px]"
          color="primary"
          variant="bordered"
          onClick={saveNode}
          size="lg"
          isLoading={isPending}
          disabled={isPending}
        >
          Сохранить
        </Button>
      </div>

      {error && (
        <div className="mt-2 text-danger-400 text-sm">{error}</div>
      )}
    </div>
  );
}
