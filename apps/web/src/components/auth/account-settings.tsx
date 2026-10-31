"use client";

import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { useMutation } from "urql";

import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { graphql } from "@/gql";
import {
  isUnauthenticated,
  noFormErrors,
  ruleErrors,
  toFormErrors,
  type FormErrors,
} from "@/lib/auth/form-errors";
import { networkErrorMessage } from "@/lib/auth/messages";
import { checkDisplayName, checkPassword, DISPLAY_NAME_MAX, PASSWORD_MIN } from "@/lib/auth/rules";
import { useResetOnHide } from "@/lib/auth/use-reset-on-hide";

import { FormAlert } from "./form-alert";

const UpdateDisplayNameMutation = graphql(`
  mutation UpdateDisplayName($input: UpdateDisplayNameInput!) {
    updateDisplayName(input: $input) {
      viewer {
        id
        displayName
      }
      errors {
        field
        code
      }
    }
  }
`);

const ChangePasswordMutation = graphql(`
  mutation ChangePassword($input: ChangePasswordInput!) {
    changePassword(input: $input) {
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

const loginAgain = "/login?next=%2Faccount";

type AccountSettingsProps = {
  loginId: string;
  displayName: string;
};

export function AccountSettings({ loginId, displayName }: AccountSettingsProps) {
  return (
    <div className="flex flex-col gap-10">
      <section aria-labelledby="login-id-heading" className="flex flex-col gap-2">
        <h2 id="login-id-heading" className="text-lg font-bold">
          ログイン ID
        </h2>
        <p className="font-mono">{loginId}</p>
        <p className="text-sm text-muted-foreground">他の人には表示されません。変えることはできません。</p>
      </section>
      <DisplayNameForm initial={displayName} />
      <PasswordForm />
    </div>
  );
}

function DisplayNameForm({ initial }: { initial: string }) {
  const router = useRouter();
  const [displayName, setDisplayName] = useState(initial);
  const [errors, setErrors] = useState<FormErrors>(noFormErrors);
  const [saved, setSaved] = useState(false);
  const [{ fetching }, update] = useMutation(UpdateDisplayNameMutation);
  useResetOnHide(() => {
    setErrors(noFormErrors);
    setSaved(false);
  });

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSaved(false);
    const invalid = ruleErrors([checkDisplayName(displayName)]);
    if (invalid) {
      setErrors(invalid);
      return;
    }
    const result = await update({ input: { displayName } });
    if (isUnauthenticated(result.error)) {
      router.replace(loginAgain);
      return;
    }
    const payload = result.data?.updateDisplayName;
    if (!payload) {
      setErrors({ fields: {}, form: networkErrorMessage });
      return;
    }
    if (payload.errors.length > 0 || !payload.viewer) {
      setErrors(toFormErrors(payload.errors));
      return;
    }
    setErrors(noFormErrors);
    setDisplayName(payload.viewer.displayName);
    setSaved(true);
    // ヘッダーの表示名を描き直す
    router.refresh();
  }

  return (
    <section aria-labelledby="display-name-heading" className="flex flex-col gap-4">
      <h2 id="display-name-heading" className="text-lg font-bold">
        表示名
      </h2>
      <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-4">
        {errors.form && <FormAlert tone="error">{errors.form}</FormAlert>}
        {saved && <FormAlert tone="success">表示名を変えました</FormAlert>}
        <TextField
          label="表示名"
          name="displayName"
          autoComplete="nickname"
          description={`他の人にも表示されます（${DISPLAY_NAME_MAX} 文字まで）`}
          required
          value={displayName}
          onValueChange={setDisplayName}
          error={errors.fields.DISPLAY_NAME}
        />
        <Button type="submit" className="self-start" disabled={fetching}>
          表示名を保存する
        </Button>
      </form>
    </section>
  );
}

function PasswordForm() {
  const router = useRouter();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [errors, setErrors] = useState<FormErrors>(noFormErrors);
  const [saved, setSaved] = useState(false);
  const [{ fetching }, change] = useMutation(ChangePasswordMutation);
  useResetOnHide(() => {
    setCurrentPassword("");
    setNewPassword("");
    setErrors(noFormErrors);
    setSaved(false);
  });

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setSaved(false);
    const invalid = ruleErrors([
      currentPassword === "" ? { field: "CURRENT_PASSWORD", code: "REQUIRED" } : null,
      checkPassword("NEW_PASSWORD", newPassword),
    ]);
    if (invalid) {
      setErrors(invalid);
      return;
    }
    const result = await change({ input: { currentPassword, newPassword } });
    if (isUnauthenticated(result.error)) {
      router.replace(loginAgain);
      return;
    }
    const payload = result.data?.changePassword;
    if (!payload) {
      setErrors({ fields: {}, form: networkErrorMessage });
      return;
    }
    if (payload.errors.length > 0) {
      setErrors(toFormErrors(payload.errors));
      return;
    }
    setErrors(noFormErrors);
    setCurrentPassword("");
    setNewPassword("");
    setSaved(true);
  }

  return (
    <section aria-labelledby="password-heading" className="flex flex-col gap-4">
      <h2 id="password-heading" className="text-lg font-bold">
        パスワード
      </h2>
      <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-4">
        {errors.form && <FormAlert tone="error">{errors.form}</FormAlert>}
        {saved && (
          <FormAlert tone="success">
            パスワードを変えました。ほかの端末では、ログインし直す必要があります
          </FormAlert>
        )}
        <TextField
          label="現在のパスワード"
          name="currentPassword"
          type="password"
          autoComplete="current-password"
          required
          value={currentPassword}
          onValueChange={setCurrentPassword}
          error={errors.fields.CURRENT_PASSWORD}
        />
        <TextField
          label="新しいパスワード"
          name="newPassword"
          type="password"
          autoComplete="new-password"
          description={`${PASSWORD_MIN} 文字以上。変えると、ほかの端末のログイン状態はすべて切れます`}
          required
          value={newPassword}
          onValueChange={setNewPassword}
          error={errors.fields.NEW_PASSWORD}
        />
        <Button type="submit" className="self-start" disabled={fetching}>
          パスワードを変える
        </Button>
      </form>
    </section>
  );
}
