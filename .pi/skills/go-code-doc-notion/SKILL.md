---
name: go-code-doc-notion
description: "Read and summarize WeKnora Go code into normalized Chinese-first Markdown documentation, with package-mode and cross-package backend module-mode templates, and optionally sync the generated docs to Notion using the configured notion tools. Use for Go package docs, backend/module architecture summaries, code-reading reports, Markdown docs, or Notion publishing of WeKnora Go documentation."
---

# Go Code Documentation + Notion Sync

Use this skill to inspect WeKnora Go code, generate structured Markdown documentation, and optionally publish it to the configured Notion database.

## Core Rules

- Inspect repository code first. Do not ask the user to explain code that is available in the repo.
- Keep output Chinese-first, bilingual where helpful: write explanations in Chinese, but preserve Go package names, file paths, symbols, routes, config keys, API names, table names, and environment variables in English/code formatting.
- Support exactly two MVP scopes:
  - **Package mode**: one Go directory/package.
  - **Module mode**: one cross-package business/domain slice.
- Use stable page titles:
  - Module mode: `WeKnora 后端模块总结 - <模块名>`
  - Package mode: `WeKnora Go Package - <package path>`
- Generate Notion-friendly Markdown: headings, bullets, short tables, short `text` flow diagrams, and concise code snippets only when they clarify a contract.
- Mark uncertainty explicitly with `待确认` / `未知`, rather than inventing behavior.
- Do not expose secrets, tokens, raw environment variables, or local credential values.
- Do not sync to Notion unless the user explicitly asks to publish/sync, or confirms after being asked.
- Do not add scripts or custom extensions for this MVP; orchestrate existing tools only.

## Mode Selection

1. If the user gives a Go directory/package path such as `internal/errors`, `internal/handler`, `client`, or `cli`, use **package mode**.
2. If the user gives a business/domain capability such as `RAG 检索`, `权限`, `数据源同步`, `模型接入层`, `Agent tools`, or `文档解析与 Chunker 分块`, use **module mode**.
3. If ambiguous, inspect likely paths and search keywords first, then ask one concise clarification question.

## Repository Inspection Workflow

Always build evidence from source files before drafting the document.

### Common orientation

Use repository tools such as:

```bash
find . -name go.mod -print
find cmd internal client cli -maxdepth 3 -type f \( -name '*.go' -o -name '*.md' \) 2>/dev/null
rg -n "<keyword>|type <Name>|func .*<Name>|Route|Group|POST|GET|PUT|DELETE" cmd internal client cli docs 2>/dev/null
```

WeKnora has multiple Go modules and code areas. Check the nearest `go.mod` for the target:

- Root backend module: repository root, with `cmd/`, `internal/`, `docreader/`, `docs/`.
- CLI module: `cli/`.
- Client module: `client/`.

Important backend layers for module mode:

- `cmd/server/`, `cmd/desktop/` — process entry/lifecycle.
- `internal/container/` — dependency injection and application assembly.
- `internal/router/` — Gin route registration, middleware, task routes.
- `internal/middleware/` — auth, RBAC, KB access, logging, request context.
- `internal/handler/`, `internal/handler/dto/` — HTTP handlers and DTO transforms.
- `internal/application/service/` — business orchestration.
- `internal/application/repository/` — persistence and query implementations.
- `internal/types/`, `internal/types/interfaces/` — domain models and contracts.
- `internal/datasource/` — external sync connector framework.
- `internal/models/` — model providers and capability abstractions.
- `internal/infrastructure/` — chunkers, parsers, web fetch/search, external infrastructure.
- `internal/agent/`, `internal/event/`, `internal/errors/`, `internal/logger/`, `internal/config/` — cross-cutting subsystems.

### Package mode inspection

For a package directory:

1. List Go and Markdown files in the directory; include `*_test.go` files.
2. Locate the nearest parent `go.mod` and derive the module/import context.
3. Read core `.go` files directly, prioritizing package docs, exported types/functions, constructors, interfaces, route handlers, services, repositories, and tests.
4. Search upstream usages and imports:

   ```bash
   rg -n "<package-name>|<ExportedType>|<ExportedFunc>" cmd internal client cli docs 2>/dev/null
   ```

5. Record package name, responsibilities, exported API, important private helpers, dependencies, side effects, goroutines, external IO, errors, concurrency/idempotency assumptions, and tests.

### Module mode inspection

For a business/domain module:

1. Search domain keywords and likely English aliases across layers:

   ```bash
   rg -n "<keyword>|<EnglishAlias>|<TypeName>|<route-fragment>" internal cmd client cli docs 2>/dev/null
   ```

2. Trace at least one happy-path flow and important edge/error paths:

   ```text
   HTTP / CLI / Task entry
     → Router / Middleware
     → Handler / DTO
     → Service orchestration
     → Repository / Connector / Model provider / Infrastructure
     → Database / Vector store / Redis / Object storage / External service
     → Response / Stream / Async status
   ```

3. Read representative files, not only grep snippets.
4. Include docs such as `docs/api/*.md`, package `README.md`, or existing module summaries when relevant.
5. Record layer responsibilities, boundary contracts, validation, permissions/tenant requirements, storage/external integrations, observability, and tests.

## Markdown Normalization Rules

Apply these rules with medium strictness:

- Start with exactly one `# <title>` matching the selected title convention.
- Include metadata block immediately after the title:
  - `生成时间：<YYYY-MM-DD HH:mm>`
  - `范围：...`
  - Package mode also includes `Go module：...`.
- Keep the section order from the selected template unless a section is clearly not applicable.
- Prefer source citations in this format: ``path/to/file.go:start-end``. If line ranges are not available, use ``path/to/file.go`` plus symbol names.
- Use Markdown tables for file maps and boundary contracts.
- Use fenced `text` diagrams for flows; avoid Mermaid unless the user asks.
- Avoid long code dumps. Include only short snippets or signatures when necessary.
- Use `## 待确认问题` / `## 后续可补充方向` for unknowns and follow-ups.
- Before Notion sync, remove secrets and local-only noise.

## Package Mode Template

Use this template for a single Go directory/package. Omit sections only when clearly irrelevant.

````markdown
# WeKnora Go Package - <package path>

> 生成时间：<YYYY-MM-DD HH:mm>
> 范围：`<directory>`
> Go module：`<module path>`

## 包定位

用 2-4 句话说明该 package 在 WeKnora 中负责什么，属于哪一层，典型调用者是谁。

## 目录与文件概览

| 文件 | 作用 | 备注 |
|---|---|---|
| `path/file.go` | ... | 关键类型/函数：`Type`, `Func` |

## 对外接口与核心类型

### 类型 / 接口

- `TypeName`：职责、关键字段、使用场景。
- `InterfaceName`：调用方、实现方、契约。

### 函数 / 方法

- `FuncName(...)`：输入、输出、副作用、错误行为。

## 关键流程

```text
调用方
  → package 函数/方法
  → 内部处理
  → 外部依赖/返回结果
```

## 依赖关系

### 上游调用者

- `path/to/caller.go:line-line`：如何调用本包。

### 下游依赖

- `package/path`：用途。
- 数据库 / Redis / 外部服务 / 文件系统：用途和注意点。

## 错误处理与边界条件

- 输入为空 / 缺字段时如何处理。
- 权限、租户、用户上下文要求。
- 超时、取消、外部服务失败。
- 并发安全或幂等性要求。

## 测试与行为保证

| 测试文件 | 覆盖内容 |
|---|---|
| `path/file_test.go` | ... |

## 代码引用

- `path/file.go:start-end`：说明。
- `path/file_test.go:start-end`：说明。

## 使用 / 修改建议

- 新增功能应复用哪些类型或函数。
- 不应绕过哪些入口或校验。
- 可能需要同步更新的文件或文档。

## 待确认问题

- 无；或列出需要用户/维护者确认的信息。
````

## Module Mode Template

Use this template for a cross-package business/domain slice. Keep it close to existing WeKnora backend Notion module summaries.

````markdown
# WeKnora 后端模块总结 - <模块名>

> 生成时间：<YYYY-MM-DD HH:mm>
> 范围：<business/domain scope>

## 模块定位

说明该模块解决什么业务问题，在 WeKnora 产品链路中的位置，以及主要入口。

## 具体业务作用

- 作用 1。
- 作用 2。
- 作用 3。

## 涉及代码范围

| 层级 | 文件 / 包 | 职责 |
|---|---|---|
| Router / Middleware | `internal/router/...` | ... |
| Handler / DTO | `internal/handler/...` | ... |
| Service | `internal/application/service/...` | ... |
| Repository | `internal/application/repository/...` | ... |
| Types / Interfaces | `internal/types/...` | ... |
| Infra / External | `internal/...` | ... |

## 核心代码引用

### <子主题 A>

- `path/file.go:start-end`：说明该段代码的角色。
- `path/file.go:start-end`：说明该段代码的角色。

### <子主题 B>

- `path/file.go:start-end`：说明该段代码的角色。

## 关键流程摘要

```text
入口
  ↓
权限 / 参数校验
  ↓
Service 编排
  ↓
Repository / 外部依赖
  ↓
状态变更 / 响应 / 异步任务
```

## 数据与边界契约

| 边界 | 输入 | 输出 | 校验 / 错误 |
|---|---|---|---|
| API → Handler | ... | ... | ... |
| Handler → Service | ... | ... | ... |
| Service → Repository | ... | ... | ... |
| Service → External | ... | ... | ... |

## 设计价值

说明当前拆分为什么合理：复用性、低耦合、扩展点、可观测性、故障隔离、权限集中化等。

## 和其他模块的关系

- 与 `<模块>` 的关系：调用 / 被调用 / 共享模型 / 共享权限。
- 与数据库、异步任务、模型层、向量库、DocReader、IM、MCP、Agent 等的关系。

## 注意点

- 权限与租户边界。
- 幂等性、并发、事务、异步任务重试。
- 配置项、环境变量、外部服务依赖。
- 常见修改风险。

## 测试与验证建议

- 单元测试：...
- 集成测试：...
- 手动验证：...

## 后续可补充方向

- 更细粒度专题。
- 需要补充的图、API 文档或测试。

## 待确认问题

- 无；或列出需要用户/维护者确认的信息。
````

## Notion Sync Workflow

Only run this workflow when the user explicitly asks to sync/publish to Notion, or after the user confirms a proposed publish action.

1. Determine the exact title using the mode convention.
2. Run `notion_status` to verify Notion is configured. Do not print tokens or raw secret values.
3. Run `notion_search` with the exact title or the most distinctive title fragment.
4. Inspect search results:
   - If exactly one exact-title page exists, call `notion_append_to_page` to append to that page.
   - If multiple exact-title pages exist, ask the user which page to use.
   - If no exact title exists but non-exact candidates appear, ask the user before publishing.
   - If no suitable page exists, call `notion_create_page` to create a new page in the configured database.
5. For existing pages, append the generated body wrapped like this:

   ```markdown
   ---

   ## 更新：<YYYY-MM-DD HH:mm> - <scope/title>

   <generated documentation body, preferably without duplicating the top-level # title>
   ```

6. For new pages, call `notion_create_page` with the full generated Markdown including the `# <title>` heading.
7. After sync, report the page action succinctly: created, appended, or skipped/needs clarification.

## Manual Usage Examples

- `/skill:go-code-doc-notion document package internal/errors`
- `/skill:go-code-doc-notion summarize package client and do not sync`
- `/skill:go-code-doc-notion document module 文档解析与 Chunker 分块`
- `/skill:go-code-doc-notion generate Notion doc for RAG 检索与问答 Pipeline and sync after confirmation`

When responding to the user, include:

- selected mode and exact title;
- files/directories inspected;
- generated Markdown or a concise summary plus where the Markdown was saved if the user requested a file;
- Notion action if sync was requested.
