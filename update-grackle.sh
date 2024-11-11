#!/usr/bin/env sh

# Have I mentioned that I dislike bash? Just look at this arcane
# and archaic syntax
function compile_grackle () {
    CWD=$(pwd)
    cd $1
    go install ./cmd/grackle
    cd $CWD
}

# Run grackle on the input go package.
#
# We will assume that:
# 1. The proto file is in this directory
# 2. We only want to output go code
# 3. The go code should be output into this directory
# 4. The desired go package matches the directory structure
# 5. Grackle is on your $PATH
function run_grackle () {
    grackle --go-output-path $1 --go-package "github.com/mit-pdos/gokv/$1" $1
}

ARGS=$(getopt -o "c:g:h" --long "compile-goose:,compile-grackle:,help" -- "$@")

eval set -- "$ARGS"
while [ $# -ge 1 ]; do
    case "$1" in
        -g | --compile-grackle)
            echo "compiling grackle $2"
            compile_grackle $2
            shift
            ;;
        -h | --help)
            cat <<EOF
usage: update-grackle.sh [--compile-grackle <grackle repo> | -c <grackle repo>] [--help]

Calls grackle on all go modules known to have proto files for grackle usage.

--compile-grackle [-g] : Takes the path to the grackle repository, recompiles and installs grackle
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

for gopkg in ${grackle_packages[@]}; do
    echo $gopkg
    run_grackle $gopkg
done
