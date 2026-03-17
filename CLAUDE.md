# Claude Configuration — img-settings

## Core Development Rules

All rules for this project are defined in the spec files below. Read them at the start of every session before making any change.

See @ai-specs/specs/base-standards.mdc for core principles (TDD, baby steps, English only, incremental commits).

See @ai-specs/specs/go-standards.mdc for Go-specific standards: package architecture, error handling, testing requirements, branch naming, commit format, CI/CD pipeline, and build system.

See @ai-specs/specs/development_guide.md for build commands, test commands, release workflow, and the Definition of Done checklist.

## Available Commands

| Command | Description |
|---------|-------------|
| `/plan-ticket <id>` | Generate a step-by-step implementation plan saved to `ai-specs/changes/<id>.md` |
| `/develop @<plan.md>` | Execute an implementation plan with TDD (failing test first) |
| `/commit` | Create a well-formed Conventional Commits message |

See @ai-specs/.commands/plan-ticket.md for `/plan-ticket` instructions.
See @ai-specs/.commands/develop.md for `/develop` instructions.
See @ai-specs/.commands/commit.md for `/commit` instructions.

## Agent

See @ai-specs/.agents/go-developer.md for the Go developer agent role and workflow.

## Project Quick Reference

```
lib/    # Core logic (EXIF, decode, resize, watermark) — CGO=0 safe, pure Go
gui/    # Fyne v2 GUI — CGO required
cli/    # CLI — CGO=0 single static binary
dist/   # Build outputs (git-ignored)
docs/   # User guides (EN + ES)
```

**Key rule**: `lib/` must never import `gui/` or `cli/`.
