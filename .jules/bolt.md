## 2025-02-18 - Repeated Regex Compilation in Go
**Learning:** Found multiple instances of `regexp.MustCompile` being called inside frequently invoked functions (like `ParseContent` in `pkg/visual/parser.go`). This leads to unnecessary CPU cycles and memory allocations on the hot path. Go's compiled `regexp.Regexp` objects are safe for concurrent use.
**Action:** Always move `regexp.MustCompile` calls to package-level variables (`var (...)`) so they are compiled exactly once at application startup.
## 2025-05-18 - Replacing string manipulation via closure in Regex logic
**Learning:** Found multiple instances of `regexp.ReplaceAllStringFunc` being called in string-heavy routines like Telegram Markdown parsers, resulting in the runtime allocating closures and heavily copying strings on each regex match (doubling execution time on string heavy code).
**Action:** Replace `ReplaceAllStringFunc` with simple `ReplaceAllString` where substitution expressions like `$1` work. For complex replacement logic (like substituting different formats per block item), instead use `FindAllStringSubmatchIndex` combined with a single pass through a `strings.Builder`. This avoids repetitive garbage collection and is exceptionally fast.

## 2025-06-25 - Unused db.QueryRow Leaks and Wastes
**Learning:** Found an unused `db.QueryRow` call assigned to a blank identifier `_` in `GetProfile`. `QueryRow` in `database/sql` executes the query immediately on the database connection. Discarding it without calling `.Scan()` wastes network I/O, database CPU cycles, and leaves a row open, potentially leaking connection pool resources.
**Action:** Always safely remove unused `db.QueryRow` executions entirely rather than silencing compilation errors with a blank identifier assignment.

## 2025-06-25 - Discarded JSON Marshaling
**Learning:** Found a useless `json.Marshal(p)` call where the byte slice error and result were discarded `_ = blob`. This unnecessarily invokes reflection logic, CPU processing, and heavy garbage collector allocations for absolutely no reason before generating the string prompt.
**Action:** Always scan formatting logic and remove totally useless data serialization steps that are blindly retained without being logged or stored.

## 2026-06-01 - Double regex evaluation
**Learning:** Found instances where `FindStringSubmatch` was used to check for match existence, and then `ReplaceAllString` was immediately called on the same regex. This evaluates the regular expression engine twice on the same text.
**Action:** Use `FindStringSubmatchIndex` or `FindAllStringSubmatchIndex` instead to find exactly where the match starts/ends and its submatches. Then simply rebuild the string using manual string concatenation or `strings.Builder`. This avoids double evaluation and is over 2x faster.
