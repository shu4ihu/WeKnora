# Go Code Documentation Templates / Workflows for WeKnora + Notion

## Research scope

This note researches how a Pi skill should guide an agent to read WeKnora Go code and produce structured Markdown architecture documentation suitable for Notion publishing. It focuses on two MVP modes:

1. **Package mode**: document one Go directory/package.
2. **Module mode**: document a cross-package business/domain slice.

Sources inspected:

- Task PRD: `.trellis/tasks/05-21-go-code-doc-notion-skill/prd.md`
- Existing Notion pages: `WeKnora 后端模块总结 - 目录`, `01 启动入口与服务生命周期`, `04 认证、RBAC 与 KB Access 权限体系`, `09 RAG 检索与问答 Pipeline`
- Repo orientation: `docs/项目梳理.md`
- Existing module docs: `internal/event/SUMMARY.md`, `internal/datasource/README.md`
- Trellis guides: `.trellis/spec/guides/cross-layer-thinking-guide.md`, `.trellis/spec/guides/code-reuse-thinking-guide.md`
- Project structure observations from `cmd/`, `internal/`, `client/`, `cli/`, and `go.mod`

## Comparable documentation patterns

### 1. Existing Notion backend module summary pattern

Current Notion module pages use a compact architecture-summary style. Common sections:

- `## 模块定位`
- `## 具体业务作用`
- `## 核心代码引用`
- `## 关键逻辑摘要` or `## 关键流程摘要`
- `## 设计价值`
- Optional: `## 和其他模块的关系`, `## 注意点`, feature-specific implications

Strong conventions already present:

- Chinese-first prose.
- Preserve Go symbols, packages, API names, config keys, and file paths in English/code formatting.
- Use concrete source references in format `path/to/file.go:start-end`.
- Emphasize “business role” and not only API or symbol inventory.
- Use simple flow diagrams in code fences for lifecycle/pipeline docs.
- One Notion page per relatively independent backend module.

This is the best fit for **module mode**, because existing pages already establish reader expectations and Notion page naming.

### 2. Existing repo Markdown module docs pattern

`internal/datasource/README.md` uses a longer module README style:

- Overview
- Architecture diagram
- Core components
- Task flow
- File structure
- Data models
- Interface contracts
- API endpoints
- Implementation guide
- Database schema

`internal/event/SUMMARY.md` uses a feature-summary style:

- Overview
- Core functionality
- File structure
- Performance / operational metrics
- Usage scenarios
- Integration steps
- Advantages
- Tests
- Follow-up suggestions

Useful takeaways:

- Include file tree when a package/subsystem spans multiple files.
- Show interfaces or structs when they define extension points.
- Include API endpoints only when the feature has handler/router exposure.
- Include integration/extension steps only when the documented area is a framework or plugin point.

### 3. Go package documentation conventions

Standard Go docs usually focus on package purpose, exported API, examples, and contracts. For this skill, pure GoDoc-style output is not enough; WeKnora needs architecture context. Still, package mode should borrow Go conventions:

- Identify import path/module context.
- State package responsibility in one paragraph.
- List exported types/functions and their intended use.
- Separate public API from internal helpers.
- Mention tests and examples.
- Mention package-level side effects, init behavior, goroutines, external IO, and concurrency assumptions when present.

### 4. Cross-layer design documentation pattern

Trellis cross-layer guide emphasizes data-flow and boundary contracts:

```text
Source → Transform → Store → Retrieve → Transform → Display
```

For WeKnora business modules this maps well to:

```text
HTTP/CLI/Task entry
  → Handler / DTO
  → Service orchestration
  → Repository / Connector / Model provider
  → Database / Vector store / External service
  → Response / Stream / Async status
```

Module-mode docs should explicitly list boundary contracts and validation points because many WeKnora features span handlers, services, repositories, middleware, tasks, and external systems.

## WeKnora-specific constraints and code-reading implications

### Multi-module repository

The repo contains at least these Go modules:

- Root module: `github.com/Tencent/WeKnora`
- `client/`
- `cli/`

The skill must not assume every package belongs to the root module. Package mode should detect module root by walking up to nearest `go.mod` and derive package/import path accordingly.

### Backend layer layout

Observed backend code layout:

- `cmd/server/`: server process entry and lifecycle.
- `cmd/desktop/`: desktop/Wails entry.
- `internal/container/`: dependency injection and application assembly.
- `internal/router/`: Gin routes, RBAC guards, async task registration.
- `internal/middleware/`: auth, RBAC, KB access, logging, request context.
- `internal/handler/`: HTTP handlers and request/response coordination.
- `internal/handler/dto/`: request/response DTO transformations.
- `internal/application/service/`: business orchestration.
- `internal/application/repository/`: persistence and query implementations.
- `internal/types/`: domain data models and shared types.
- `internal/types/interfaces/`: service/repository contracts.
- `internal/datasource/`: connector framework and external sync adapters.
- `internal/models/`: model providers and model capability abstractions.
- `internal/infrastructure/`: chunkers, doc parsers, web fetch/search, etc.
- `internal/agent/`: agent runtime, tools, skills, memory, token handling.

Module-mode documentation should use these layers as an inspection map.

### Existing Notion title conventions

PRD requires:

- Module mode: `WeKnora 后端模块总结 - <模块名>`
- Package mode: `WeKnora Go Package - <package path>`

The skill should produce the title before the body and use it for Notion search/upsert behavior.

### Existing Notion sync behavior

Available Notion tools support search, read, create, and append. They do not expose a direct page replacement API. Therefore MVP sync workflow should be append/create:

1. Determine exact page title.
2. Search Notion for that title.
3. If exact title match exists, append a timestamped update section.
4. If no exact match exists, create a new page in configured database.
5. Never print or persist Notion token or raw secrets.

### Code citation convention

Existing Notion pages use source references like:

```text
internal/application/service/session_knowledge_qa.go:22-64
```

The skill should instruct agents to cite specific ranges for key files. If exact line numbers are unavailable or expensive, agents should still cite file paths and symbol names, but line ranges should be preferred for Notion-quality docs.

### Language/style convention

Use a Chinese-first, bilingual technical style:

- Section headings and explanations in Chinese.
- Preserve Go packages, filenames, types, methods, routes, environment variables, database tables, and API endpoint names in English/code style.
- Avoid translating established technical identifiers.
- Avoid dumping huge code blocks; include short snippets only when they clarify a contract or control flow.

## Recommended safe inspection workflow

The skill should guide the agent through this workflow before asking the user for more code context:

### Step 1: Determine scope and mode

If user names a directory/package, use **package mode**.

If user names a capability/domain such as “RAG 检索”, “数据源同步”, “权限”, “Agent tools”, use **module mode** and search across layers.

If ambiguous, inspect repository first, then ask one focused clarification question.

### Step 2: Locate candidate files

Package mode:

- List files in target directory.
- Identify `*_test.go`, generated files, docs, and package comments.
- Locate nearest `go.mod`.
- Search for package imports/usages from other packages.

Module mode:

- Search by domain keywords in `internal/types`, `internal/types/interfaces`, `internal/handler`, `internal/router`, `internal/application/service`, `internal/application/repository`, `internal/middleware`, `internal/container`, and relevant `cmd/` entries.
- Include docs such as `docs/api/*.md` or existing module README when relevant.
- Identify tests that encode behavior.

### Step 3: Build a code map

For each file, record:

- Layer and responsibility.
- Important exported types/functions/methods.
- Important constants/config/env/table names.
- Caller/callee relationships.
- External dependencies: DB, Redis/Asynq, MinIO/COS/S3, vector stores, LLM providers, DocReader gRPC, Notion/Feishu/Yuque, etc.

### Step 4: Trace flows and contracts

For module mode, trace at least one happy-path flow and important edge/error paths:

- Entry point: route/handler/task/CLI/tool.
- Request/DTO shape.
- Auth/RBAC/KB access guards if applicable.
- Service orchestration.
- Repository/external provider calls.
- Persistence/state changes.
- Response/stream/task result.
- Observability: logs, tracing spans, events, metrics.

### Step 5: Draft Markdown

Use the templates below. Mark unknowns explicitly instead of inventing details.

### Step 6: Notion preparation

Before sync:

- Verify title convention.
- Remove secrets and local-only noise.
- Ensure Markdown is Notion-friendly: headings, bullets, tables, short code fences.
- If appending, wrap in a timestamped update heading.

## Recommended package-mode template

Use title convention: `WeKnora Go Package - <package path>`

```markdown
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
```

Package mode should be concise but concrete. It should not become a full business-module document unless the package itself is the module boundary.

## Recommended module-mode template

Use title convention: `WeKnora 后端模块总结 - <模块名>`

```markdown
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
```

Module mode should be close to the current Notion module-summary style but add `涉及代码范围` and `数据与边界契约` for better maintainability.

## Recommended Notion append wrapper

When appending to an existing page, wrap the generated body with a timestamped heading to avoid overwriting old content:

```markdown
---

## 更新：<YYYY-MM-DD HH:mm> - <scope/title>

<generated documentation body, without duplicate top-level title if desired>
```

For a new page, include the full title as `# ...` at the top.

## MVP boundaries and future enhancements

### Include in MVP

- Safe code inspection workflow.
- Package mode and module mode templates.
- Exact Notion title conventions.
- Search-and-append/create sync workflow.
- Source citations with file paths and preferred line ranges.
- Chinese-first bilingual style.
- No secrets in output.

### Leave out of MVP

- Perfect AST/symbol extraction.
- Automatic background syncing on file changes.
- Full Notion page replacement/upsert unless a custom tool is added.
- Diagrams beyond simple text flow diagrams.
- Exhaustive public API reference for every symbol.

## Practical skill instructions to encode

A future `SKILL.md` should instruct the agent to:

1. Inspect code first; do not ask the user to explain code that is available in the repo.
2. Choose package mode vs module mode based on user wording and discovered scope.
3. Use repo search/read tools to collect file map, types, routes, services, repositories, tests, and docs.
4. Produce Markdown using the correct template and title convention.
5. For Notion sync, use configured Notion tools: search exact title, append timestamped update if found, otherwise create page.
6. Confirm before publishing if the user did not explicitly request sync.
7. Never expose tokens or secrets.
8. Report generated title, files inspected, and Notion page action succinctly.
