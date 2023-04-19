#!/bin/bash

# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "external" data source
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
