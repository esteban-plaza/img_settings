# Command: plan-ticket

## Usage

```
/plan-ticket <ticket-id>
```

Example: `/plan-ticket IMG-42`

## What This Command Does

Given a ticket ID and its description, this command generates a detailed, step-by-step implementation plan for a change to img-settings. The plan is saved to `ai-specs/changes/<ticket-id>.md`.

## Instructions for the AI

1. **Read the ticket**: Ask the user for the ticket description if not already provided.
2. **Read the relevant source files**: Use the architecture in `go-standards.mdc` to identify which packages and files are affected.
3. **Identify the scope**:
   - Which functions in `lib/` need to change?
   - Does the GUI processor need updating?
   - Does the CLI need a new flag or behaviour?
   - Are new tests required?
4. **Generate the plan** with the following sections:

---

### Plan Template

```markdown
# Implementation Plan: <ticket-id> — <short title>

## Overview
<2-3 sentence description of the change and why it is needed>

**Key principles applied:**
- TDD: failing test before implementation
- Baby steps: one commit per step
- lib/ isolation: core logic stays in lib/, no GUI imports

## Architecture Context

### Files to Modify
- `lib/watermark.go` — <reason>
- `gui/processor.go` — <reason>
- `cli/main.go` — <reason>

### New Files to Create
- `lib/watermark_test.go` — unit tests for new functions

## Implementation Steps

### Step 0: Create Feature Branch
git checkout main && git pull origin main
git checkout -b feature/<ticket-id>-<short-description>

---

### Step 1: Write Failing Tests
File: `lib/watermark_test.go`
Action: Add table-driven tests for the new/changed function.
Tests must fail before the implementation exists.

---

### Step N: Implement <function name>
File: `lib/watermark.go`
Action: <detailed description>
Expected function signature: ...
Implementation notes: ...

---

### Step N+1: Update GUI / CLI (if needed)
...

---

### Step N+2: Verify
go vet ./...
go test ./lib/ -coverprofile=coverage.out
go tool cover -func=coverage.out  # verify ≥ 80% for lib/

---

### Step N+3: Commit
git add <files>
git commit -m "feat(lib): <subject>"

## Definition of Done
- [ ] Tests written and passing
- [ ] go vet clean
- [ ] lib/ coverage ≥ 80%
- [ ] Exported symbols have doc comments
- [ ] USER_GUIDE updated if user-facing behaviour changed
- [ ] Commit follows Conventional Commits format
```

---

5. **Save the plan** to `ai-specs/changes/<ticket-id>.md`.
6. **Do not implement** — the plan is reviewed before execution.
