#!/bin/bash

# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "external" data source
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
