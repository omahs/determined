We should present an example + reasoning behind this + an importance score out of 5. 

# Error handling

## Wrap errors from external packages

We should almost always wrap errors when dealing with external packages. 

```diff
if err := query.Scan(ctx); err != nil {
-   return err
+   return fmt.Errorf("running query: %w", err)
}
```

```TODO on exact rules especially around json / yaml```

4 / 5 importance.

1. Very important for making our logs readable. Seeing "no rows" errors in master logs and having no idea where it came from is disappointing. 

## Don't start wrapping error message with "error"

```diff
if err := query.Scan(ctx); err != nil {
- return fmt.Errorf("error running query: %w", err)
+ return fmt.Errorf("running query: %w", err)
}
```

1 / 5 importance.

1. To make user facing error messages are consistent.

## Prefer fmt.Errorf to errors.Wrap

```diff
if err := something(); err != nil {
-    return errors.Wrap(err, "something")
+    return fmt.Errorf("something: %w", err)
}
```

2 / 5 importance.

1. Consistency throughout the code base. 
2. The alternative ```errors.Wrap``` we are currently using a package that has no maintainers and is archive mode. 
3. Using ```errors.Wrap``` can have some really subtitle bugs for example this function is wrong and is wrong in a really hard to debug way. The following function will return nil.

```go
err := returnNilError()
if err != nil {
    return err
}

err2 := returnNonNilError()
if err2 != nil {
    // using wrong "err".
    return errors.Wrap(err, "this whole return will be nil")
}
```

