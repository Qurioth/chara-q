import type { Metadata } from "next";
import { Suspense } from "react";

import { AccountSettings } from "@/components/auth/account-settings";
import { requireViewer } from "@/lib/auth/viewer";

export const metadata: Metadata = { title: "アカウントの設定", robots: { index: false } };

export default function AccountPage() {
  return (
    <div className="mx-auto flex w-full max-w-xl flex-col gap-8 px-4 py-8">
      <h1 className="text-2xl font-bold">アカウントの設定</h1>
      <Suspense>
        <AccountContent />
      </Suspense>
    </div>
  );
}

async function AccountContent() {
  const viewer = await requireViewer("/account");
  return <AccountSettings loginId={viewer.loginId} displayName={viewer.displayName} />;
}
