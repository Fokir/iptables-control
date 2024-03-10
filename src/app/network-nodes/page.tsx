import { getNetworkNodes } from "@/network-nodes/network-nodes-manager";
import { Fragment } from "react";
import NetworkNodesTable from "@/app/network-nodes/network-nodes-table";
import NetworkNodeForm from "@/app/network-nodes/network-node-form";

export const dynamic = 'force-dynamic';
export default async function networkNodesPage() {
  const nodes = await getNetworkNodes();

  return (
    <Fragment>
      <div className="w-full flex items-center">
        <h1 className="text-3xl font-bold">Узлы маршуризации</h1>
      </div>

      <div className="w-full flex items-center mt-6">
        <NetworkNodeForm></NetworkNodeForm>
      </div>

      <div className="w-full mt-12">
        <NetworkNodesTable nodes={nodes}></NetworkNodesTable>
      </div>
    </Fragment>
  );
}
