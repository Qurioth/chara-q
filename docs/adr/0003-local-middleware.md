# ADR 0003: ローカル開発用のミドルウェア（Postgres 18 と RustFS）

- ステータス: 採用
- 日付: 2026-10-30

## 背景

ローカルでは Docker Compose で Neon と Cloudflare R2 の代わりを動かす。
当初は「Neon と同じメジャーバージョンの Postgres」と「MinIO など S3 互換ストレージ」を想定していた。

## 決定

- **Postgres は 18 を使う**（`postgres:18.x-alpine`）。Neon は 2026-06 以降、新規プロジェクトの
  既定を Postgres 18 にしているため、本番もこれに合わせる。
  - Postgres 18 以降の公式イメージはデータを `/var/lib/postgresql/<メジャー>/docker` に置くため、
    ボリュームは `/var/lib/postgresql` にマウントする。
- **S3 互換ストレージは RustFS を使う**（`rustfs/rustfs`）。
  - MinIO は 2025-10 からコミュニティ版の Docker イメージとバイナリの配布をやめており、
    Docker Hub の `minio/minio` も取得できなくなっている。
  - RustFS は Apache 2.0 ライセンスで、MinIO と同じポート（API 9000 / コンソール 9001）で動き、
    1 コンテナで済む。
  - 開発用バケット `chara-q-local` は、Compose の `storage-init`（aws-cli）が作る。
- アプリは S3 互換 API だけを使い、RustFS 固有の機能には依存しない。本番の R2 とは
  エンドポイントと認証情報だけを切り替える。

## 結果

- RustFS は 1.0 になって間もないため、問題が出た場合は SeaweedFS などへ置き換える。
  置き換えはアプリに影響しない（S3 互換 API だけを使うため）。
