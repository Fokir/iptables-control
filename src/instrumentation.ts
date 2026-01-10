export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { getDatabaseConnection } = await import(
      "@/iptables/iptables-database"
    );

    const { execSyncGroup, scanSystemRules, saveRules } = await import("@/iptables/iptables-exec");

    const connection = await getDatabaseConnection();
    await connection.read();

    const systemRules = await scanSystemRules();
    let hasChanges = false;

    for (const systemGroup of systemRules) {
      const existsInDb = connection.data.groups.some(
        (dbGroup) =>
          dbGroup.targetIp === systemGroup.targetIp &&
          dbGroup.destinationIp === systemGroup.destinationIp &&
          dbGroup.targetReverseIp === systemGroup.targetReverseIp &&
          systemGroup.ports.every((sysPort) =>
            dbGroup.ports.some(
              (dbPort) =>
                dbPort.value === sysPort.value &&
                dbPort.valueTarget === sysPort.valueTarget &&
                sysPort.protocols.every((p) => dbPort.protocols.includes(p)),
            ),
          ),
      );

      if (!existsInDb) {
        connection.data.groups.push(systemGroup);
        hasChanges = true;
      }
    }

    if (hasChanges) {
      await connection.write();
    }

    for (const group of connection.data.groups) {
      await execSyncGroup(group);
    }

    if (hasChanges) {
      await saveRules();
    }
  }
}
