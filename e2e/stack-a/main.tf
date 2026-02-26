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

resource "null_resource" "stack_a" {
  triggers = {
    stack = "stack-a"
  }
}

resource "local_file" "out" {
  content  = "stack-a output"
  filename = "${path.module}/out.txt"
}
