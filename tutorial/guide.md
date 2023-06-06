# Install Go (version 1.18 or newer): https://go.dev/doc/install

--------------------------------------------------------------------------------

# First-time Perennial setup:
* To install Coq, first install [opam](https://opam.ocaml.org/).  Here's a [link
  to opam's "how to install" page](https://opam.ocaml.org/doc/Install.html).
* Install Coq version 8.16.1 as described [here](https://coq.inria.fr/opam-using.html).
* Set up an editor for Coq as described [here](https://coq.inria.fr/user-interfaces.html).
* Clone the Perennial repository: https://github.com/mit-pdos/perennial
* In the Perennial repo, run `git submodule update --init --recursive` to
  download some dependencies.


# Working with Perennial
In order to be able to "step through" a file, all of the files that it depends
on must first be compiled. There are two ways of building a file and its
dependencies. After building `src/program_proof/tutorial/basics/proof.v` (as
described below), try opening it an editor set up with Coq and stepping through
it.

## Light build of a file:
This builds the given file and its transitive dependencies.  It skips checking
proofs (i.e. stuff ending with `Qed.` or `Admitted.`), but still compiles
theorem statements so that files can be imported, and the theorems can be used.
This is convenient during development.
Example:
```
make src/program_proof/tutorial/basics/proof.vos
```

## Full build of a file:
This fully builds the given file and its dependencies. It actually checks all
proofs, so it takes longer than the "light build". You should do a full build in
case you change a definition in a file that lots of other files import to make
sure nothing is broken.
Example:
```
make src/program_proof/tutorial/basics/proof.vo
```

--------------------------------------------------------------------------------

# Re-generating translation of code:

Goose is a tool to translate Go code (`.go` files) into a formal "GooseLang"
program in Coq (some definitions in `.v` files). This is how Coq proofs with
Perennial are connected to actual Go code.

Clone [https://github.com/tchajed/goose] somewhere.
Clone [https://github.com/mit-pdos/gokv] somewhere.
Then, from the root of the Perennial repository, run
```
./etc/update-goose.py --compile  --goose /path/to/goose --gokv /path/to/gokv
```
This will update the files in the `external/Goose/github_com/mit_pdos/gokv`
directory in Perennial.
