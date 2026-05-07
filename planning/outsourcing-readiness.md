# Outsourcing Readiness

## Summary
This tool is small and coherent enough to hand off for routine maintenance, but not yet ideal for fully independent outsourcing.

## What It Already Does
- Fetches Ultimate Guitar tabs by id or song URL.
- Saves output under an artist folder.
- Supports local transposition to a target key.
- Supports plain output by default and optional markup output.
- Lets the caller choose the parent output directory.

## What Makes It Reasonably Handoffable
- The codebase is single-purpose and small.
- The current behavior is already verified manually against live tabs.
- Dependencies are minimal: one existing scraper library plus the Go standard library.

## What Still Makes It Fragile
- Key inference is heuristic when UG metadata is wrong or missing.
- There are no automated tests for URL parsing, transposition, or file output.
- Live fetches depend on Ultimate Guitar availability and API behavior.
- The project has no release packaging, install script, or documented build workflow yet.

## Ready For Outsourcing?
- For small changes: yes.
- For broad refactors or reliability work: not yet.

## Recommendation
- Outsource only narrow tasks with clear acceptance criteria.
- Keep ownership of anything involving key inference, output formatting, or scraper behavior until tests are added.
