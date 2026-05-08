---
name: writing-skills
description: Use when creating new skills, editing existing skills, or verifying skills work before deployment
---

# Writing Skills

## Overview

**Writing skills IS Test-Driven Development applied to process documentation.**

**Personal skills live in agent-specific directories (`~/.claude/skills` for Claude Code, `~/.agents/skills/` for Codex)**

You write test cases (pressure scenarios with subagents), watch them fail (baseline behavior), write the skill (documentation), watch tests pass (agents comply), and refactor (close loopholes).

**Core principle:** If you didn't watch an agent fail without the skill, you don't know if the skill teaches the right thing.

**REQUIRED BACKGROUND:** You MUST understand superpowers:test-driven-development before using this skill. That skill defines the fundamental RED-GREEN-REFACTOR cycle. This skill adapts TDD to documentation.

## What is a Skill?

A **skill** is a reference guide for proven techniques, patterns, or tools. Skills help future Claude instances find and apply effective approaches.

**Skills are:** Reusable techniques, patterns, tools, reference guides

**Skills are NOT:** Narratives about how you solved a problem once

## TDD Mapping for Skills

| TDD Concept | Skill Creation |
|-------------|----------------|
| **Test case** | Pressure scenario with subagent |
| **Production code** | Skill document (SKILL.md) |
| **Test fails (RED)** | Agent violates rule without skill (baseline) |
| **Test passes (GREEN)** | Agent complies with skill present |
| **Refactor** | Close loopholes while maintaining compliance |
| **Write test first** | Run baseline scenario BEFORE writing skill |
| **Watch it fail** | Document exact rationalizations agent uses |
| **Minimal code** | Write skill addressing those specific violations |
| **Watch it pass** | Verify agent now complies |
| **Refactor cycle** | Find new rationalizations → plug → re-verify |

The entire skill creation process follows RED-GREEN-REFACTOR.

## When to Create a Skill

**Create when:**
- Technique wasn't intuitively obvious to you
- You'd reference this again across projects
- Pattern applies broadly (not project-specific)
- Others would benefit

**Don't create for:**
- One-off solutions
- Standard practices well-documented elsewhere
- Project-specific conventions (put in CLAUDE.md)
- Mechanical constraints (if it's enforceable with regex/validation, automate it—save documentation for judgment calls)

## Skill Types

### Technique
Concrete method with steps to follow

### Pattern
Way of thinking about problems

### Reference
API docs, syntax guides, tool documentation

## Directory Structure

```
skills/
  skill-name/
    SKILL.md              # Main reference (required)
    supporting-file.*     # Only if needed
```

**Flat namespace** - all skills in one searchable namespace

## SKILL.md Structure

**Frontmatter (YAML):**
- Two required fields: `name` and `description`
- `name`: Use letters, numbers, and hyphens only (no parentheses, special chars)
- `description`: Start with "Use when..." to focus on triggering conditions
- **NEVER summarize the skill's process or workflow**

```markdown
---
name: Skill-Name-With-Hyphens
description: Use when [specific triggering conditions and symptoms]
---

# Skill Name

## Overview
What is this? Core principle in 1-2 sentences.

## When to Use
Bullet list with SYMPTOMS and use cases

## Core Pattern
Before/after code comparison

## Quick Reference
Table or bullets for scanning common operations

## Common Mistakes
What goes wrong + fixes
```

## Claude Search Optimization (CSO)

**Critical for discovery:** Future Claude needs to FIND your skill

### Key Rules

- **Description = When to Use, NOT What the Skill Does**
- Start with "Use when..." to focus on triggering conditions
- Include specific symptoms, situations, and contexts
- Write in third person
- **NEVER summarize the skill's process or workflow**

```yaml
# ❌ BAD: Summarizes workflow
description: Use when executing plans - dispatches subagent per task with code review

# ✅ GOOD: Just triggering conditions
description: Use when executing implementation plans with independent tasks in the current session

# ✅ GOOD: Triggering conditions only
description: Use when implementing any feature or bugfix, before writing implementation code
```

### Keyword Coverage

Use words Claude would search for:
- Error messages: "timeout", "race condition", "flaky"
- Symptoms: "hanging", "zombie", "pollution"
- Tools: Actual commands, library names, file types

### Token Efficiency

- getting-started workflows: <150 words each
- Frequently-loaded skills: <200 words total
- Other skills: <500 words

### Cross-Referencing

Use explicit markers:
- ✅ `**REQUIRED SUB-SKILL:** Use superpowers:test-driven-development`
- ❌ `@skills/testing/test-driven-development` (force-loads, burns context)

## The Iron Law

```
NO SKILL WITHOUT A FAILING TEST FIRST
```

This applies to NEW skills AND EDITS to existing skills.

## Testing All Skill Types

### Discipline-Enforcing Skills (rules/requirements)
Test with: Academic questions, pressure scenarios, rationalizations

### Technique Skills (how-to guides)
Test with: Application scenarios, variation scenarios, edge cases

### Pattern Skills (mental models)
Test with: Recognition scenarios, application scenarios

### Reference Skills (documentation/APIs)
Test with: Retrieval scenarios, application scenarios

## RED-GREEN-REFACTOR for Skills

### RED: Write Failing Test (Baseline)
Run pressure scenario WITHOUT the skill. Document exact behavior and rationalizations.

### GREEN: Write Minimal Skill
Write skill addressing specific rationalizations from baseline.

### REFACTOR: Close Loopholes
Add explicit counters for new rationalizations. Re-test until bulletproof.

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "Skill is obviously clear" | Clear to you ≠ clear to other agents |
| "Testing is overkill" | Untested skills have issues. Always. |
| "I'll test if problems emerge" | Problems = agents can't use skill |

## Anti-Patterns

- ❌ Narrative Example: "In session 2025-10-03, we found..."
- ❌ Multi-Language Dilution: example in 5 languages
- ❌ Generic Labels: helper1, step2, pattern4

## The Bottom Line

**Creating skills IS TDD for process documentation.**

Same Iron Law: No skill without failing test first.
Same cycle: RED → GREEN → REFACTOR.
Same benefits: Better quality, fewer surprises, bulletproof results.