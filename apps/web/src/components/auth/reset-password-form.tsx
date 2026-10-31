"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { useMutation } from "urql";

import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { graphql } from "@/gql";
import { noFormErrors, ruleErrors, toFormErrors, type FormErrors } from "@/lib/auth/form-errors";
import { lockedMessage, networkErrorMessage } from "@/lib/auth/messages";
import { checkPassword, checkRequired, PASSWORD_MIN } from "@/lib/auth/rules";
import { useResetOnHide } from "@/lib/auth/use-reset-on-hide";

import { FormAlert } from "./form-alert";

const ResetPasswordMutation = graphql(`
  mutation ResetPassword($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      ok
      errors {
        field
        code
      }
      lockedUntil
    }
  }
`);

export function ResetPasswordForm() {
  const router = useRouter();
  const [loginId, setLoginId] = useState("");
  const [resetCode, setResetCode] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [errors, setErrors] = useState<FormErrors>(noFormErrors);
  const [{ fetching }, resetPassword] = useMutation(ResetPasswordMutation);
  useResetOnHide(() => {
    setLoginId("");
    setResetCode("");
    setNewPassword("");
    setErrors(noFormErrors);
  });

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const invalid = ruleErrors([
      loginId.trim() === "" ? { field: "LOGIN_ID", code: "REQUIRED" } : null,
      checkRequired("RESET_CODE", resetCode),
      checkPassword("NEW_PASSWORD", newPassword),
    ]);
    if (invalid) {
      setErrors(invalid);
      return;
    }

    const result = await resetPassword({ input: { loginId, resetCode, newPassword } });
    const payload = result.data?.resetPassword;
    if (!payload) {
      setErrors({ fields: {}, form: networkErrorMessage });
      return;
    }
    if (!payload.ok) {
      const formErrors = toFormErrors(payload.errors);
      if (payload.errors.some((err) => err.code === "LOCKED")) {
        formErrors.form = lockedMessage("reset", payload.lockedUntil);
      }
      setErrors(formErrors);
      return;
    }
    router.replace("/login?reset=1");
  }

  return (
    <div className="flex flex-col gap-6">
      <p className="text-sm leading-relaxed text-muted-foreground">
        パスワードを忘れたときは、管理者に再設定コードを発行してもらってください。
      </p>
      <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-4">
        {errors.form && <FormAlert tone="error">{errors.form}</FormAlert>}
        <TextField
          label="ログイン ID"
          name="loginId"
          autoComplete="username"
          required
          value={loginId}
          onValueChange={setLoginId}
          error={errors.fields.LOGIN_ID}
        />
        <TextField
          label="再設定コード"
          name="resetCode"
          autoComplete="off"
          description="管理者から受け取ったコード"
          required
          value={resetCode}
          onValueChange={setResetCode}
          error={errors.fields.RESET_CODE}
        />
        <TextField
          label="新しいパスワード"
          name="newPassword"
          type="password"
          autoComplete="new-password"
          description={`${PASSWORD_MIN} 文字以上`}
          required
          value={newPassword}
          onValueChange={setNewPassword}
          error={errors.fields.NEW_PASSWORD}
        />
        <Button type="submit" size="lg" disabled={fetching}>
          パスワードを再設定する
        </Button>
      </form>
      <p className="text-sm">
        <Link href="/login" className="text-primary underline underline-offset-4">
          ログインに戻る
        </Link>
      </p>
    </div>
  );
}
