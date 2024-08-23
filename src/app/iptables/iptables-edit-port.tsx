import {Button, Input, Select, SelectItem} from "@nextui-org/react";
import {useEffect, useState} from "react";
import {IptablesDatabasePortInterface} from "@/iptables/interfaces/iptables-database-port.interface";
import {TrashIconComponent} from "@/app/components/trash-icon.component";

export type IptablesEditRuleProps = {
  port: IptablesDatabasePortInterface;
  onDelete: () => void;
  onChange: (port: IptablesDatabasePortInterface) => void;
};
export default function IptablesEditPort({
                                           port: portProps,
                                           onDelete,
                                           onChange,
                                         }: IptablesEditRuleProps) {
  const [port, setPort] = useState(portProps.value);
  const [portTarget, setPortTarget] = useState(portProps.valueTarget);
  const [protocols, setProtocols] = useState([...portProps.protocols]);

  useEffect(() => {
    if (typeof onChange === "function")
      onChange({
        id: portProps.id,
        value: port,
        valueTarget: portTarget ?? port,
        protocols: [...protocols],
      });
  }, [port, portTarget, protocols]);

  function onDeleteHandler(): void {
    if (typeof onDelete === "function") onDelete();
  }

  return (
      <div className="flex gap-2">
        <Input
            className="mb-2"
            type="number"
            label="Порт"
            size="sm"
            value={port.toString()}
            onValueChange={(value) => setPort(value ? parseInt(value) : 0)}
        />

        <Input
            className="mb-2"
            type="number"
            label="Порт назначения"
            size="sm"
            value={(portTarget ?? port).toString()}
            onValueChange={(value) => setPortTarget(value ? parseInt(value) : 0)}
        />

        <Select
            size="sm"
            label="Протоколы"
            selectionMode="multiple"
            selectedKeys={protocols}
            onSelectionChange={(value) =>
                setProtocols(Array.from(value) as Array<"udp" | "tcp">)
            }
        >
          <SelectItem key="udp" value="udp">
            UDP
          </SelectItem>
          <SelectItem key="tcp" value="tcp">
            TCP
          </SelectItem>
        </Select>

        <Button
            isIconOnly
            color="danger"
            size="lg"
            aria-label="Удалить"
            onPress={onDeleteHandler}
        >
          <TrashIconComponent maxWidth={16}/>
        </Button>
      </div>
  );
}
