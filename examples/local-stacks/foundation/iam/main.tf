terraform {
  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.0"
    }
  }
}

resource "null_resource" "foundation_iam" {
  triggers = {
    stack = "foundation/iam"
  }
}

resource "local_file" "out" {
  content  = "foundation/iam output"
  filename = "${path.module}/out.txt"
}
