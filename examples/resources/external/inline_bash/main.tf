terraform {
  required_providers {
    toolbox = {
      source = "bryan-bar/toolbox"
    }
  }
}

locals {
  mapped = {"one": "item1", "two": "item2"}
  listed = ["one", "item1", "two", "item2"]
  mixed = {"mapped": local.mapped, "listed": local.listed}
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
