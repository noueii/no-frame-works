# Common Mistake Patterns Rubric

You are reviewing Go backend code for common mistakes that AI-generated code frequently produces. These are not architectural concerns — they are concrete, pattern-matchable errors.

This rubric will grow over time as new patterns are discovered.

## Rules

### 1. Use errors.New and fmt.Errorf correctly

Use `errors.New()` for static error messages (sentinel errors). Use `fmt.Errorf()` only when wrapping an existing error with `%w` or adding dynamic context. Never use `fmt.Errorf()` for static strings.

❌ Wrong — fmt.Errorf for static strings:
```go
return fmt.Errorf("user not found")
return fmt.Errorf("validation failed")
return fmt.Errorf("unauthorized")
```

✅ Correct — errors.New for static strings:
```go
var ErrUserNotFound = errors.New("user not found")
return ErrUserNotFound
```

✅ Correct — fmt.Errorf for wrapping with context:
```go
return fmt.Errorf("failed to find user: %w", err)
return fmt.Errorf("insert post: %w", err)
```

### 2. Always wrap errors with %w, not %v or %s

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

### 3. Don't shadow sentinel errors with wrapping

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
