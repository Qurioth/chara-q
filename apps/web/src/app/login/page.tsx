import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { Suspense } from "react";

import { AuthPage } from "@/components/auth/auth-page";
import { LoginForm } from "@/components/auth/login-form";
import { safeNext } from "@/lib/auth/safe-next";
import { getViewer } from "@/lib/auth/viewer";

export const metadata: Metadata = { title: "ログイン" };

export default function LoginPage({ searchParams }: PageProps<"/login">) {
  return (
    <AuthPage title="ログイン">
      {/* ログイン状態と ?next= はリクエストごとに変わるため、<Suspense> の中で読む（cacheComponents） */}
      <Suspense>
        <LoginGate searchParams={searchParams} />
      </Suspense>
    </AuthPage>
  );
}

async function LoginGate({ searchParams }: Pick<PageProps<"/login">, "searchParams">) {
  const params = await searchParams;
  const next = safeNext(typeof params.next === "string" ? params.next : null);
  if (await getViewer()) {
    redirect(next);
  }
  return <LoginForm next={next} reset={params.reset === "1"} />;
}
