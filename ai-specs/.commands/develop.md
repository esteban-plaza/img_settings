# Command: develop

## Usage

```
/develop @ai-specs/changes/<ticket-id>.md
```

Example: `/develop @ai-specs/changes/IMG-42.md`

## What This Command Does

Executes the implementation plan referenced by the given file, following each step exactly as described. Uses TDD: writes the failing test first, then implements the code to make it pass.

## Instructions for the AI

1. **Read the plan file** completely before writing any code.
2. **Execute steps in order** — never skip or reorder steps.
3. **TDD discipline**:
   - Write the test first.
   - Run `go test ./...` — confirm it fails.
   - Implement the minimum code to make it pass.
   - Run `go test ./...` — confirm it passes.
4. **After each step**:
   - Run `go vet ./...` — must be clean.
   - Commit with a Conventional Commits message referencing the ticket.
5. **Verify coverage** after all steps:
   ```bash
   go test ./lib/ -coverprofile=coverage.out
   go tool cover -func=coverage.out
   ```
   Coverage for `lib/` must be ≥ 80%.
6. **Update documentation** if the plan requires it:
   - `docs/USER_GUIDE.md`
   - `docs/USER_GUIDE.es.md`
   - `README.md`
7. **Do not push** without explicit user confirmation.
