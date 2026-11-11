"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { useMutation } from "urql";

import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { graphql } from "@/gql";
import { noFormErrors, ruleErrors, toFormErrors, type FormErrors } from "@/lib/auth/form-errors";
import { networkErrorMessage } from "@/lib/auth/messages";
import {
  checkDisplayName,
  checkLoginId,
  checkPassword,
  checkRequired,
  DISPLAY_NAME_MAX,
  LOGIN_ID_MAX,
  LOGIN_ID_MIN,
  PASSWORD_MIN,
} from "@/lib/auth/rules";
import { useResetOnHide } from "@/lib/auth/use-reset-on-hide";

import { FormAlert } from "./form-alert";

const RegisterMutation = graphql(`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      viewer {
        id
      }
      errors {
        field
        code
      }
    }
  }
`);

export function RegisterForm() {
  const router = useRouter();
  const [invitationCode, setInvitationCode] = useState("");
  const [loginId, setLoginId] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [errors, setErrors] = useState<FormErrors>(noFormErrors);
  const [{ fetching }, register] = useMutation(RegisterMutation);
  useResetOnHide(() => {
    setInvitationCode("");
    setLoginId("");
    setPassword("");
    setDisplayName("");
    setErrors(noFormErrors);
  });

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const invalid = ruleErrors([
      checkRequired("INVITATION_CODE", invitationCode),
      checkLoginId(loginId),
      checkPassword("PASSWORD", password),
      checkDisplayName(displayName),
    ]);
    if (invalid) {
      setErrors(invalid);
      return;
    }

    const result = await register({ input: { invitationCode, loginId, password, displayName } });
    const payload = result.data?.register;
    if (!payload) {
      setErrors({ fields: {}, form: networkErrorMessage });
      return;
    }
    if (payload.errors.length > 0) {
      setErrors(toFormErrors(payload.errors));
      return;
    }
    router.replace("/");
    router.refresh();
  }

  return (
    <div className="flex flex-col gap-6">
      <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-4">
        {errors.form && <FormAlert tone="error">{errors.form}</FormAlert>}
        <TextField
          label="招待コード"
          name="invitationCode"
          autoComplete="off"
          description="管理者から受け取ったコード"
          required
          value={invitationCode}
          onValueChange={setInvitationCode}
          error={errors.fields.INVITATION_CODE}
        />
        <TextField
          label="ログイン ID"
          name="loginId"
          autoComplete="username"
          description={`半角英数字・_・- の ${LOGIN_ID_MIN}〜${LOGIN_ID_MAX} 文字。他の人には表示されません`}
          required
          maxLength={LOGIN_ID_MAX}
          value={loginId}
          onValueChange={setLoginId}
          error={errors.fields.LOGIN_ID}
        />
        <TextField
          label="パスワード"
          name="password"
          type="password"
          autoComplete="new-password"
          description={`${PASSWORD_MIN} 文字以上`}
          required
          value={password}
          onValueChange={setPassword}
          error={errors.fields.PASSWORD}
        />
        <TextField
          label="表示名"
          name="displayName"
          autoComplete="nickname"
          description={`他の人にも表示されます（${DISPLAY_NAME_MAX} 文字まで）。後から変えられます`}
          required
          value={displayName}
          onValueChange={setDisplayName}
          error={errors.fields.DISPLAY_NAME}
        />
        <Button type="submit" size="lg" disabled={fetching}>
          登録する
        </Button>
      </form>
      <p className="text-sm">
        登録済みの方は{" "}
        <Link href="/login" className="text-primary underline underline-offset-4">
          ログイン
        </Link>
      </p>
    </div>
  );
}
