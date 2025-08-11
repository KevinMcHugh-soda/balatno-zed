# Agent Guidelines

These instructions apply to the entire repository.

## Code Style
- Format Go code with `go fmt` before committing.
- Document exported functions and types where appropriate.

## Testing
- Run all tests and ensure they pass before committing.  
  ```bash
  go test ./...
  ```

Tests cover the Ante/Blind system and poker hand evaluation

## Architecture
- Preserve the clean separation between game logic and user interface:
  - Game engine remains UI‑agnostic.
- Communication occurs through event emitters/handlers. 
- See docs/ARCHITECTURE.md for details

## Documentation
- Update relevant files under docs/ and the README when behavior or configuration changes.
- Use Markdown with descriptive headings.

## Commit & PR Conventions
- Start commits with a present tense noun describing what the commit adds (e.g. Refactors the FooBar service)
- Be concise and crisp
- In pull requests, summarize the change and list tests run.


This file guides contributors on formatting, testing, architectural boundaries, documentation, and PR etiquette, aligning with existing project practices.


