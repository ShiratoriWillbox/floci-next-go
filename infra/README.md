# インフラ補助ファイル

## `s3-cors.json`

ブラウザからプリサインド URL へ直接 `PUT` するために、オブジェクトストレージ（S3 または互換 API）のバケットに CORS を設定します。

- **ローカル**: `docker-compose.yml` の `floci-createbuckets` がバケット作成後に `put-bucket-cors` を実行します。
- **本番 AWS**: バケットに同内容を適用してください。

```bash
aws s3api put-bucket-cors \
  --bucket "$S3_BUCKET" \
  --cors-configuration file://infra/s3-cors.json
```

ローカル開発では `AllowedOrigins: ["*"]` で問題になりにくいです。本番ではフロントのオリジン（例: `https://app.example.com`）に絞ることを推奨します。
