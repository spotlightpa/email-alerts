#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

GOPKG="github.com/spotlightpa/email-alerts/pkg/emailalerts"
URL=${DEPLOY_PRIME_URL:-http://local.dev}
LDFLAGS="-X '$GOPKG.DeployURL=$URL'"
CGO_ENABLED=0 GOBIN=$THIS_DIR/functions go install -ldflags "$LDFLAGS" ./cmd/...
