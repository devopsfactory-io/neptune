# GCS backend example

Neptune stores stack lock files in **Google Cloud Storage (GCS)**. This example uses `object_storage: gs://your-bucket` in `.neptune.yaml`.

## Setup

1. Replace `gs://your-bucket` in `.neptune.yaml` with your bucket (and optional prefix), e.g. `gs://my-neptune-locks` or `gs://my-bucket/neptune/prefix`.
2. Set credentials for the Google Cloud client (see [Neptune object storage docs](https://github.com/devopsfactory-io/neptune/blob/main/docs/object-storage.md)):
   - **Service account key**: Set `GOOGLE_APPLICATION_CREDENTIALS` to the path of your service account JSON key file.
   - **Application Default Credentials**: Alternatively use [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) (e.g. `gcloud auth application-default login`).

No GCP resources are created by the Terraform in this example (only `null_resource` and `local_file`); Neptune uses GCS only for lock files.
