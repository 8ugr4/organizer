## 1.1.2 (2026-04-27)

## 1.1.1 (2026-04-27)

## 1.1.0 (2026-02-07)

### Feat

- sort files depending on their year or year/month

### Refactor

- consolidate same calls in async and sync
- get rid of stupidly huge if tree
- move exif logic to it's own file
- get functions should return errors
- getFileDate function now works correctly. maybe squash this into other feat func

## 1.0.0 (2026-02-05)

### Feat

- introduce org-dir subcommand, preparation for sort-img
- sub-dir option is now possible
- allow async copy
- allow dry-run
ci: add github actions, pre-commit file
build: rename tool, add test target
docs: improve README

### Fix

- default async value should be false
- separate sub-dirs now work
- use correct field
- async works correctly and faster.

### Refactor

- add pre-commit hooks and apply suggestions
- general cleanup and Makefile improvements
- cleanup
- add verbose, async flags, add hyperfine target
- improve processdir func
- split function
- clean-up code, multiple changes
- workaround for unspecified csv logger, fix this later
- auto-set dst_path if not set by user
- do not log recursive processes percentage.
