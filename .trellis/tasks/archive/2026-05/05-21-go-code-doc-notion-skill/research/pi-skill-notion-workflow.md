# Research: Pi skill + Notion workflow for Go code documentation

## Scope

Active task: `.trellis/tasks/05-21-go-code-doc-notion-skill`

Research target: design a Pi skill that inspects Go code, generates structured Markdown, and optionally publishes/upserts that documentation through the existing Notion tools.

## Sources consulted

- Pi docs: `C:\Users\lam_zhang\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent\docs\skills.md`
- Pi docs: `C:\Users\lam_zhang\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent\docs\extensions.md`
- Pi extension examples index: `C:\Users\lam_zhang\AppData\Roaming\npm\node_modules\@earendil-works\pi-coding-agent\examples\extensions\README.md`
- Pi dynamic resource example: `examples/extensions/dynamic-resources/index.ts` and `examples/extensions/dynamic-resources/SKILL.md`
- Pi custom tool example: `examples/extensions/hello.ts`
- Project settings: `.pi/settings.json`
- Existing project skills: `.pi/skills/trellis-*/SKILL.md`
- Notion extension status/tool behavior via `notion_status` and `notion_query_database`
- Repo layout via `find` over `cmd/`, `internal/`, `client/`, `cli/`, and `go.mod` files

## Key Pi conventions

### Skill discovery and structure

- Project-local skills are discovered from `.pi/skills/`.
- A skill is normally a directory containing `SKILL.md`.
- Required frontmatter:
  - `name`: lowercase letters, numbers, hyphens, max 64 chars.
  - `description`: required, max 1024 chars; this controls when the agent loads the skill.
- Pi also supports skill commands when enabled. This repo has `enableSkillCommands: true`, so a skill named `go-code-doc-notion` would be invokable as `/skill:go-code-doc-notion`.
- Relative references inside a skill should be resolved against the skill directory.
- Skills are progressive-disclosure resources: the model sees name/description first, then reads `SKILL.md` on demand.

### Project settings mapping

Current `.pi/settings.json` already supports the desired shape:

```json
{
  "enableSkillCommands": true,
  "extensions": ["./extensions/trellis/index.ts"],
  "skills": ["./skills"],
  "prompts": ["./prompts"]
}
```

Because settings load `./skills` relative to `.pi/`, a new project skill under `.pi/skills/go-code-doc-notion/SKILL.md` should be discovered without settings changes.

### Extension conventions relevant to this task

- Project-local extensions can live under `.pi/extensions/*.ts` or `.pi/extensions/*/index.ts`.
- Extensions can register custom tools via `pi.registerTool()`, commands via `pi.registerCommand()`, and dynamic skills/prompts/themes through the `resources_discover` event.
- A minimal custom tool follows the `defineTool` / `pi.registerTool()` pattern from `examples/extensions/hello.ts`.
- Dynamic resources can inject a `SKILL.md` with `resources_discover`, but this is unnecessary if the skill is static under `.pi/skills/`.
- Extensions are more powerful but run TypeScript with full permissions and add runtime/maintenance burden.

## Existing Notion tool behavior

`notion_status` reports:

- `NOTION_TOKEN`: configured
- `NOTION_DATABASE_ID`: configured
- `NOTION_TITLE_PROPERTY`: `文档名`

`notion_query_database` returned existing backend module summary pages, including:

- `WeKnora 后端模块总结 - 目录`
- `WeKnora 后端模块总结 - 13 错误处理、日志与可观测性`
- `WeKnora 后端模块总结 - 12 外部数据源同步`
- `WeKnora 后端模块总结 - 11 文档解析与 Chunker 分块`
- `WeKnora 后端模块总结 - 10 模型接入层 Chat Embedding Rerank Provider`

Available tools in this session include:

- `notion_status`
- `notion_search`
- `notion_read_page`
- `notion_query_database`
- `notion_create_page`
- `notion_append_to_page`

Observed implication: the MVP can implement an append-or-create publishing policy. It should not promise full content replacement unless a future extension/tool adds replacement support.

## WeKnora repo mapping

Repo Go layout includes multiple Go modules and code areas:

- Root module: `go.mod` at repo root, with backend code in `cmd/`, `internal/`, `docreader/`, `docs/`, etc.
- CLI module: `cli/go.mod`, code under `cli/`.
- Client module: `client/go.mod`, code under `client/`.
- Backend entry points: `cmd/server/`, `cmd/desktop/`.
- Backend layers/domains observed under `internal/`:
  - `internal/router`
  - `internal/handler`
  - `internal/handler/dto`
  - `internal/application/service`
  - `internal/application/repository`
  - `internal/application/repository/retriever`
  - `internal/infrastructure/chunker`
  - `internal/infrastructure/docparser`
  - `internal/infrastructure/web_fetch`
  - `internal/infrastructure/web_search`
  - `internal/models/{chat,embedding,rerank,provider,vlm,asr}`
  - `internal/datasource/connector/{notion,feishu,yuque}`
  - `internal/types` and `internal/types/interfaces`
  - `internal/event`, `internal/errors`, `internal/logger`, `internal/config`, `internal/container`

This supports two useful documentation scopes:

1. **Package mode**: inspect one directory/package, e.g. `internal/handler`, `internal/application/service/retriever`, `client`.
2. **Module mode**: inspect a cross-layer business slice, e.g. “知识库”, “模型接入层”, “文档解析与 Chunker 分块”, spanning router/handler/service/repository/types/infrastructure.

## Comparable implementation patterns

### Pattern 1: Pure static Pi skill under `.pi/skills/`

**Shape**

```text
.pi/skills/go-code-doc-notion/
└── SKILL.md
```

`SKILL.md` contains:

- Trigger description for Go code documentation and Notion publishing.
- Workflow steps for package mode and module mode.
- Repository inspection commands to run before asking the user for context.
- Markdown documentation template.
- Notion sync procedure: search exact/same title; append timestamped update if found; create page if not found.
- Safety constraints: do not expose tokens; do not publish without explicit user request/confirmation.

**Pros**

- Lowest complexity; no runtime code.
- Fits Pi skills model exactly.
- Uses existing tools (`read`, `bash`, `notion_*`) and current Notion configuration.
- Easy to review and safe to commit.
- Works with current `.pi/settings.json` without changes.

**Cons**

- Relies on agent following instructions consistently.
- No deterministic helper for exact-title Notion upsert beyond documented search/append/create behavior.
- No automated Go AST parsing; code extraction is tool/agent-driven.

**Fit for MVP**

Best fit. The task is primarily an agent workflow and documentation template, not a product/runtime integration.

### Pattern 2: Static skill plus helper scripts/templates

**Shape**

```text
.pi/skills/go-code-doc-notion/
├── SKILL.md
├── templates/
│   ├── package.md
│   └── module.md
└── scripts/
    └── inspect-go-package.sh  # optional
```

Potential helper scripts could list Go files, packages, public symbols, imports, or route registrations. Templates keep the Markdown format consistent and reduce `SKILL.md` length.

**Pros**

- More repeatable package inspection.
- Keeps `SKILL.md` readable while providing detailed templates/references.
- Still uses normal skill discovery and existing Notion tools.
- Good path if the template becomes long.

**Cons**

- Scripts add cross-platform concerns. This environment uses bash, but project users may run Pi on Windows shells.
- Scripts must be maintained and tested.
- If scripts inspect symbols, the MVP risks drifting toward AST tooling and away from the simple workflow.

**Fit for MVP**

Acceptable if limited to Markdown reference templates. Avoid executable scripts in MVP unless a clear need appears. A single `SKILL.md` can embed the template initially.

### Pattern 3: Skill plus custom Pi extension/tool for Notion upsert

**Shape**

```text
.pi/skills/go-code-doc-notion/SKILL.md
.pi/extensions/go-doc-notion/index.ts
```

The extension might register a custom tool such as `notion_upsert_markdown_doc` that accepts `{ title, markdown, mode }`, searches the database, and append-or-create in one deterministic operation.

**Pros**

- Encapsulates Notion upsert policy.
- Reduces repeated LLM orchestration for search/append/create.
- Can add consistent timestamp formatting and page-title normalization.

**Cons**

- Requires TypeScript extension code and validation.
- Needs access to Notion client APIs or must wrap existing extension behavior, which may not be directly exposed across extensions.
- Higher permission/security footprint.
- Duplicates or bypasses already-configured Notion tools unless carefully integrated.
- Out of scope if current tools are sufficient.

**Fit for MVP**

Not recommended for MVP. Revisit only if pure skill upsert is unreliable, page replacement becomes required, or multiple docs must be published in batch.

### Pattern 4: Full custom extension with command-driven documentation wizard

**Shape**

```text
.pi/extensions/go-doc-wizard/index.ts
```

The extension registers `/go-doc` and possibly tools/UI prompts to ask mode, package path/module name, Notion sync choice, and then inject a prompt or call tools.

**Pros**

- Polished interactive UX.
- Can standardize mode selection and confirmation.
- Can dynamically load a bundled skill via `resources_discover` if packaged.

**Cons**

- Over-engineered for current need.
- Extensions cannot themselves perform the semantic code review unless they call the model indirectly via injected messages; the agent still needs the skill-like instructions.
- More code and more failure modes.

**Fit for MVP**

Not recommended. Useful later if this becomes a shared productized Pi package.

## Common workflow conventions to encode in the skill

### Initial mode detection

The skill should infer or ask for one of two modes:

- **Package mode**: user gives a Go directory/package path.
  - Title: `WeKnora Go Package - <package path>`
  - Example: `WeKnora Go Package - internal/handler`
- **Module mode**: user gives a business/domain module name.
  - Title: `WeKnora 后端模块总结 - <模块名>`
  - Example: `WeKnora 后端模块总结 - 文档解析与 Chunker 分块`

If the user request is ambiguous, inspect likely paths first, then ask one concise clarification.

### Safe inspection before questions

The skill should instruct the agent to inspect repository facts before asking the user for code context. Suggested commands:

```bash
find . -name go.mod -print
find <target-dir> -maxdepth 2 -name '*.go' -print
rg -n "<module keyword>|type <Name>|func .*<Name>|Route|Group|POST|GET" internal cmd client cli
```

For package mode:

- List files in the directory.
- Read core `.go` files directly.
- Identify package name, exported types/functions, imports, tests, and neighboring packages.

For module mode:

- Search by module keywords in `internal/router`, `internal/handler`, `internal/application`, `internal/types`, `internal/infrastructure`, `internal/models`, `client`, and `cli` as relevant.
- Trace request/data flow across router → handler → service → repository/infrastructure/model/provider.
- Read representative files rather than relying only on grep snippets.

### Markdown template conventions

Recommended Chinese-first bilingual technical style:

```markdown
# <Title>

## 1. 模块定位 / Purpose

## 2. 代码范围 / Code Scope
- `<path>` — description

## 3. 核心职责 / Responsibilities

## 4. 关键文件与类型 / Key Files & Types

## 5. 对外接口 / Public APIs
- HTTP routes / handlers
- Go exported functions/types
- Client/CLI entry points if applicable

## 6. 数据流与调用链 / Data Flow

## 7. 依赖关系 / Dependencies

## 8. 配置、存储与外部系统 / Config, Storage, Integrations

## 9. 错误处理与可观测性 / Errors & Observability

## 10. 测试与验证 / Tests

## 11. 运维与开发注意事项 / Operational Notes

## 12. 待确认问题 / Open Questions
```

For Notion append updates, wrap generated content under a timestamp heading:

```markdown
---

## 更新记录 / Update - <YYYY-MM-DD HH:mm>

<generated documentation>
```

### Notion sync convention

When the user explicitly requests Notion sync:

1. Run `notion_status` to verify configured state without printing secrets.
2. Determine the page title using the mode convention.
3. Use `notion_search` with the exact title or distinctive title fragment.
4. If a same-title page exists:
   - Use `notion_append_to_page` with a timestamped update section.
5. If no same-title page exists:
   - Use `notion_create_page` with the generated Markdown.
6. Do not expose `NOTION_TOKEN` or environment details.
7. If multiple candidates are found, choose an exact title match; otherwise ask user before publishing.

## Recommendation

Implement the MVP as **Pattern 1: a pure static Pi skill** under:

```text
.pi/skills/go-code-doc-notion/SKILL.md
```

Use a strong `description` so Pi loads it for requests involving Go code documentation, architecture/module summaries, Markdown docs, or Notion publishing. Include package-mode and module-mode workflows, the Markdown template, and explicit append-or-create Notion sync behavior.

Do **not** build a custom extension for MVP. Existing Notion tools already support the required search/create/append workflow, and `.pi/settings.json` already discovers project skills. Add helper files only if the template becomes too large; avoid executable helper scripts unless later acceptance criteria require deterministic symbol extraction or batch generation.

## Follow-up implementation notes

- Candidate skill name: `go-code-doc-notion`.
- Candidate path: `.pi/skills/go-code-doc-notion/SKILL.md`.
- The frontmatter description should mention:
  - Go code reading/summarization.
  - Package and cross-package module documentation.
  - Structured Markdown output.
  - Optional Notion sync using existing Notion tools.
- Manual validation can include:
  - Confirm skill frontmatter has valid `name` and `description`.
  - Confirm `.pi/settings.json` loads `./skills` and `enableSkillCommands` is true.
  - Invoke `/skill:go-code-doc-notion` in Pi after reload.
  - Test package mode on a small package such as `internal/errors` or `client`.
  - Test module mode without Notion sync first, then with sync after user confirmation.
