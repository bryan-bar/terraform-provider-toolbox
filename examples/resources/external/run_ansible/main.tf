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

resource "toolbox_external" "ansible_script" {
  program = [
    "bash",
    "${path.module}/run.sh"
  ]
  query = {
    mount_points = base64encode(jsonencode(local.mount_points))
    ssh_user = base64encode(jsonencode(local.ssh_user))
    ip_address = base64encode(jsonencode(local.ip_address))
  }
}

output "ansible_output" {
  value = toolbox_external.ansible_script.result
}

resource "toolbox_external" "ansible_script_nomounts" {
  program = [
    "bash",
    "${path.module}/run.sh"
  ]
  query = {
    ssh_user = base64encode(jsonencode(local.ssh_user))
    ip_address = base64encode(jsonencode(local.ip_address))
  }
}

output "ansible_nomounts" {
  value = toolbox_external.ansible_script_nomounts.result
}

resource "toolbox_external" "ansible_base64" {
  program = [
    "bash",
    "${path.module}/run_base64.sh"
  ]
  query = {
    mount_points = base64encode(jsonencode(local.mount_points))
    ssh_user = base64encode(jsonencode(local.ssh_user))
    ip_address = base64encode(jsonencode(local.ip_address))
  }
}

output "ansible_base64" {
  value = {for k,v in toolbox_external.ansible_base64.result: k=>jsondecode(base64decode(v))}
}

output "ansible_base64_raw" {
  value = toolbox_external.ansible_base64.result
}

