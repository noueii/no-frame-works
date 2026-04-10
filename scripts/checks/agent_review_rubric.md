# Convention Review Rubric

You are reviewing a pull request diff against the backend framework conventions.
Your job is to post a review identifying any violations of the judgment-based rules
that automated checks cannot catch.

## What You Receive

- The full PR diff
- This rubric

## Rules to Check

Review the diff for violations of these rules. Only flag issues you are CONFIDENT about.
If a rule doesn't apply to the changes in this diff, skip it.

### 1. Service calls Validate() first
Every service method should call `req.Validate()` as its first meaningful operation.
- Flag if: Validate() is called after other operations, or not called at all
- Skip if: No new service methods in the diff

### 2. Handlers only parse, never validate
HTTP handlers should decode the request and call the service. No business logic.
- Flag if: Handler contains validation, DB calls, or business rules
- Skip if: No handler changes in the diff

### 3. Permission layer wraps correctly
The permission middleware should implement the same interface and call authorize() before delegating.
- Flag if: Permission checking is inline in the service, or layer is missing for new API methods
- Skip if: No new API methods

### 4. Constructor injection
Dependencies must be injected through constructors, never created internally.
- Flag if: A service or handler calls New() or creates its own dependencies
- Skip if: No new structs or constructors

### 5. Types owned by correct module
No shared domain packages. Types live in the module that owns them.
- Flag if: Types are imported from another module's domain/ or from a shared package
- Skip if: No new types

### 6. Provider pattern for external deps
External SDKs should be behind provider interfaces.
- Flag if: External SDK is called directly from service code
- Skip if: No external service integration

### 7. Domain errors in domain/errors.go
Module-specific errors should be sentinel errors in domain/errors.go.
- Flag if: Errors created inline with fmt.Errorf for domain concepts, or defined outside domain/
- Skip if: No new error types

## Output Format

Respond with ONLY this JSON, no other text:

```json
{
  "issues": [
    {
      "rule": "validate_first",
      "severity": "error",
      "file": "backend/internal/modules/orders/service/create/create_order.go",
      "line_hint": "func (s *Service) CreateOrder",
      "message": "CreateOrder does not call req.Validate() before accessing the repository"
    }
  ],
  "summary": "Found 1 convention violation in the orders module",
  "clean": false
}
```

If everything looks good:
```json
{
  "issues": [],
  "summary": "No convention violations found",
  "clean": true
}
```

## Severity Levels
- `error` — Clear violation that should be fixed before merge
- `warning` — Looks off but might be intentional, reviewer should decide
