#!/bin/bash

set -eu
set -o pipefail

THIS_FILE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CI="${THIS_FILE_DIR}/../../wg-app-platform-runtime-ci"
. "$CI/shared/helpers/release-note-helpers.bash"
. "$CI/shared/helpers/git-helpers.bash"
unset THIS_FILE_DIR

# ex. version_range="v0.343.0...v0.344.0"
version_range="${1:?Please provide the start and end versions you want to generate release notes for './generate-release-notes.bash START_REF...END_REF' }"
# ex. local_start_ref="v0.343.0"
local_start_ref=$(get_start_ref_from_range "${version_range}")
# ex. local_end_ref="v0.344.0"
local_end_ref=$(get_end_ref_from_range "${version_range}")

get_non_bot_commits "${local_start_ref}" "${local_end_ref}"
echo ""

START_REF_DOCKER_DRIVER=$(git rev-parse "${local_start_ref}:src/code.cloudfoundry.org/dockerdriver")
END_REF_DOCKER_DRIVER=$(git rev-parse "${local_end_ref}:src/code.cloudfoundry.org/dockerdriver")
pushd src/code.cloudfoundry.org/dockerdriver > /dev/null
  display_go_mod_diff "${START_REF_DOCKER_DRIVER}" "${END_REF_DOCKER_DRIVER}" go.mod "dockerdriver"
  echo ""
popd > /dev/null

display_go_mod_diff "${local_start_ref}" "${local_end_ref}" src/code.cloudfoundry.org/mapfs/go.mod "mapfs"
echo ""

display_go_mod_diff "${local_start_ref}" "${local_end_ref}" src/code.cloudfoundry.org/mapfs-performance-acceptance-tests/go.mod "mapfs-performance-acceptance-tests"
echo ""

display_go_mod_diff "${local_start_ref}" "${local_end_ref}" src/code.cloudfoundry.org/nfsbroker/go.mod "nfsbroker"
echo ""

display_go_mod_diff "${local_start_ref}" "${local_end_ref}" src/code.cloudfoundry.org/nfsbroker/go.mod "nfsv3driver"
echo ""

display_blob_change_info "${local_start_ref}" "${local_end_ref}" config/blobs.yml
