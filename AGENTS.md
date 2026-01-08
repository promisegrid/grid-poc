# Repository Guidelines

## Project Structure & Module Organization
- Root Go module lives at the repo root (`go.mod`) with core sources like `interfaces.go` and tests like `interfaces_test.go`.
- `x/` contains experimental prototypes and notes; many subfolders are standalone Go modules (`x/<name>/go.mod`).
- Documentation and exploratory writeups are primarily in `README.md` and `x/**/*.md`.
- Local Grokker state (e.g., `.grok`) and generated binaries should stay uncommitted.

## Build, Test, and Development Commands
- `go test ./...` runs the root module test suite.
- `go test -v ./...` runs tests verbosely (matches the default in `devloop.sh`).
- `./devloop.sh` watches for file changes, runs tests, and emits audio cues for pass/fail.
- For experiments, run tests within the module directory, e.g. `cd x/ipld-path && go test ./...`.

## Coding Style & Naming Conventions
- Go code follows standard `gofmt` formatting (tabs, canonical import grouping).
- Package names are short and lower-case.
- Keep edits small and focused; do not remove comments or docs—update them when they drift.

## Testing Guidelines
- Use Go’s standard `testing` package with `*_test.go` filenames.
- Prefer deterministic tests; avoid network calls unless explicitly required.
- Table-driven tests are encouraged when validating multiple cases.

## Task Tracking (TODO)
- Track tasks and plans in `TODO/`, with a `TODO/TODO.md` index of small tasks.
- Number TODOs with zero-padded IDs (e.g., `007`), don’t renumber, and mark completed items as `DONE` after the number.

## Commit & Pull Request Guidelines
- Recent history uses short, imperative one-line subjects (often lowercase); use concise, imperative, capitalized subjects going forward.
- Commit bodies include a section per changed file with bullet summaries.
- PRs should include a concise summary, relevant test commands run, and linked issues; include before/after notes when behavior or output changes.
