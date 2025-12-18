#!/usr/bin/env bash

set -euo pipefail

# Cleans up custom roles and role assignments created by MPF (names prefixed tmp-rol-).
# Usage: SUBSCRIPTION_ID=<id> ./cleanup-tmp-roles.sh
# Optional: TMP_ROLE_PREFIX (default: tmp-rol-)

if ! command -v az >/dev/null 2>&1; then
  echo "Azure CLI (az) is required" >&2
  exit 1
fi
if ! command -v jq >/dev/null 2>&1; then
  echo "jq is required" >&2
  exit 1
fi

SUBSCRIPTION_ID="${SUBSCRIPTION_ID:-${1:-}}"
ROLE_PREFIX="${TMP_ROLE_PREFIX:-tmp-rol-}"

if [[ -z "${SUBSCRIPTION_ID}" ]]; then
  SUBSCRIPTION_ID=$(az account show --query id -o tsv 2>/dev/null || true)
fi

if [[ -z "${SUBSCRIPTION_ID}" ]]; then
  echo "Subscription not provided and no default found. Set SUBSCRIPTION_ID or login." >&2
  exit 1
fi

az account set --subscription "${SUBSCRIPTION_ID}"

query="[?starts_with(roleName, '$ROLE_PREFIX')].[id,name,roleName]"
defs_json=$(az role definition list --custom-role-only true --subscription "${SUBSCRIPTION_ID}" \
  --query "${query}" -o json)

if [[ "$(echo "${defs_json}" | jq 'length')" -eq 0 ]]; then
  echo "No tmp-rol* custom roles found"
  exit 0
fi

echo "${defs_json}" | jq -c '.[]' | while read -r def; do
  role_id=$(echo "${def}" | jq -r '.[0]')
  role_name=$(echo "${def}" | jq -r '.[1]')
  role_display=$(echo "${def}" | jq -r '.[2]')

  echo "Processing role ${role_display} (${role_name})"

  assignments=$(az role assignment list --all --subscription "${SUBSCRIPTION_ID}" --role "${role_id}" --query '[].id' -o tsv)
  if [[ -n "${assignments}" ]]; then
    echo "${assignments}" | while read -r assign_id; do
      [[ -z "${assign_id}" ]] && continue
      echo "Deleting assignment ${assign_id}"
      az role assignment delete --ids "${assign_id}"
    done
  else
    echo "No assignments for ${role_display}"
  fi

  echo "Deleting role definition ${role_name}"
  az role definition delete --name "${role_name}" --subscription "${SUBSCRIPTION_ID}"
  echo "Done ${role_display}"
  echo
done

