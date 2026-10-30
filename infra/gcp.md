# Google Cloud

API（Go）を Cloud Run で動かすための手順。コマンドは `gcloud` CLI の例。
`<...>` は実際の値に置き換える。

```sh
PROJECT_ID=<プロジェクト ID>
REGION=asia-southeast1
GITHUB_REPO=Qurioth/chara-q
```

## 1. プロジェクトと API の有効化

```sh
gcloud projects create "$PROJECT_ID"
gcloud billing projects link "$PROJECT_ID" --billing-account=<請求先アカウント ID>
gcloud config set project "$PROJECT_ID"

gcloud services enable \
  run.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  iamcredentials.googleapis.com \
  sts.googleapis.com \
  billingbudgets.googleapis.com
```

## 2. 予算アラート（必須）

課金を有効にしたら、他のリソースを作る前に設定する。

```sh
gcloud billing budgets create \
  --billing-account=<請求先アカウント ID> \
  --display-name="chara-q monthly budget" \
  --budget-amount=<月の上限。例: 1000JPY> \
  --filter-projects="projects/$PROJECT_ID" \
  --threshold-rule=percent=0.5 \
  --threshold-rule=percent=0.9 \
  --threshold-rule=percent=1.0 \
  --threshold-rule=percent=1.0,basis=forecasted-spend
```

- 通知は請求先アカウントの管理者にメールで届く。届くことを確認しておく。
- 予算アラートは通知だけで、課金は止まらない。超えた場合は手でサービスを止める。

## 3. Artifact Registry

```sh
gcloud artifacts repositories create chara-q \
  --repository-format=docker \
  --location="$REGION"
```

古いイメージが溜まらないように、クリーンアップポリシー（例: 最新 10 件を残す）を設定する。

## 4. サービスアカウント

| 名前 | 用途 | ロール |
|---|---|---|
| `chara-q-api` | Cloud Run の実行時 | 使うシークレットへの `roles/secretmanager.secretAccessor` |
| `github-deployer` | GitHub Actions からのデプロイ | `roles/run.developer`、`roles/artifactregistry.writer`、`chara-q-api` への `roles/iam.serviceAccountUser` |

```sh
gcloud iam service-accounts create chara-q-api
gcloud iam service-accounts create github-deployer

DEPLOYER="github-deployer@$PROJECT_ID.iam.gserviceaccount.com"
RUNTIME="chara-q-api@$PROJECT_ID.iam.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" --member="serviceAccount:$DEPLOYER" --role=roles/run.developer
gcloud artifacts repositories add-iam-policy-binding chara-q --location="$REGION" \
  --member="serviceAccount:$DEPLOYER" --role=roles/artifactregistry.writer
gcloud iam service-accounts add-iam-policy-binding "$RUNTIME" \
  --member="serviceAccount:$DEPLOYER" --role=roles/iam.serviceAccountUser
```

サービスアカウントの **鍵は作らない**。GitHub Actions からは次の Workload Identity 連携で認証する。

## 5. Workload Identity 連携（GitHub Actions 用）

```sh
gcloud iam workload-identity-pools create github --location=global

gcloud iam workload-identity-pools providers create-oidc github \
  --location=global \
  --workload-identity-pool=github \
  --issuer-uri="https://token.actions.githubusercontent.com" \
  --attribute-mapping="google.subject=assertion.sub,attribute.repository=assertion.repository,attribute.ref=assertion.ref" \
  --attribute-condition="assertion.repository == '$GITHUB_REPO' && assertion.ref == 'refs/heads/main'"

PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
gcloud iam service-accounts add-iam-policy-binding "$DEPLOYER" \
  --role=roles/iam.workloadIdentityUser \
  --member="principalSet://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/github/attribute.repository/$GITHUB_REPO"
```

- `attribute-condition` で、このリポジトリの main ブランチからの実行だけを許可する。

## 6. Secret Manager

```sh
# 値は標準入力から渡す（シェルの履歴に残さない）
gcloud secrets create database-url --replication-policy=automatic --data-file=-
gcloud secrets add-iam-policy-binding database-url \
  --member="serviceAccount:$RUNTIME" --role=roles/secretmanager.secretAccessor
```

R2 のアクセスキーなど、後の Spec で必要になったシークレットも同じ手順で追加する。

## 7. Cloud Run

初回のデプロイは GitHub Actions の `deploy-api.yml`（`workflow_dispatch`）で行う。
サービスの主な設定は次のとおり（ワークフローで指定する）。

| 設定 | 値 |
|---|---|
| リージョン | `asia-southeast1` |
| サービスアカウント | `chara-q-api` |
| 最小インスタンス数 | **0**（常時起動しない） |
| 最大インスタンス数 | 小さく始める（例: 2） |
| 認証 | 未認証の呼び出しを許可（アプリ側で JWT を検証する） |
| 環境変数 | `APP_ENV=production`、`CORS_ALLOWED_ORIGINS=<Vercel のドメイン>` |
| シークレット | `DATABASE_URL=database-url:latest` |

- `PORT` は Cloud Run が渡すので設定しない。
- 本番のマイグレーションをいつ・どう実行するか（Cloud Run ジョブ、デプロイ前のステップなど）は、
  本番環境構築の Spec で決める。

## 8. GitHub Actions の変数

リポジトリの Settings → Secrets and variables → Actions → **Variables** に登録する
（どれも秘密情報ではない）。

| 名前 | 値の例 |
|---|---|
| `GCP_PROJECT_ID` | `<プロジェクト ID>` |
| `GCP_REGION` | `asia-southeast1` |
| `GCP_WORKLOAD_IDENTITY_PROVIDER` | `projects/<番号>/locations/global/workloadIdentityPools/github/providers/github` |
| `GCP_DEPLOY_SERVICE_ACCOUNT` | `github-deployer@<プロジェクト ID>.iam.gserviceaccount.com` |
| `GCP_RUNTIME_SERVICE_ACCOUNT` | `chara-q-api@<プロジェクト ID>.iam.gserviceaccount.com` |
| `CLOUD_RUN_CORS_ALLOWED_ORIGINS` | `https://<Vercel のドメイン>` |
