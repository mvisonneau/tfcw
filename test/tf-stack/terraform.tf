provider "local" {
   version = "~> 1.4.0"
}

variable "credentials" {}

resource "local_file" "credentials_file" {
  filename        = "./credentials"
  file_permission = "0600"
  content         = var.credentials
}
