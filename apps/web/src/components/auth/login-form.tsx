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
import { useResetOnHide } from "@/lib/auth/use-reset-on-hide";

import { FormAlert } from "./form-alert";

const LoginMutation = graphql(`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      viewer {
        id
      }
      errors {
        field
        code
      }
      lockedUntil
    }
  }
`);

type LoginFormProps = {
  /** ログインの後に戻る画面（safeNext で検証済み） */
  next: string;
  /** パスワードを再設定した直後か */
  reset?: boolean;
};

export function LoginForm({ next, reset = false }: LoginFormProps) {
  const router = useRouter();
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [errors, setErrors] = useState<FormErrors>(noFormErrors);
  const [{ fetching }, login] = useMutation(LoginMutation);
  useResetOnHide(() => {
    setLoginId("");
    setPassword("");
    setErrors(noFormErrors);
  });

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    // ログイン ID の規則は確かめない（規則に合わなくても、サーバーは同じ失敗の文言を返す）
    const invalid = ruleErrors([
      loginId.trim() === "" ? { field: "LOGIN_ID", code: "REQUIRED" } : null,
      password === "" ? { field: "PASSWORD", code: "REQUIRED" } : null,
    ]);
    if (invalid) {
      setErrors(invalid);
      return;
    }

    const result = await login({ input: { loginId, password } });
    const payload = result.data?.login;
    if (!payload) {
      setErrors({ fields: {}, form: networkErrorMessage });
      return;
    }
    if (payload.errors.length > 0) {
      const formErrors = toFormErrors(payload.errors);
      if (payload.errors.some((err) => err.code === "LOCKED")) {
        formErrors.form = lockedMessage("login", payload.lockedUntil);
      }
      setErrors(formErrors);
      setPassword("");
      return;
    }
    // ヘッダー（Server Components）も描き直す
    router.replace(next);
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      {reset && !errors.form && (
        <FormAlert tone="success">パスワードを再設定しました。新しいパスワードでログインしてください</FormAlert>
      )}
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
          label="パスワード"
          name="password"
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onValueChange={setPassword}
          error={errors.fields.PASSWORD}
        />
        <Button type="submit" size="lg" disabled={fetching}>
          ログイン
        </Button>
      </form>
      <ul className="flex flex-col gap-2 text-sm">
        <li>
          招待コードをお持ちの方は{" "}
          <Link href="/register" className="text-primary underline underline-offset-4">
            ユーザー登録
          </Link>
        </li>
        <li>
          パスワードを忘れた方は{" "}
          <Link href="/reset-password" className="text-primary underline underline-offset-4">
            パスワードの再設定
          </Link>
        </li>
      </ul>
    </div>
  );
}
