# Research: what a git merge driver gets and must do

Official page: https://git-scm.com/docs/gitattributes, section "Defining a custom merge driver", read 2026-11-26T01:20:46Z with `curl -sSL --fail`; the HTML bytes have sha256 `3c2a1c79e123a33d7968c8725c361340ae9106232d13de4ddabbb9ecdc794be7`. Local git: 2.54.0 (Apple Git-157).

| Question | Quote |
| --- | --- |
| Where is the driver defined? | "The definition of a merge driver is done in the .git/config file, not in the gitattributes file" |
| What are the arguments? | "These three tokens are replaced with the names of temporary files that hold the contents of these versions when the command line is built." (`%O`, `%A`, `%B`; `%P` is "the pathname in which the merged result will be stored") |
| Where does the result go, what does the exit mean? | "The merge driver is expected to leave the result of the merge in the file named with %A by overwriting it, and exit with zero status if it managed to merge them cleanly, or non-zero if there were conflicts." |
| A crash? | "When the driver crashes (e.g. killed by SEGV), it is expected to exit with non-zero status that are higher than 128, and in such a case, the merge results in a failure (which is different from producing a conflict)." |
| Merges of merge bases | "The merge.*.recursive variable specifies what other merge driver to use when the merge driver is called for an internal merge between common ancestors, when there are more than one. When left unspecified, the driver itself is used for both internal merge and the final merge." |

What this decides: a non-zero exit below 128 with `%A` untouched is a conflict that shows only our side, so every refusal first writes git's text merge into `%A` (a Rust panic exits 101, which git reads as a conflict, so the work is run inside `catch_unwind`). The driver receives file versions only as the three temporary files; the page promises nothing about the work tree. The earlier driver's docstring records a measurement with git 2.54: git calls the driver before it writes any merged file to the work tree (for merge, rebase, cherry-pick and `merge-tree`, which has no work tree to write). So a hash cannot be taken from a file on disk at that moment.

## Not read

- A fresh measurement of the "before the work tree is written" order: not run here; the tests of this SPEC fail if hashes come from the work tree, because a merged hash differs from our side's hash.
- How GitHub's server merges: it runs no custom driver (stated in the repository's own document, not re-read from GitHub here).
