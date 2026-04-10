# Common Mistake Patterns Rubric

You are reviewing Go backend code for common mistakes that AI-generated code frequently produces. These are not architectural concerns — they are concrete, pattern-matchable errors.

This rubric will grow over time as new patterns are discovered.

## Rules

### 1. Never use fmt.Errorf — always use errors.Errorf

This project uses `github.com/go-errors/errors` for stack traces. All error wrapping and creation must use this package, never `fmt.Errorf`. Import `"github.com/go-errors/errors"` instead of stdlib `"errors"`.

❌ Wrong — fmt.Errorf:
```go
return fmt.Errorf("failed to find user: %w", err)
return fmt.Errorf("insert post: %w", err)
return fmt.Errorf("user not found")
```

✅ Correct — errors.Errorf:
```go
return errors.Errorf("failed to find user: %w", err)
return errors.Errorf("insert post: %w", err)
```

✅ Correct — errors.Errorf for sentinel errors:
```go
var ErrUserNotFound = errors.Errorf("user not found")
```

### 2. Never import stdlib "errors" — use go-errors

❌ Wrong:
```go
import "errors"
```

✅ Correct:
```go
import "github.com/go-errors/errors"
```

The `go-errors/errors` package is a drop-in replacement — it has `errors.New`, `errors.Is`, `errors.As`, `errors.Errorf`, and adds stack traces.

### 3. Use errors.Errorf for static error messages, not errors.New

Use `errors.Errorf()` for sentinel error declarations. `errors.New()` from go-errors returns `*errors.Error` (not `error`), while `errors.Errorf()` returns `error`.

❌ Wrong:
```go
var ErrNotFound = errors.New("not found")
```

✅ Correct:
```go
var ErrNotFound = errors.Errorf("not found")
```

### 4. Always wrap errors with %w, not %v or %s

When wrapping errors, use `%w` so the error chain is preserved and callers can use `errors.Is()` and `errors.As()`. Using `%v` or `%s` loses the error chain.

❌ Wrong:
```go
return fmt.Errorf("failed to create post: %v", err)
return fmt.Errorf("query failed: %s", err.Error())
```

✅ Correct:
```go
return fmt.Errorf("failed to create post: %w", err)
```

### 5. Don't shadow sentinel errors with wrapping

When a service returns a domain error that the handler needs to match with `errors.Is()`, don't double-wrap it in a way that hides the original.

❌ Wrong — wrapping a sentinel error loses it:
```go
if existing == nil {
    return post.PostView{}, fmt.Errorf("post lookup failed: %w", post.ErrPostNotFound)
    // handler can still match this, but the message is misleading — nothing "failed"
}
```

✅ Correct — return sentinel errors directly:
```go
if existing == nil {
    return post.PostView{}, post.ErrPostNotFound
}
```

✅ Also correct — wrap infrastructure errors, not domain errors:
```go
existing, err := repo.FindByID(ctx, req.ID)
if err != nil {
    return post.PostView{}, fmt.Errorf("failed to find post: %w", err)  // infra error, wrap is fine
}
if existing == nil {
    return post.PostView{}, post.ErrPostNotFound  // domain error, return directly
}
```

## Output Format

Only flag violations where you are at least 80% confident. Skip rules that don't apply to the diff. When in doubt, don't flag it.

For each violation, provide:
- Rule name
- File path and line
- The problematic code
- Brief explanation of what's wrong and the fix
