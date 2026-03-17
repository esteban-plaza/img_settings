# Command: commit

## Usage

```
/commit
```

## What This Command Does

Creates a well-formed git commit for the current staged changes, following the Conventional Commits specification used by this project's semantic-release setup.

## Instructions for the AI

1. Run `git diff --staged` to review exactly what is staged.
2. Determine the correct commit type:
   - `feat` — new user-visible feature
   - `fix` — bug fix
   - `docs` — documentation only
   - `chore` — build, CI, or tooling change
   - `refactor` — code restructure, no behaviour change
   - `test` — adding or fixing tests
   - `perf` — performance improvement
3. Determine the scope (the affected package or area):
   - `lib`, `gui`, `cli`, `ci`, `deps`, `docs`, `build`
4. Write a subject line: imperative mood, max 72 characters, no period.
5. If the change is substantial, add a body explaining **why** (not what).
6. If it introduces a breaking change, add `BREAKING CHANGE: <description>` in the footer.
7. Propose the commit message to the user before executing.
8. Run:
   ```bash
   git commit -m "<type>(<scope>): <subject>"
   ```

## Examples

```
feat(cli): add --quality flag for JPEG output quality
fix(lib): correct EXIF orientation handling for tag value 8
chore(ci): cache Go modules between CI runs
test(lib): add table-driven tests for OutputPath
docs: update USER_GUIDE with ARW troubleshooting steps
```
