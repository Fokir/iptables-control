import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

export const config = {
  matcher: ["/", "/iptables", "/network-nodes"],
};

export async function middleware(request: NextRequest) {
  const basicAuth = request.headers.get("authorization");
  const url = request.nextUrl;

  if (basicAuth) {
    try {
      const parts = basicAuth.split(" ");
      if (parts.length !== 2 || parts[0] !== "Basic") {
        throw new Error("Invalid auth format");
      }

      const authValue = parts[1];
      const decoded = atob(authValue);
      const colonIndex = decoded.indexOf(":");

      if (colonIndex === -1) {
        throw new Error("Invalid credentials format");
      }

      const user = decoded.substring(0, colonIndex);
      const pwd = decoded.substring(colonIndex + 1);

      const validUser = process.env.BASIC_AUTH_USER;
      const validPassWord = process.env.BASIC_AUTH_PASSWORD;

      if (validUser && validPassWord && user === validUser && pwd === validPassWord) {
        return NextResponse.next();
      }
    } catch {
      // Invalid auth header format, fall through to auth page
    }
  }

  url.pathname = "/auth";

  return NextResponse.rewrite(url);
}
