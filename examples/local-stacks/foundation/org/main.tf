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

resource "null_resource" "foundation_org" {
  triggers = {
    stack = "foundation/org"
  }
}

resource "local_file" "out" {
  content  = "foundation/org output"
  filename = "${path.module}/out.txt"
}
