# using-8liang-kit Skill Design

## Overview
This skill teaches AI assistants how to leverage the `github.com/8liang/kit` package instead of reinventing common Go utilities. It's structured as a reference skill following the TDD-documentation approach and `agentskills.io` specification so it can be installed via `npx skill`.

## Frontmatter (YAML)
```yaml
---
name: using-8liang-kit
description: Use when writing Go code that needs common utilities like event handling, excel parsing, weighted random selection, redis online status, time passing checks, or config parsing.
---
```

## Structure
1. **Overview**: Brief explanation of the `8liang/kit` package.
2. **When to Use**: Concrete triggers (e.g., "Need to check if a day has passed since last action", "Need to pick an item based on weights").
3. **Quick Reference Table**: Maps problems to `kit` modules.
4. **Implementation Examples**:
   - `timepass`: Code snippet showing `timepass.Advance(t)`.
   - `weighted`: Code snippet showing `Selector.Add` and `Selector.Draw`.
   - `onlinestore`: Snippet for `Heartbeat` and `GetOnlineUsers`.
5. **Red Flags**: Things the AI should NOT do (e.g., write a custom weighted randomizer when `weighted` exists).

## Deployment for `npx skill`
To make it installable via `npx skill`:
- The skill will be created as `SKILL.md` in a standalone directory (e.g., `skills/using-8liang-kit/SKILL.md`).
- This allows it to be published to a Git repository or npm, fulfilling the user's requirement.
