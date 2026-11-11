import type { Metadata } from "next";

import { AuthPage } from "@/components/auth/auth-page";
import { ResetPasswordForm } from "@/components/auth/reset-password-form";

export const metadata: Metadata = { title: "パスワードの再設定" };

export default function ResetPasswordPage() {
  return (
    <AuthPage title="パスワードの再設定">
      <ResetPasswordForm />
    </AuthPage>
  );
}
