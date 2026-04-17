# using-8liang-kit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a Reference Skill for AI assistants to use the `github.com/8liang/kit` package properly.

**Architecture:** A standalone skill directory `skills/using-8liang-kit/` containing a `SKILL.md` file following the `agentskills.io` specification with YAML frontmatter.

**Tech Stack:** Markdown, YAML.

---

### Task 1: Create the Skill Directory and File

**Files:**
- Create: `skills/using-8liang-kit/SKILL.md`

- [ ] **Step 1: Create the skill directory**

```bash
mkdir -p skills/using-8liang-kit
```

- [ ] **Step 2: Write the `SKILL.md` file**

```markdown
---
name: using-8liang-kit
description: Use when writing Go code that needs common utilities like event handling, excel parsing, weighted random selection, redis online status, time passing checks, or config parsing.
---

# using-8liang-kit

## Overview
This skill teaches AI assistants to leverage the `github.com/8liang/kit` package instead of reinventing common Go utilities. The `kit` package provides production-ready implementations for common backend and game server requirements.

## When to Use
- You need to trigger events across minute/hour/day/week boundaries (e.g., daily resets).
- You need a generic pub/sub event system for internal modules.
- You need to read/write Excel files or export them to JSON/Structs.
- You need to listen on TCP or Unix Domain Sockets.
- You need a weighted random selector (e.g., for gacha/loot tables).
- You need to manage user online status and heartbeats via Redis.
- You need to compile `.proto` files and inject Go tags.
- You need seed-based random generation.
- You need to parse Viper configs from various sources (local, HTTP, etcd).
- You need to get the local non-loopback IP or check if running in Docker/K8s.

## Quick Reference

| Module | Core Purpose | Key Functions / Types |
|--------|--------------|-----------------------|
| `timepass` | Time progression checks | `timepass.Advance(t)` |
| `event` | Generic events | `event.NewEvent(type, payload)` |
| `excel` | Excel operations | `excel.ExportExcelDataToJson` |
| `listener` | Network listening | `listener.NewListener(net, addr)` |
| `weighted` | Weighted random selection | `weighted.New[T]()`, `Add()`, `Draw()` |
| `onlinestore` | Redis online status | `onlinestore.New()`, `Heartbeat()` |
| `protobuf` | Protobuf compilation | `protobuf.Compile()` |
| `random` | Seeded random | `random.New()` |
| `viperparser`| Config loading | `viperparser.Unmarshal()` |
| `kit` | System utilities | `kit.IP()`, `kit.IsInDocker()` |

## Implementation Examples

### 1. Weighted Random Selection
Use `weighted` instead of writing custom random logic:
```go
import "github.com/8liang/kit/weighted"

// Initialize
selector := weighted.New[string]()

// Add items with weights
selector.Add("SSR_Item", 1)
selector.Add("SR_Item", 9)
selector.Add("R_Item", 90)

// Draw an item
result, err := selector.Draw()
if err != nil {
    // Handle error (e.g., empty selector)
}
```

### 2. Time Pass Checks (e.g., Daily Reset)
Use `timepass` to check if a boundary has been crossed:
```go
import "github.com/8liang/kit/timepass"

// Implement the timepass.Handler interface
// Advance(t time.Time) will trigger your handler's OnMinutePass, OnHourPass, OnDayPass, etc.
timepass.Advance(time.Now())
```

### 3. Online Store (Redis)
Use `onlinestore` for managing user presence:
```go
import "github.com/8liang/kit/onlinestore"

store := onlinestore.New(redisClient)
// Record heartbeat
store.Heartbeat(ctx, "user_123")

// Get online users
users, err := store.GetOnlineUsers(ctx)
```

## Red Flags - STOP and Rethink
- **Writing a custom weighted randomizer:** Delete it. Use `github.com/8liang/kit/weighted`.
- **Writing custom cron/daily reset logic:** Delete it. Use `github.com/8liang/kit/timepass`.
- **Writing custom Redis ZSet online tracking:** Delete it. Use `github.com/8liang/kit/onlinestore`.
- **Reinventing event buses for simple passing:** Use `github.com/8liang/kit/event`.
```

- [ ] **Step 3: Verify the file exists and is formatted correctly**

```bash
cat skills/using-8liang-kit/SKILL.md | head -n 10
```
Expected output: should start with `---` and the `name` and `description` YAML fields.
