---
name: review
description: >
  Comprehensive code review for the no-frame-works Go backend. Use when asked to
  review code, review PR, or check changes against architecture rules.
---

# Architecture Code Review

Review the provided Go code against the no-frame-works layered architecture rules.
Apply the appropriate rubric from `.agents/rules/` based on the layer.

## Review Targets

Analyze files created or modified in this session. If no specific files are provided,
review recent git changes:

```bash
git diff --name-only HEAD
git diff --cached --name-only
```

## Rubrics to Apply

### 1. Handler Layer (`.agents/rules/handler.md`)
- Only oapi.* + API contract types
- No domain.* or model.* imports
- Pure transformers, no business logic

### 2. Service Layer (`.agents/rules/service.md`)
- req.Validate() called FIRST
- req.CheckPermission() called SECOND
- Returns view types, not domain models
- One function per file

### 3. Repository Layer (`.agents/rules/repository.md`)
- No raw SQL — go-jet only
- MODEL() with MutableColumns for updates
- toModel/toDomain in repository

### 4. Domain Layer (`.agents/rules/domain.md`)
- No infrastructure imports
- Pure business logic only
- Sentinel errors in domain/errors.go

### 5. Flow Integrity (`.agents/rules/flow.md`)
- All steps in request lifecycle present
- Correct types at each boundary
- No layer bypassing

### 6. Common Mistakes (`.agents/rules/patterns.md`)
- `fmt.Errorf` → `errors.Errorf`
- stdlib errors → go-errors
- `%w` wrapping required

### 7. Caveman Review Format (`.agents/caveman-review.md`)
- Terse, actionable comments
- Severity: 🔴🟠🟡🔵❓
- Format: `L<line>: <problem>. <fix>.`

## Output Format

Start with overall verdict, then findings:

```
## Review: <n> file(s)

### Verdict
[✅ Clean] / [⚠️ Needs work] / [❌ Significant violations]

### Findings
🔴 bug: file.go:L42 — Problem. Fix.
🟠 conv: file.go:L88 — Violation. Correct pattern.
🟡 risk: file.go:L23 — Fragile. Alternative.
🔵 nit: file.go:L55 — Style.
❓ q: file.go:L60 — Question.

### Layer Breakdown
- Handlers: <n> violation(s)
- Services: <n> violation(s)
- Repositories: <n> violation(s)
- Domain: <n> violation(s)
- Common: <n> violation(s)
```

## Severity Guide

| Prefix | Meaning | When to use |
|--------|---------|-------------|
| 🔴 bug | Will cause incident | Broken behavior, crashes |
| 🟠 conv | Architecture violation | Wrong layer, wrong type, skipped step |
| 🟡 risk | Works but fragile | Race conditions, missing null check |
| 🔵 nit | Style | Naming, formatting, minor issues |
| ❓ q | Question | Need clarification from author |