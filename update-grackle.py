#!/usr/bin/env python3

import argparse
import os
import subprocess as sp

grackle_packages = [
    "tutorial/kvservice",
    "tutorial/lockservice",
    "tutorial/objectstore/chunk",
    "tutorial/objectstore/dir",
]

# Use $PATH grackle if in a nix shell, otherwise the standard go install
if "IN_NIX_SHELL" in os.environ:
    grackle_cmd = ["grackle"]
else:
    grackle_cmd = ["go", "run", "github.com/mjschwenne/grackle/cmd/grackle@latest"]


def compile_grackle(grackle_repo):
    original_working_directory = os.getcwd()
    os.chdir(grackle_repo)
    sp.run(["go", "install", "./cmd/grackle"])
    os.chdir(original_working_directory)


def run_grackle(pkg, debug=False):
    opts = [
        "--go-output-path",
        pkg,
        "--go-package",
        f"github.com/mit-pdos/gokv/{pkg}",
        pkg,
    ]
    if debug:
        opts.insert(4, "--debug")
        print(opts)
    try:
        sp.run(
            grackle_cmd + opts,
            timeout=1,
        )
    except sp.TimeoutExpired:
        print(f"Grackle timed out on package: {pkg}")


def main():
    parser = argparse.ArgumentParser(
        prog="update-grackle",
        description="Update the grackled files in gokv",
    )
    parser.add_argument(
        "-c",
        "--compile",
        help="Path to the grackle repository. If set, compile and install grackle from this repository",
    )
    parser.add_argument(
        "-d",
        "--debug",
        action="store_true",
        help="Print grackle output to stdout rather than update the files",
    )

    args = parser.parse_args()
    if args.compile is not None:
        compile_grackle(args.compile)

    for pkg in grackle_packages:
        run_grackle(pkg, args.debug)


if __name__ == "__main__":
    main()
