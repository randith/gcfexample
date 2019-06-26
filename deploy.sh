#!/usr/bin/env bash
set -ex

# Create or update

pushd "$(dirname "$0")"

# [[ -z "${1}" ]] && echo "missing first argument REGION (i.e. us)" && exit 1
REGION=us

# [[ -z "${2}" ]] && echo "missing second argument ENVIRONMENT (i.e. staging)" && exit 1
ENVIRONMENT=sandbox

# [[ -z "${3}" ]] && echo "missing third argument GOOGLE_COMPUTE_ZONE (i.e. us-central1)" && exit 1
GOOGLE_COMPUTE_ZONE=us-central1

# [[ -z "${4}" ]] && echo "missing fourth argument GCP_PROJECT (i.e. ce-staging-216319)" && exit 1
GCP_PROJECT=ce-admin-218121


gcloud functions deploy gcfexample --trigger-http --region ${GOOGLE_COMPUTE_ZONE} --runtime go111 --timeout 540 --memory 2gb --project ${GCP_PROJECT}

popd
