import {IptablesDatabaseGroupInterface} from "@/iptables/interfaces/iptables-database-group.interface";
import {IptablesStrategyInterface} from "@/iptables/interfaces/iptables-strategy.interface";
import {WindowsStrategy} from "@/iptables/_strategies/windows.strategy";
import {UnixStrategy} from "@/iptables/_strategies/unix.strategy";
import {DarwinStrategy} from "@/iptables/_strategies/darwin.strategy";

const strategyIptables: IptablesStrategyInterface = (() => {
    switch (process.platform) {
        case "darwin":
            return new DarwinStrategy();
        case "win32":
            return new WindowsStrategy();
        default:
            return new UnixStrategy();
    }
})();

export async function execAppendGroup(group: IptablesDatabaseGroupInterface) {
    await strategyIptables.addGroupRules(group);
}

export async function execRemoveGroup(group: IptablesDatabaseGroupInterface) {
    await strategyIptables.removeGroupRules(group);
}

export async function execSyncGroup(group: IptablesDatabaseGroupInterface) {
    await strategyIptables.syncGroupRules(group);
}

export async function scanSystemRules() {
    return strategyIptables.scanSystemRules();
}

export async function saveRules() {
    await strategyIptables.saveRules();
}
