import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

export const config = {
  matcher: ["/", "/iptables", "/network-nodes"],
};

export async function middleware(request: NextRequest) {
  const basicAuth = request.headers.get("authorization");
  const url = request.nextUrl;

  if (basicAuth) {
    const authValue = basicAuth.split(" ")[1];
    const [user, pwd] = atob(authValue).split(":");

    const validUser = process.env.BASIC_AUTH_USER;
    const validPassWord = process.env.BASIC_AUTH_PASSWORD;

    if (user === validUser && pwd === validPassWord) {
      return NextResponse.next();
    }
  }

  url.pathname = "/auth";

  return NextResponse.rewrite(url);
}
