"use client";

import { CopyIcon } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useMutation } from "urql";

import { Button } from "@/components/ui/button";
import { graphql } from "@/gql";
import { isForbidden, isUnauthenticated } from "@/lib/auth/form-errors";
import { forbiddenMessage, formatDateTime, networkErrorMessage } from "@/lib/auth/messages";
import { useResetOnHide } from "@/lib/auth/use-reset-on-hide";

import { FormAlert } from "./form-alert";

const IssueInvitationMutation = graphql(`
  mutation IssueInvitation {
    issueInvitation {
      code
      expiresAt
    }
  }
`);

/** 招待コードを発行する（管理者だけ）。コードは発行したときにしか表示できない */
export function InvitationIssuer() {
  const router = useRouter();
  const [issued, setIssued] = useState<{ code: string; expiresAt: string } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [{ fetching }, issue] = useMutation(IssueInvitationMutation);
  // 画面を離れたら、発行したコードを DOM に残さない（コードは発行したときにしか表示しない）
  useResetOnHide(() => {
    setIssued(null);
    setError(null);
    setCopied(false);
  });

  async function handleIssue() {
    setError(null);
    setCopied(false);
    const result = await issue({});
    if (isUnauthenticated(result.error)) {
      router.replace("/login?next=%2Fadmin%2Finvitations");
      return;
    }
    if (isForbidden(result.error)) {
      setError(forbiddenMessage);
      return;
    }
    const payload = result.data?.issueInvitation;
    if (!payload) {
      setError(networkErrorMessage);
      return;
    }
    setIssued(payload);
  }

  async function handleCopy() {
    if (!issued) return;
    await navigator.clipboard.writeText(issued.code);
    setCopied(true);
  }

  return (
    <div className="flex flex-col gap-6">
      <p className="text-sm leading-relaxed text-muted-foreground">
        招待コードは 1 回だけ使え、発行から 7 日で使えなくなります。招待する人に、チャットなどで渡してください。
      </p>
      <Button type="button" size="lg" className="self-start" disabled={fetching} onClick={handleIssue}>
        招待コードを発行する
      </Button>
      {error && <FormAlert tone="error">{error}</FormAlert>}
      {issued && (
        <section aria-labelledby="issued-heading" className="flex flex-col gap-3 rounded-lg border border-border p-4">
          <h2 id="issued-heading" className="font-bold">
            発行した招待コード
          </h2>
          <div className="flex flex-wrap items-center gap-2">
            <code className="rounded-md bg-muted px-2 py-1 font-mono text-lg break-all">{issued.code}</code>
            <Button type="button" variant="outline" onClick={handleCopy}>
              <CopyIcon aria-hidden />
              コピー
            </Button>
          </div>
          <p className="text-sm">有効期限: {formatDateTime(issued.expiresAt)}</p>
          {copied && <p role="status" className="text-sm">コピーしました</p>}
          <p className="text-sm text-muted-foreground">
            この画面を離れると、このコードは二度と表示できません。
          </p>
        </section>
      )}
    </div>
  );
}
