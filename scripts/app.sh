#!/usr/bin/env sh

# =============================================================
# Static variables
# =============================================================

CMD_PATH=./cmd/dnska

BuildVersion=$(git describe --tags)
BuildTimestamp=$(date +%s)
BuildHost=$(hostname)

if [ -f /etc/machine-id ]; then
  BuildMachineID=$(cat /etc/machine-id)
fi

BuildIsDirty=false

[ -z "$(git status -s)" ] || BuildIsDirty=true

LD_FLAGS=$(cat <<-END
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._Version=${BuildVersion}
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._Timestamp=${BuildTimestamp}
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._Host=${BuildHost}
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._MachineID=${BuildMachineID}
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._User=${USER}
    -X=github.com/rokkerruslan/dnska/internal/diagnostics._IsDirty=${BuildIsDirty}
END
)

# =============================================================
# Commands
# =============================================================

install() {
  GGO_ENABLED=0 go install -tags netgo -ldflags="${LD_FLAGS}" $CMD_PATH
}

build() {
  CGO_ENABLED=0 go build -tags netgo -ldflags="${LD_FLAGS}" $CMD_PATH
}

# =============================================================
# Entrypoint
# =============================================================

USAGE=$(
  cat <<-END
app - entry point for control

USAGE
  ./app COMMAND

COMMANDS

  install          install binary to GOBIN path
  build            build binary and store in current directory
  help             print this docs

EXAMPLES

  $ ./app test

END
)

if [ ! -f ".root" ]; then
  echo "Script must be run from root project directory."
  echo 'We detect the ".root" file to verify that.'
  echo
  exit 1
fi

case $1 in
i | install)
  install
  ;;
b | build)
  build
  ;;
h | help | "")
  echo "$USAGE"
  ;;
*)
  echo "$1 command does not exist, check help"
  ;;
esac
