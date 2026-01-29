# Regex Path Sorting Fix

## Problem

NGINX Gateway Fabric did not order regex (`RegularExpression`) location blocks by path length. Since NGINX evaluates regex locations in config file order (first match wins), a catch-all pattern like `/.*` could match before a more specific pattern like `/perform.*`.

## Solution

Modified the `sortPathRules` function to sort regex paths by descending length, ensuring longer/more specific patterns appear first in the generated NGINX config.

## Changes

### `internal/controller/state/dataplane/sort.go`

Added two functions:

```go
func sortPathRules(pathRules []PathRule)
func pathRuleLess(i, j PathRule) bool
```

Sorting logic:
- **Regex paths**: Sorted by descending length (longer = higher priority)
- **Same-length regex**: Alphabetical for deterministic ordering
- **Non-regex paths**: Alphabetical by path, then by PathType (unchanged from original)

### `internal/controller/state/dataplane/configuration.go`

Replaced inline sort with `sortPathRules()` call.

## Result

Before:
```nginx
location ~ /.* { ... }        # Matches everything first
location ~ /perform.* { ... } # Never reached for /perform_logout
```

After:
```nginx
location ~ /perform.* { ... } # Matches /perform_logout
location ~ /.* { ... }        # Fallback
```

## Tests

- `sort_test.go`: Unit tests for `sortPathRules`
- `servers_test.go`: Integration test verifying location order in generated config
