## 2025-02-18 - Repeated Regex Compilation in Go
**Learning:** Found multiple instances of `regexp.MustCompile` being called inside frequently invoked functions (like `ParseContent` in `pkg/visual/parser.go`). This leads to unnecessary CPU cycles and memory allocations on the hot path. Go's compiled `regexp.Regexp` objects are safe for concurrent use.
**Action:** Always move `regexp.MustCompile` calls to package-level variables (`var (...)`) so they are compiled exactly once at application startup.
