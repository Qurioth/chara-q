import type { Metadata } from "next";
import { Suspense } from "react";

import { InvitationIssuer } from "@/components/auth/invitation-issuer";
import { requireAdmin } from "@/lib/auth/viewer";

export const metadata: Metadata = { title: "招待コードの発行", robots: { index: false } };

export default function InvitationsPage() {
  return (
    <div className="mx-auto flex w-full max-w-xl flex-col gap-8 px-4 py-8">
      {/* 見出しも含めて、管理者だと確かめてから描く */}
      <Suspense>
        <AdminOnly />
      </Suspense>
    </div>
  );
}

// 管理者でなければ 404、ログインしていなければログイン画面へ移る
async function AdminOnly() {
  await requireAdmin("/admin/invitations");
  return (
    <>
      <h1 className="text-2xl font-bold">招待コードの発行</h1>
      <InvitationIssuer />
    </>
  );
}
