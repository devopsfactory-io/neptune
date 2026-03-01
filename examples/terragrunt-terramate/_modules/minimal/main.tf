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
  backend "local" {
    path = "terraform.tfstate"
  }
}

variable "stack_name" {
  type = string
}

resource "null_resource" "stack" {
  triggers = {
    stack = var.stack_name
  }
}

resource "local_file" "out" {
  content  = "${var.stack_name} output"
  filename = "${path.module}/out.txt"
}
