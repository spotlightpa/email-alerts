#!/bin/bash

set -eu -o pipefail

# Get the directory that this script file is in
THIS_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

cd "$THIS_DIR"

GOPKG="github.com/spotlightpa/email-alerts/pkg/emailalerts"
BUILD_VERSION="$(git rev-parse --short HEAD)"
URL=${DEPLOY_PRIME_URL:-http://local.dev}
LDFLAGS="-X '$GOPKG.BuildVersion=$BUILD_VERSION' -X '$GOPKG.DeployURL=$URL'"
GOBIN=$THIS_DIR/functions go install -ldflags "$LDFLAGS" ./cmd/...
