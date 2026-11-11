import type { Metadata } from "next";
import { redirect } from "next/navigation";
import { Suspense } from "react";

import { AuthPage } from "@/components/auth/auth-page";
import { RegisterForm } from "@/components/auth/register-form";
import { getViewer } from "@/lib/auth/viewer";

export const metadata: Metadata = { title: "ユーザー登録" };

export default function RegisterPage() {
  return (
    <AuthPage title="ユーザー登録">
      <Suspense>
        <RegisterGate />
      </Suspense>
    </AuthPage>
  );
}

async function RegisterGate() {
  if (await getViewer()) {
    redirect("/");
  }
  return <RegisterForm />;
}
