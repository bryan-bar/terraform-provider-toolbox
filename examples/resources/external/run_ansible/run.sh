# Exit if any of the intermediate steps fail
set -e

# Handle stdin from Terraform "external" data source
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
