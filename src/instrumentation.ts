export async function register() {
  if (process.env.NEXT_RUNTIME === "nodejs") {
    const { getDatabaseConnection } = await import(
      "@/iptables/iptables-database"
    );

    const { execAppendGroup } = await import("@/iptables/iptables-exec");

    const connection = await getDatabaseConnection();
    await connection.read();

    for (const group of connection.data.groups) {
      await execAppendGroup(group);
    }
  }
}
