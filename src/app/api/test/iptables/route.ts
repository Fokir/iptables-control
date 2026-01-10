import { NextRequest, NextResponse } from "next/server";
import {
  addIptablesGroup,
  deleteGroup,
  getIptablesGroups,
  updateIptablesGroup,
} from "@/iptables/iptables-manager";
import { IptablesDatabaseGroupInterface } from "@/iptables/interfaces/iptables-database-group.interface";

// GET /api/test/iptables - Get all groups
export async function GET() {
  try {
    const groups = await getIptablesGroups();
    return NextResponse.json({ success: true, groups });
  } catch (error) {
    return NextResponse.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}

// POST /api/test/iptables - Add new group
export async function POST(request: NextRequest) {
  try {
    const group: IptablesDatabaseGroupInterface = await request.json();
    await addIptablesGroup(group);
    return NextResponse.json({ success: true, group });
  } catch (error) {
    return NextResponse.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}

// PUT /api/test/iptables - Update group
export async function PUT(request: NextRequest) {
  try {
    const group: IptablesDatabaseGroupInterface = await request.json();
    await updateIptablesGroup(group);
    return NextResponse.json({ success: true, group });
  } catch (error) {
    return NextResponse.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}

// DELETE /api/test/iptables - Delete group
export async function DELETE(request: NextRequest) {
  try {
    const group: IptablesDatabaseGroupInterface = await request.json();
    await deleteGroup(group);
    return NextResponse.json({ success: true });
  } catch (error) {
    return NextResponse.json(
      { success: false, error: String(error) },
      { status: 500 }
    );
  }
}
