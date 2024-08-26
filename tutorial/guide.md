# First-time Perennial setup
* Install Go (version 1.22.5 as of writing): https://go.dev/doc/install
* To install Coq, first install [opam](https://opam.ocaml.org/).  Here's a [link
  to opam's "how to install" page](https://opam.ocaml.org/doc/Install.html).
* Install Coq (version 8.19.2 as of writing) as described [here](https://coq.inria.fr/opam-using.html).
* Set up an editor for Coq as described [here](https://coq.inria.fr/user-interfaces.html).
* Clone the Perennial repository: https://github.com/mit-pdos/perennial
* In the Perennial repo, run `git submodule update --init --recursive` to
  download its dependencies.
* Build the basic Perennial libraries by running `make -j4 src/goose_lang/prelude.vos`.
  You may want to adjust the number of parallel jobs (`-j4`) based on
  the number of cores in your machine.

# Using Perennial
In order to be able to "step through" a file, all of the files that it depends
on must first be compiled. There are two ways of building a file and its
dependencies: with proof-checking or without proof-checking.  For example,
suppose we want to work on the proof for
[gokv/tutorial/basics](https://github.com/mit-pdos/gokv/blob/main/tutorial/basics/basics.go);
the proof lives in `src/program_proof/tutorial/basics/proof.v`.

## Light (without proof-checking) build of a file
This builds the given file and its transitive dependencies.  It skips checking
proofs (i.e. stuff ending with `Qed.` or `Admitted.`), but still compiles
theorem statements so that files can be imported, and the theorems can be used.
This is convenient during development.
Example:
```
make -j4 src/program_proof/tutorial/basics/proof.vos
```

## Full (with proof-checking) build of a file
This fully builds the given file and its dependencies. It actually checks all
proofs, so it takes longer than the "light build". You should do a full build in
case you change a definition in a file that lots of other files import to make
sure nothing is broken. We typically do full builds when changing low-level
definitions, and as part of the CI process for the Perennial repo.
Example:
```
make -j4 src/program_proof/tutorial/basics/proof.vo
```

# Translating Go code for proving with Perennial

Goose is a tool to translate Go code (`.go` files) into a formal "GooseLang"
program in Coq (represented using definitions in `.v` files). This is how
Coq proofs with Perennial are connected to actual Go code.

For example, to translate the `gokv/tutorial/basics` code, you can follow
these steps:

- Clone https://github.com/goose-lang/goose somewhere.
- Clone https://github.com/mit-pdos/gokv somewhere.
- From the root of the Perennial repository, run
  ```
  ./etc/update-goose.py --compile --goose /path/to/goose --gokv /path/to/gokv
  ```

This will update the files in the `external/Goose/github_com/mit_pdos/gokv`
directory in Perennial, including
[`external/Goose/github_com/mit_pdos/gokv/tutorial/basics.v`](https://github.com/mit-pdos/perennial/blob/master/external/Goose/github_com/mit_pdos/gokv/tutorial/basics.v) corresponding to the [gokv/tutorial/basics](https://github.com/mit-pdos/gokv/blob/main/tutorial/basics/basics.go) package.

We typically check in the generated Goose code into the Perennial
repository, so unless you changed the `gokv` source code, running the
above `update-gooes.py` command should not change the generated files.

# Warm-up exercise 1

Build the dependencies for the `basics.v` proof by running `make
src/program_proof/tutorial/basics/proof.vos` in your Perennial
checkout.  Open `src/program_proof/tutorial/basics/proof.v`
in an editor set up with Coq, and step through the proofs in
that file.  It might be helpful to refer to the Go source code in
[basics.go](https://github.com/mit-pdos/gokv/blob/main/tutorial/basics/basics.go)
to understand what code is being verified.

Hint: You might find it helpful to this [Iris Proof Mode
reference](https://gitlab.mpi-sws.org/iris/iris/-/blob/master/docs/proof_mode.md)
to understand the Iris tactics and their syntax.

# Warm-up exercise 2

Add methods to `basics.go` for unregistering a key from the map (the
opposite of the existing `registerLocked` and `Register` methods).
Re-generate the Goose code once you've modified `basics.go`.  State
theorems specifying your new methods in `basics/proof.v`, and prove
those theorems.

Hint: in Coq, `delete k m` is the deletion of key `k` from gmap `m`.
