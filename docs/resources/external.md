---
page_title: "toolbox_external Resource - terraform-provider-toolbox"
description: |-
  The external resource allows an external program implementing a specific protocol (defined below) to act as a resource, exposing arbitrary data for use elsewhere in the Terraform configuration.
  As of now, this resource will be re-created if any value is changed.Similar to null_resource combined with trigger but the output can be saved into our state.
  Warning This mechanism is provided as an "escape hatch" for exceptional situations where a first-class Terraform provider is not more appropriate. Its capabilities are limited in comparison to a true resource, and implementing a resource via an external program is likely to hurt the portability of your Terraform configuration by creating dependencies on external programs and libraries that may not be available (or may need to be used differently) on different operating systems.
  Warning Terraform Enterprise does not guarantee availability of any particular language runtimes or external programs beyond standard shell utilities, so it is not recommended to use this resource within configurations that are applied within Terraform Enterprise.
---

# toolbox_external

The `external` resource allows an external program implementing a specific protocol (defined below) to act as a resource, exposing arbitrary data for use elsewhere in the Terraform configuration.
As of now, this resource will be re-created if any value is changed.Similar to null_resource combined with trigger but the output can be saved into our state.

**Warning** This mechanism is provided as an "escape hatch" for exceptional situations where a first-class Terraform provider is not more appropriate. Its capabilities are limited in comparison to a true resource, and implementing a resource via an external program is likely to hurt the portability of your Terraform configuration by creating dependencies on external programs and libraries that may not be available (or may need to be used differently) on different operating systems.

**Warning** Terraform Enterprise does not guarantee availability of any particular language runtimes or external programs beyond standard shell utilities, so it is not recommended to use this resource within configurations that are applied within Terraform Enterprise.

## Example Usage

#### Inline Bash
```terraform
terraform {
  required_providers {
    toolbox = {
      source = "EnterpriseDB/toolbox"
    }
  }
}

locals {
  mapped = {"one": "item1", "two": "item2"}
  listed = ["one", "item1", "two", "item2"]
  mixed = {"mapped": local.mapped, "listed": local.listed}
}

resource "toolbox_external" "basic" {
  create = true
  read = false
  update = false
  delete = false
  // Recreate is used to force a destroy - create cycle
  recreate = {
    one = "two"
  }
  query = {
    one = "two"
  }
  program = [
    "bash",
    "-c",
    <<EOF
      # query is passed to stdin as a JSON object and
      # will contain the reserved keys "stage" and "old_result"
      read input

      # Return a json object to be stored in the result attribute
      # "old_result" is not allowed to prevent duplicate old_result when passed back into query
      printf '{"one":"two"}'
    EOF
  ]
}

resource "toolbox_external" "mapped" {
  program = [
    "bash",
    "-c",
    <<EOF
      somevar=${jsonencode(local.mapped)}
      jq -n --arg somevar "$somevar" \
         '{"stdout":$somevar}'  
    EOF
  ]
}

resource "toolbox_external" "mapped_encoded" {
  program = [
    "bash",
    "-c",
    <<EOF
      somevar=${base64encode(jsonencode(local.mapped))}
      jq -n --arg somevar "$somevar" \
         '{"stdout":$somevar}'
    EOF
  ]
}

output "mapped_results" {
  value = [
    toolbox_external.mapped.result,
    toolbox_external.mapped_encoded.result,
    {for k,v in toolbox_external.mapped_encoded.result: k=>jsondecode(base64decode(v))},
  ]
}

resource "toolbox_external" "listed" {
  program = [
    "bash",
    "-c",
    <<EOF
      somevar=${jsonencode(local.listed)}
      jq -n --arg somevar "$somevar" \
         '{"stdout":$somevar}'
    EOF
  ]
}

resource "toolbox_external" "listed_encoded" {
  program = [
    "bash",
    "-c",
    <<EOF
      somevar=${base64encode(jsonencode(local.listed))}
      jq -n --arg somevar "$somevar" \
         '{"stdout":$somevar}'
    EOF
  ]
}

output "listed_results" {
  value = [
    toolbox_external.listed.result,
    toolbox_external.listed_encoded.result,
    {for k,v in toolbox_external.listed_encoded.result: k=>jsondecode(base64decode(v))},
  ]
}

resource "toolbox_external" "mixed" {
  program = ["bash", "-c",<<EOF
    somevar=${jsonencode(local.mixed)}
    jq -n --arg somevar "$somevar" \
       '{"stdout":$somevar}'
  EOF
  ]
}

resource "toolbox_external" "mixed_encoded" {
  program = ["bash", "-c",<<EOF
    somevar=${base64encode(jsonencode(local.mixed))}
    jq -n --arg somevar "$somevar" \
       '{"stdout":$somevar}'
  EOF
  ]
}

output "mixed_results" {
  value = [
    toolbox_external.mixed.result,
    toolbox_external.mixed_encoded.result,
    {for k,v in toolbox_external.mixed_encoded.result: k=>jsondecode(base64decode(v))},
  ]
}
```

#### Inline Ansible
```terraform
terraform {
  required_providers {
    toolbox = {
      source = "EnterpriseDB/toolbox"
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
```
```yaml
---
- hosts: localhost
  gather_facts: false
  tasks:
    - name: test
      ansible.builtin.shell:
        cmd: >
          echo "Echo from playbook"
```

## External Program Protocol

The external program described by the `program` attribute must implement a
specific protocol for interacting with Terraform, as follows.

The program must read all of the data passed to it on `stdin`, and parse
it as a JSON object. The JSON object contains the contents of the `query`
argument and its values will always be strings.

The program must then produce a valid JSON object on `stdout`, which will
be used to populate the `result` attribute exported to the rest of the
Terraform configuration. This JSON object must again have all of its
values as strings. On successful completion it must exit with status zero.

If the program encounters an error and is unable to produce a result, it
must print a human-readable error message (ideally a single line) to `stderr`
and exit with a non-zero status. Any data on `stdout` is ignored if the
program returns a non-zero status.

All environment variables visible to the Terraform process are passed through
to the child program.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `program` (List of String) A list of strings, whose first element is the program to run and whose subsequent elements are optional command line arguments to the program. Terraform does not execute the program through a shell, so it is not necessary to escape shell metacharacters nor add quotes around arguments containing spaces.

### Optional

- `create` (Boolean) Run on create: enabled by default
- `delete` (Boolean) Run on delete: disabled by default
- `query` (Map of String) A map of string values to pass to the external program as the query arguments. If not supplied, the program will receive an empty object as its input.
- `read` (Boolean) Run on read: disabled by default
- `recreate` (Map of String) A map of string values to force a replace on the resource. If not supplied, the resource will not be replaced.
- `update` (Boolean) Run on update: disabled by default
- `working_dir` (String) Working directory of the program. If not supplied, the program will run in the current directory.

### Read-Only

- `id` (String) The id of the resource. This will always be set to `-`
- `result` (Map of String) A map of string values returned from the external program.
- `stage` (String) The stage of the resource.

## Processing JSON in shell scripts

Since the external resource protocol uses JSON, it is recommended to use
the utility [`jq`](https://stedolan.github.io/jq/) to translate to and from
JSON in a robust way when implementing a resource in a shell scripting
language.

The following example shows some input/output boilerplate code for a
resource implemented in bash:

```terraform
terraform {
  required_providers {
    toolbox = {
      source = "EnterpriseDB/toolbox"
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
```
```shell
#!/bin/bash

# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "toolbox_external" resource
# A parameter "query" of type map(string) is passed to stdin
# In order to control the expected output, parameters use base64 encoding
# ex: query = {
#       "mount_points" = base64encode(jsonencode(var.machine.spec.additional_volumes[*].mount_point))
#       "ssh_user"     = base64encode(var.operating_system.ssh_user)
#       "ip_address"   = base64encode(aws_instance.machine.public_ip)
#     }
# stdin: {
#  "ip_address": "NTIuOTEuMjMwLjEzNQ==",
#  "mount_points": "WyIvb3B0L3BnX2RhdGEiLCIvb3B0L3BnX3dhbCJd",
#  "ssh_user": "cm9ja3k="
# }
#
# Grab stdin with 'jq' and
# insert decoded values into an associative array
TERRAFORM_INPUT=$(jq '.')
declare -A INPUT_MAPPING
for key in $(echo "${TERRAFORM_INPUT}" | jq -r 'keys_unsorted|.[]'); do
    INPUT_MAPPING["$key"]=$(echo "$TERRAFORM_INPUT" | jq -r .[\"$key\"] | base64 -d)
done

json='{}'
for key in "${!INPUT_MAPPING[@]}"; do
    json=$( jq -n --arg json "$json" \
                  --arg key "$key" \
                  --arg value "${INPUT_MAPPING["$key"]}" \
                  '$json | fromjson + { ($key): ($value) }' 
    )
done
echo "$json"
```

## Processing JSON with ansible playbooks
Since the external resource protocol uses JSON, it is recommended to use
the [`posix collection's json_callback`](https://docs.ansible.com/ansible/latest/collections/ansible/posix/json_callback.html).

```terraform
terraform {
  required_providers {
    toolbox = {
      source = "EnterpriseDB/toolbox"
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
```
```yaml
---
- hosts: localhost
  gather_facts: false
  tasks:
    - name: test
      ansible.builtin.shell:
        cmd: >
          echo "Echo from playbook - {{ mount_points|default('mount points not set') }}"
```
```shell
# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "toolbox_external" resource
# A parameter "query" of type map(string) is passed to stdin
# In order to control the expected output, parameters use base64 encoding
# Grab stdin with 'jq' and
# insert decoded values into an associative array
TERRAFORM_INPUT=$(jq '.')
declare -A INPUT_MAPPING
for key in $(echo "${TERRAFORM_INPUT}" | jq -r 'keys_unsorted|.[]'); do
    INPUT_MAPPING["$key"]=$(echo "$TERRAFORM_INPUT" | jq -r .[\"$key\"] | base64 -d)
done

# Re-create json objects from input
json='{}'
for key in "${!INPUT_MAPPING[@]}"; do
    json=$( jq -n --arg json "$json" \
                  --arg key "$key" \
                  --arg value "${INPUT_MAPPING["$key"]}" \
                  '$json | fromjson + { ($key): ($value) }' 
    )
done

ansible_output=$(ANSIBLE_STDOUT_CALLBACK=json ANSIBLE_WHITELIST_CALLBACK=json ansible-playbook playbook.yml --extra-vars "$json")
output=$(echo $ansible_output | jq -r .plays[0].tasks[0].hosts.localhost.stdout)    
jq -n --arg output "$output" '{"stdout":$output}'
```
```shell
# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "toolbox_external" resource
# A parameter "query" of type map(string) is passed to stdin
# In order to control the expected output, parameters use base64 encoding
# Grab stdin with 'jq' and
# insert decoded values into an associative array
TERRAFORM_INPUT=$(jq '.')
declare -A INPUT_MAPPING
for key in $(echo "${TERRAFORM_INPUT}" | jq -r 'keys_unsorted|.[]'); do
    INPUT_MAPPING["$key"]=$(echo "$TERRAFORM_INPUT" | jq -r .[\"$key\"] | base64 -d)
done

# Re-create json objects from input
json='{}'
for key in "${!INPUT_MAPPING[@]}"; do
    json=$( jq -n --arg json "$json" \
                  --arg key "$key" \
                  --arg value "${INPUT_MAPPING["$key"]}" \
                  '$json | fromjson + { ($key): ($value) }' 
    )
done

ansible_output=$(ANSIBLE_STDOUT_CALLBACK=json ANSIBLE_WHITELIST_CALLBACK=json ansible-playbook playbook.yml --extra-vars "$json")
output=$(echo $ansible_output | base64)    
jq -n --arg output "$output" '{"stdout":$output}'
```


## JSON Processing example:
```shell
#!/bin/bash

# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "toolbox_external" resource
# A parameter "query" of type map(string) is passed to stdin
# In order to control the expected output, parameters can use base64 encoding
# ex: query = {
#       "mount_points" = base64encode(jsonencode(var.machine.spec.additional_volumes[*].mount_point))
#       "ssh_user"     = base64encode(var.operating_system.ssh_user)
#       "ip_address"   = base64encode(aws_instance.machine.public_ip)
#     }
# stdin: {
#  "ip_address": "NTIuOTEuMjMwLjEzNQ==",
#  "mount_points": "WyIvb3B0L3BnX2RhdGEiLCIvb3B0L3BnX3dhbCJd",
#  "ssh_user": "cm9ja3k="
# }
#
# Grab stdin with 'jq' and
# insert decoded values into an associative array
TERRAFORM_INPUT=$(jq '.')
declare -A INPUT_MAPPING
for key in $(echo "${TERRAFORM_INPUT}" | jq -r 'keys_unsorted|.[]'); do
    # INPUT_MAPPING["ip_address"]=52.91.164.228
    # INPUT_MAPPING["mount_points"]=["/opt/pg_data","/opt/pg_wal"]
    # $(echo ${INPUT_MAPPING["mount_points"]} | jq -r .[0]) -> /opt/pg_data
    # $(echo ${INPUT_MAPPING["mount_points"]} | jq .[0]) -> "/opt/pg_data"
    # INPUT_MAPPING["ssh_user"]=rocky
    INPUT_MAPPING["$key"]=$(echo "$TERRAFORM_INPUT" | jq -r .[\"$key\"] | base64 -d)
    #echo "DEBUG: key value: $key ${INPUT_MAPPING["$key"]}" >> /tmp/terraform.log
done

# stdout must be returned as a json object
# referenced within terraform from result: 
# toolbox_external.<name>.result
# stderr passed through to terraform as is.
# Safely produce a JSON object containing the result value.
# jq will ensure that the value is properly quoted
# and escaped to produce a valid JSON string.
# jq -n --arg arg0 "$TERRAFORM_INPUT" '{"passed":$arg0}'
json='{}'
for key in "${!INPUT_MAPPING[@]}"; do
    json=$( jq -n --arg json "$json" \
                  --arg key "$key" \
                  --arg value "${INPUT_MAPPING["$key"]}" \
                  '$json | fromjson + { ($key): ($value) }' 
    )
done
echo "$json"
```
