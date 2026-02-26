# Object storage

Neptune stores stack lock files in object storage. Supported backends and their required environment variables are below.

## Google Cloud Storage (GCS)

- **URL format**: `gs://bucket` or `gs://bucket/prefix`
- **Credentials**: Set `GOOGLE_APPLICATION_CREDENTIALS` to the path of a service account key file, or use [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) (e.g. `gcloud auth application-default login`).

## AWS S3

- **URL format**: `s3://bucket` or `s3://bucket/prefix`
- **Credentials**: The AWS SDK reads credentials from the environment or default chain. Set:
  - `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for access keys, and
  - `AWS_REGION` (e.g. `us-east-1`).
  Alternatively use IAM roles (e.g. on EC2 or in GitHub Actions with OIDC) with no env vars.

## MinIO (S3-compatible)

- **URL format**: `s3://bucket` or `s3://bucket/prefix` (same as S3)
- **Credentials**: Use the same variables as AWS S3 for access and secret keys. Point the client to your MinIO endpoint:
  - `AWS_ENDPOINT_URL_S3` (or `AWS_ENDPOINT_URL`) to the MinIO URL, e.g. `http://minio.example.com:9000`
- **Region**: Set `AWS_REGION` if required by your MinIO setup (e.g. `us-east-1`).
