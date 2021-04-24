# manual test:
# run
# ../../bin/timeit terraform apply -auto-approve
# and hit CTRL-C

terraform {
  required_version = ">= 0.14.9"
}

resource "null_resource" "sleep" {
  provisioner "local-exec" {
    command = "sleep 10"
  }
}
