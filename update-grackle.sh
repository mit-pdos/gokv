#!/usr/bin/env bash

if [[ -z "$IN_NIX_SHELL" ]]; then 
    # Not in nix shell, use the default go installs
    GRACKLE="go run github.com/mjschwenne/grackle/cmd/grackle@latest"
else 
    # In a nix shell, expect grackle to be on the PATH and use that
    GRACKLE="grackle"
fi

# Have I mentioned that I dislike bash? Just look at this arcane
# and archaic syntax
compile_grackle () {
    CWD=$(pwd)
    cd "$1" || return
    go install ./cmd/grackle
    cd "$CWD" || exit
}

# Run grackle on the input go package.
#
# We will assume that:
# 1. The proto file is in this directory
# 2. We only want to output go code
# 3. The go code should be output into this directory
# 4. The desired go package matches the directory structure
run_grackle () {
    $GRACKLE --go-output-path "$1" --go-package "github.com/mit-pdos/gokv/$1" "$1"
}

ARGS=$(getopt -o "c:h" --long "compile-grackle:,help" -- "$@")

eval set -- "$ARGS"
while [ $# -ge 1 ]; do
    case "$1" in
        -c | --compile-grackle)
            echo "compiling grackle $2"
            compile_grackle "$2"
            shift
            ;;
        -h | --help)
            cat <<EOF
usage: update-grackle.sh [--compile-grackle <grackle repo> | -c <grackle repo>] [--help | -h]

Calls grackle on all go modules known to have proto files for grackle usage.

--compile-grackle [-c] : Takes the path to the grackle repository, recompiles and installs grackle
EOF
            shift
            exit 0
            ;;
        --)
            shift
            break
            ;;
    esac
    shift
done

grackle_packages=(
    "tutorial/kvservice"
    "tutorial/lockservice"
    "tutorial/objectstore/chunk"
    "tutorial/objectstore/dir"
)

for gopkg in "${grackle_packages[@]}"; do
    run_grackle "$gopkg"
done
