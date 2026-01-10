import {
  Autocomplete,
  AutocompleteItem,
  Button,
  Checkbox,
  Divider,
  Input,
  Modal,
  ModalBody,
  ModalContent,
  ModalFooter,
  ModalHeader, Snippet,
} from "@nextui-org/react";
import {Fragment, useEffect, useState} from "react";
import {IptablesDatabaseGroupInterface} from "@/iptables/interfaces/iptables-database-group.interface";
import {validateGroup} from "@/iptables/functions/validate-group";
import {useRouter} from "next/navigation";
import {NetworkNodesDatabaseInterface} from "@/network-nodes/interfaces/network-nodes-database.interface";
import IptablesEditPort from "@/app/iptables/iptables-edit-port";
import {updateOrCreateGroup} from "@/iptables/iptables-manager";
import {IptablesDatabasePortInterface} from "@/iptables/interfaces/iptables-database-port.interface";
import {generateId} from "@/utils/generate-id";

export type IptablesEditCardProps = {
  group: IptablesDatabaseGroupInterface;
  nodes: NetworkNodesDatabaseInterface["nodes"];
  isOpen: boolean;
  onClose: () => void;
};
export default function IptablesEditModal({
                                            group: groupProps,
                                            nodes,
                                            isOpen,
                                            onClose,
                                          }: IptablesEditCardProps) {
  const router = useRouter();

  const [isPending, setIsPending] = useState(false);

  const [name, setName] = useState("");
  const [targetIp, setTargetIp] = useState("");
  const [targetReverseIp, setTargetReverseIp] = useState("");
  const [destinationIp, setDestinationIp] = useState("");
  const [enabled, setEnabled] = useState(true);

  const [ports, setPorts] = useState<IptablesDatabasePortInterface[]>([]);

  const [hasErrors, setHasErrors] = useState<boolean>(false);

  useEffect(() => {
    setName(groupProps.name);
    setTargetIp(groupProps.targetIp);
    setTargetReverseIp(groupProps.targetReverseIp);
    setDestinationIp(groupProps.destinationIp);
    setPorts(groupProps.ports.map((port) => ({...port})));
    setEnabled(groupProps.enabled);
  }, [groupProps]);

  async function saveCard() {
    const group: IptablesDatabaseGroupInterface = {
      id: groupProps.id,
      name,
      targetIp,
      targetReverseIp,
      destinationIp,
      ports,
      enabled,
    };

    if (validateGroup(group)) {
      setHasErrors(false);
      setIsPending(true);

      await updateOrCreateGroup(group);

      router.refresh();

      setIsPending(false);
      onClose();
    } else {
      setHasErrors(true);
    }
  }

  function addPort(): void {
    setPorts((ports) => [
      ...ports,
      {
        id: generateId(),
        protocols: ["udp", "tcp"],
        value: 0,
        valueTarget: 0,
      },
    ]);
  }

  function removePort(deleteIndex: number): void {
    setPorts((ports) => ports.filter((port, index) => deleteIndex !== index));
  }

  function updatePort(
      portIndex: number,
      newPort: IptablesDatabasePortInterface,
  ): void {
    setPorts((ports) =>
        ports.map((port, index) => {
          if (portIndex === index) {
            return newPort;
          }
          return port;
        }),
    );
  }

  function closeModal(): void {
    if (typeof onClose === "function") onClose();
  }

  return (
      <Modal isOpen={isOpen} onOpenChange={(status) => !status && onClose()} size="xl">
        <ModalContent>
          {(onClose) => (
              <>
                <ModalHeader className="flex flex-col gap-1">
                  Правило маршрутизации
                </ModalHeader>

                <ModalBody>
                  <Input
                      className="w-full"
                      type="text"
                      label="Название группы"
                      size="sm"
                      value={name}
                      onValueChange={setName}
                  />

                  <Checkbox onValueChange={setEnabled} isSelected={enabled}>
                    Правила активированы
                  </Checkbox>

                  <Autocomplete
                      allowsCustomValue
                      label="Входной интерфейс"
                      variant="bordered"
                      className="w-full"
                      defaultItems={nodes}
                      onInputChange={(value) => setTargetIp(value)}
                      inputValue={targetIp}
                  >
                    {(item) => (
                        <AutocompleteItem key={item.ip} description={item.name}>
                          {item.ip}
                        </AutocompleteItem>
                    )}
                  </Autocomplete>

                  <Autocomplete
                      allowsCustomValue
                      label="Выходной интерфейс"
                      variant="bordered"
                      className="w-full"
                      defaultItems={nodes}
                      onInputChange={(value) => setTargetReverseIp(value)}
                      inputValue={targetReverseIp}
                  >
                    {(item) => (
                        <AutocompleteItem key={item.ip} description={item.name}>
                          {item.ip}
                        </AutocompleteItem>
                    )}
                  </Autocomplete>

                  <Autocomplete
                      allowsCustomValue
                      label="Интерфейс назначения"
                      variant="bordered"
                      className="w-full"
                      defaultItems={nodes}
                      onInputChange={(value) => setDestinationIp(value)}
                      inputValue={destinationIp}
                  >
                    {(item) => (
                        <AutocompleteItem key={item.ip} description={item.name}>
                          {item.ip}
                        </AutocompleteItem>
                    )}
                  </Autocomplete>

                  <div className="w-full">
                    {ports.map((port, index) => (
                        <Fragment key={port.id}>
                          <Divider/>
                          <div className="mt-2">
                            <IptablesEditPort
                                port={port}
                                onDelete={() => removePort(index)}
                                onChange={(port) => updatePort(index, port)}
                            />
                          </div>
                        </Fragment>
                    ))}

                    <Button
                        className="mt-6"
                        color="primary"
                        variant="flat"
                        onPress={addPort}
                    >
                      Добавить порт
                    </Button>
                  </div>

                  {hasErrors && <div className="w-full">
                    <span className="text-danger-300">
                      Данные группы содержат ошибки
                    </span>
                  </div>}
                </ModalBody>

                <ModalFooter>
                  <Button color="danger" variant="light" onPress={closeModal}>
                    Отмена
                  </Button>

                  <Button
                      color="primary"
                      variant="bordered"
                      onPress={saveCard}
                      disabled={isPending}
                      isLoading={isPending}
                  >
                    Сохранить
                  </Button>
                </ModalFooter>
              </>
          )}
        </ModalContent>
      </Modal>
  );
}
