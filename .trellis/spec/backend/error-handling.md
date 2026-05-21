# Error Handling

> How errors are handled in this project.

---

## Overview

<!--
Document your project's error handling conventions here.

Questions to answer:
- What error types do you define?
- How are errors propagated?
- How are errors logged?
- How are errors returned to clients?
-->

(To be filled by the team)

---

## Error Types

<!-- Custom error classes/types -->

(To be filled by the team)

---

## Error Handling Patterns

<!-- Try-catch patterns, error propagation -->

(To be filled by the team)

---

## API Error Responses

<!-- Standard error response format -->

(To be filled by the team)

---

## Common Mistakes

<!-- Error handling mistakes your team has made -->

### Scenario: External connector fallback must only happen on typed not-found errors

#### 1. Scope / Trigger

- Trigger: External datasource connectors that probe multiple resource shapes/endpoints (for example, Notion page → data_source → database) need deterministic fallback behavior.
- Applies to: `internal/datasource/connector/*` clients and connector orchestration code.

#### 2. Signatures

- Connector fetch APIs should preserve typed sentinel errors with wrapping:
  - `FetchAll(ctx context.Context, config *types.DataSourceConfig, resourceIDs []string) ([]types.FetchedItem, error)`
  - `FetchIncremental(ctx context.Context, config *types.DataSourceConfig, cursor *types.SyncCursor) ([]types.FetchedItem, *types.SyncCursor, error)`
- Client helpers that probe remote resources should return project datasource sentinels such as:
  - `datasource.ErrResourceNotFound`
  - `datasource.ErrInvalidCredentials`
  - transient/rate-limit errors wrapped with `%w` when possible.

#### 3. Contracts

- Fallback to an alternate endpoint is allowed only when the current endpoint returned `datasource.ErrResourceNotFound`.
- Authentication, permission, rate-limit, and server errors must be returned to the caller and must not be treated as “try the next resource type”.
- Incremental sync discovery/query failures must fail the sync instead of returning an empty result set, because empty discovery can be interpreted as source deletion by downstream diff logic.
- Retried HTTP requests with JSON bodies must rebuild the request/body for each attempt; do not reuse a consumed `http.Request.Body`.

#### 4. Validation & Error Matrix

| Condition | Correct behavior |
| --- | --- |
| Page lookup returns `ErrResourceNotFound` | Probe database/data_source path if that connector supports it |
| Page/data_source lookup returns `ErrInvalidCredentials` or permission error | Return the error; no fallback |
| Database query returns rate-limit/server/transient error | Return the error; no empty successful sync |
| Incremental `SearchPages`/discovery fails | Return error and no new cursor |
| HTTP POST gets retried after 429/5xx | Recreate request with a fresh body reader |

#### 5. Good/Base/Bad Cases

- Good: Notion direct page ID returns 404 from `/pages/{id}`, then the connector probes data_source/database endpoints.
- Base: Notion database ID queries successfully and returns fetched records.
- Bad: Invalid Notion token returns 401 from `/data_sources/{id}` and the connector silently falls back to `/databases/{id}`, hiding the credential error.
- Bad: Incremental Notion discovery fails and the connector returns an empty page set, causing cursor-only pages to be marked deleted.

#### 6. Tests Required

- Unit test 404 fallback from the first endpoint to the alternate endpoint.
- Unit test non-404 errors do not fallback and preserve `errors.Is` against the sentinel error.
- Unit test incremental discovery/query errors return no items and no cursor.
- Unit test POST retry handlers receive the same JSON body on every attempt.

#### 7. Wrong vs Correct

##### Wrong

```go
page, err := client.GetPage(ctx, id)
if err != nil {
    // Any error is treated as “not a page”, including 401/429/500.
    return c.fetchDatabase(ctx, client, id, visited), nil
}
```

##### Correct

```go
page, err := client.GetPage(ctx, id)
if err != nil {
    if !errors.Is(err, datasource.ErrResourceNotFound) {
        return nil, fmt.Errorf("get page %s: %w", id, err)
    }
    return c.fetchDatabaseWithError(ctx, client, id, visited)
}
```
