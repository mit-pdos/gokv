# Overview
The code for the replicated state machine library vRSM and the key-value store
vKV are in the `simplepb` directory. Inside, there are packages for
primary/backup replica servers (`pb`), a config service (`config`), a
storage library (`simplelog`), a clerk (`clerk`), and some applications
including the key-value store vKV (`apps`).

The proofs for all of this code are in the [Perennial
repository](https://github.com/mit-pdos/perennial/tree/master/src/program_proof/simplepb).

# Grove SOSP'23 artifact

The Grove artifact is in its own repository (with its own README)
[here](https://github.com/mit-pdos/grove-artifact). It includes a specific
commit of this repository as a git submodule. The artifact repository describes
how to run the experiments described in the SOSP'23 paper.
