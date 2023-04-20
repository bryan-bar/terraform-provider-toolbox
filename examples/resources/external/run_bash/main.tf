terraform {
  required_providers {
    toolbox = {
      source = "bryan-bar/toolbox"
    }
  }
}

locals {
  mount_points = ["/test0", "/test2", "/test3/dir"]
  ssh_user = "rocky"
  ip_address = "1.2.3.4"
}

resource "toolbox_external" "bash_script" {
  program = [
    "bash",
    "${path.module}/run.sh"
  ]

  query = {
    # arbitrary map from strings to strings, passed
    # to the external program as the data query.
    mount_points = base64encode(jsonencode(local.mount_points))
    ssh_user = base64encode(local.ssh_user)
    ip_address = base64encode(local.ip_address)
  }
}

output "results" {
  value = [
    toolbox_external.bash_script.result
  ]
}
