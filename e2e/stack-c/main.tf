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

resource "null_resource" "stack_c" {
  triggers = {
    stack = "stack-c"
  }
}

resource "local_file" "out" {
  content  = "stack-c output"
  filename = "${path.module}/out.txt"
}
