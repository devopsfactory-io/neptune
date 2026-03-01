# S3 backend example

Neptune stores stack lock files in **AWS S3** (or S3-compatible storage such as MinIO). This example uses `object_storage: s3://your-bucket` in `.neptune.yaml`.

## Setup

1. Replace `s3://your-bucket` in `.neptune.yaml` with your bucket (and optional prefix), e.g. `s3://my-neptune-locks` or `s3://my-bucket/neptune/prefix`.
2. Set environment variables for the AWS SDK (see [Neptune object storage docs](https://github.com/devopsfactory-io/neptune/blob/main/docs/object-storage.md)):
   - **AWS S3**: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` (e.g. `us-east-1`). Alternatively use IAM roles (e.g. GitHub Actions OIDC) with no env vars.
   - **MinIO (S3-compatible)**: Same as above, plus `AWS_ENDPOINT_URL_S3` (or `AWS_ENDPOINT_URL`) pointing to your MinIO URL, e.g. `http://minio.example.com:9000`. Set `AWS_REGION` if your MinIO setup requires it.

No AWS resources are created by the Terraform in this example (only `null_resource` and `local_file`); Neptune uses S3 only for lock files.
