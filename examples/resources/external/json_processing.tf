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

resource "toolbox_external" "json_processing" {
  program = [
    "bash",
    "${path.module}/json-processing.sh"]

  query = {
    # arbitrary map from strings to strings, passed
    # to the external program as the data query.
    mount_points = local.mount_points
    ssh_user = local.ssh_user
    ip_address = local.ip_address
  }
}

output "results" {
  value = [
    toolbox_external.json_processing.results
  ]
}
