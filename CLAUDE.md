<!-- SPECKIT START -->
**Current Implementation Plan**: `specs/001-mt5-mcp-integration/plan.md`

For architecture, tech stack, design decisions, and project structure, read the plan. Key artifacts:
- `spec.md` — Requirements + user stories (clarifications complete)
- `research.md` — MT5 WebAPI, libraries, gRPC daemon patterns
- `data-model.md` — Entity definitions + relationships
- `contracts/mcp-tools.md` — gRPC + MCP tool contracts
- `quickstart.md` — Setup guide
- `plan.md` — This document (Phase 0-2 planning)

<!-- SPECKIT END -->

## graphify

This project has a graphify knowledge graph at graphify-out/.

Rules:
- Before answering architecture or codebase questions, read graphify-out/GRAPH_REPORT.md for god nodes and community structure
- If graphify-out/wiki/index.md exists, navigate it instead of reading raw files
- For cross-module "how does X relate to Y" questions, prefer `graphify query "<question>"`, `graphify path "<A>" "<B>"`, or `graphify explain "<concept>"` over grep — these traverse the graph's EXTRACTED + INFERRED edges instead of scanning files
- After modifying code files in this session, run `graphify update .` to keep the graph current (AST-only, no API cost)
