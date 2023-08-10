terraform {
  required_providers {
    toolbox = {
      source = "bryan-bar/toolbox"
    }
  }
}

resource "toolbox_external" "ansible" {

  program = [
    "bash",
    "-c",
    <<EOF
      output=$(ANSIBLE_STDOUT_CALLBACK=json ansible-playbook playbook.yml | jq -r .plays[0].tasks[0].hosts.localhost.stdout)
      jq -n --arg output "$output" '{"stdout":$output}'
    EOF
  ]
}

output "test" {
  value = toolbox_external.ansible.result
}
