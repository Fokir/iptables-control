"use client";

import {
  Link,
  Navbar,
  NavbarBrand,
  NavbarContent,
  NavbarItem,
} from "@nextui-org/react";

import { LogoComponent } from "@/app/components/logo.component";
import { usePathname } from "next/navigation";

export function HeaderComponent() {
  const pathname = usePathname();

  return (
    <Navbar>
      <NavbarBrand>
        <LogoComponent maxWidth={24} />

        <p className="font-bold text-inherit ml-2">System control</p>
      </NavbarBrand>

      <NavbarContent className="hidden sm:flex gap-4" justify="center">
        <NavbarItem isActive={pathname === "/iptables"}>
          <Link color="foreground" href="/iptables">
            Маршрутизация
          </Link>
        </NavbarItem>

        <NavbarItem isActive={pathname === "/network-nodes"}>
          <Link color="foreground" href="/network-nodes" aria-current="page">
            Узлы маршрутизации
          </Link>
        </NavbarItem>
      </NavbarContent>
    </Navbar>
  );
}
