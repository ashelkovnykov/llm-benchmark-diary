## Summary

This folder contains experiments that ask an LLM to generate code from scratch, with no existing codebase to leverage / modify.

## Table of Contents

- **[tui/](./tui/README.md)**: Experiment goal is to create a TUI that allows the user to load a locally-hosted SQL DB and visualize a simple summary of the contents. Tier 3 problem, possibly straying into tier 4: expected output is around 2000-2500 tokens.

## Appendix

### Size Classification

These experiments use the following informal classification guide based on the size of a solution in tokens, if it were output in a single response:

| Tier | Size | Output Token Count | Description | Use Cases |
| --- | --- | --- | --- | --- |
| 1 | Tiny | <= 100 | Snippet / Patch | Bug fixes, regex expressions, new CSS rules, import blocks |
| 2 | Small | 101 - 500 | Script / Component | DB migrations, API route handlers, basic unit tests, simple helper functions |
| 3 | Medium | 501 - 2000 | Feature | Project skeletons, test suite, error handling, logging |
| 4 | Large | 2001 - 4000 | Module | Complete MVC appliation layer, new microservices, integration tests, libraries |
| 5 | Huge | 4001 - 8000 | Application | Large scale refactors, complete CLI tools, standalone applications |
| 6 | Massive | > 8000 | Project | Complete web applications, games, SDKs |

Any modern model should be able to handle tier 1 problems easily. Models specialized for coding should also easily handle tier 2-3 problems. Frontier models should be able to solve tier 4 and possibly even tier 5 problems. Tier 6 problems typically require chain reasoning and multi-agent orchestration.
