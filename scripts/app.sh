#!/usr/bin/env sh

# Fail fast so a broken build never produces a half-built or mislabeled binary:
#   e - exit immediately if any command exits with a non-zero status, instead of
#       carrying on (e.g. don't keep going if `go build` fails).
#   u - treat a reference to an unset variable as an error rather than silently
#       expanding to an empty string, so a typo in a variable name is caught
#       (e.g. an empty -ldflags value) instead of shipping bad build metadata.
set -eu

# =============================================================
# Build variables
# =============================================================

CMD_PATH=./cmd/dnska
DIAG=github.com/rokkerruslan/dnska/internal/diagnostics

# `vcs.modified` from debug.BuildInfo already reports a dirty tree, so we do not
# ask `git describe` for --dirty here (it would double up as "hash-dirty (dirty)").
BuildVersion=$(git describe --tags --always)
BuildHost=${HOSTNAME:-}
if [ -z "$BuildHost" ]; then
  BuildHost=$(hostname 2>/dev/null || true)
fi
if [ -z "$BuildHost" ]; then
  BuildHost=$(uname -n 2>/dev/null || true)
fi
if [ -z "$BuildHost" ]; then
  BuildHost=unknown
fi

BuildMachineID=
if [ -f /etc/machine-id ]; then
  BuildMachineID=$(cat /etc/machine-id)
elif [ "$(uname)" = "Darwin" ]; then
  BuildMachineID=$(ioreg -rd1 -c IOPlatformExpertDevice | awk -F'"' '/IOPlatformUUID/ {print $4}')
fi

LD_FLAGS="\
-X \"${DIAG}._Version=${BuildVersion}\" \
-X \"${DIAG}._Host=${BuildHost}\" \
-X \"${DIAG}._MachineID=${BuildMachineID}\" \
-X \"${DIAG}._User=${USER:-}\""

# =============================================================
# Commands
# =============================================================

install() {
  CGO_ENABLED=0 go install -tags netgo,osusergo -ldflags="${LD_FLAGS}" $CMD_PATH
}

build() {
  CGO_ENABLED=0 go build -tags netgo,osusergo -ldflags="${LD_FLAGS}" $CMD_PATH
}

run() {
  # Run the server straight from source. Any extra arguments ("$@") are
  # forwarded to `dnska app` (e.g. --endpoints-file-path, --dump-dir).
  go run $CMD_PATH app "$@"
}

# =============================================================
# Entrypoint
# =============================================================

USAGE=$(
  cat <<-END
app - entry point for control

USAGE
  ./scripts/app.sh COMMAND

COMMANDS

  install          install binary to GOBIN path
  build            build binary and store in current directory
  run              run "dnska app" from source (extra args are forwarded)
  help             print this docs

EXAMPLES

  $ ./scripts/app.sh build
  $ ./scripts/app.sh run --endpoints-file-path ./configs/endpoints.local.toml

END
)

if [ ! -f ".root" ]; then
  echo "Script must be run from root project directory."
  echo 'We detect the ".root" file to verify that.'
  echo
  exit 1
fi

# ${1:-} expands to the first argument, or an empty string if it is unset.
# The `:-` default is required because `set -u` aborts on a bare `$1` when the
# script is run with no arguments; the empty string then falls through to the
# help branch below.
case "${1:-}" in
i | install)
  install
  ;;
b | build)
  build
  ;;
r | run)
  shift
  run "$@"
  ;;
h | help | "")
  echo "$USAGE"
  ;;
*)
  echo "$1 command does not exist, check help"
  ;;
esac
