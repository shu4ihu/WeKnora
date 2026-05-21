# Go Code Documentation Notion Skill

## Goal

Build a Pi skill that helps read and summarize Go code into structured Markdown documentation, then uses the configured Notion extension/tools to sync the generated documentation into the target Notion database.

## What I already know

* The user wants a skill for reading Go code, producing Markdown docs, and syncing to a corresponding Notion database through the Notion extension.
* A Trellis task was created at `.trellis/tasks/05-21-go-code-doc-notion-skill`.
* Pi skills are discovered from `.pi/skills/` via `SKILL.md` files with required `name` and `description` frontmatter.
* Project `.pi/settings.json` already enables skill commands and loads project skills from `./skills`.
* The Notion extension is configured: `NOTION_TOKEN` is present, `NOTION_DATABASE_ID` is configured, and title property is `文档名`.
* The target Notion database already contains WeKnora backend module summary pages, including an index page and numbered module docs.
* This repository contains multiple Go modules: root `github.com/Tencent/WeKnora`, `client`, and `cli`.
* Backend Go code is spread across `cmd/`, `internal/`, `cli/`, and `client/`, with many domain modules under `internal/application`, `internal/handler`, `internal/router`, `internal/types`, etc.

## Assumptions (temporary)

* The MVP should be a project-local Pi skill under `.pi/skills/`, not a product feature in WeKnora backend/frontend.
* The skill should orchestrate existing tools (`read`, `bash`, `notion_*`) rather than creating a custom TypeScript extension unless needed.
* The generated docs should be human-readable architecture/module docs, not API reference generated solely from Go doc comments.
* The skill should support both single Go package/directory documentation and business-module documentation that can span multiple packages/layers.
* Generated Markdown should use a Chinese-first, bilingual technical style: Chinese section headings and explanations, while preserving Go package names, types, functions, API names, and file paths in English/code format.
* Sync should search for an existing same-title Notion page and append a new update section when found; otherwise create a new page.

## Open Questions

* None — requirements are ready for final confirmation.

## Requirements (evolving)

* Provide a Pi skill with a clear trigger description for Go code reading/documentation and Notion publishing.
* The skill must guide the agent to inspect Go code directly before asking the user for code context.
* The skill must output structured Markdown suitable for Notion conversion.
* The skill must normalize generated Markdown content with medium strictness: stable title, metadata, section order, citation, table, code-fence, unknowns, and Notion append format are required, while irrelevant sections may be omitted when clearly not applicable.
* The default documentation template should match the existing Notion backend-module summary style, adapted for package mode and module mode.
* The skill must use the configured Notion extension/tools for publishing when the user requests sync.
* For Notion sync, the MVP must search for an existing same-title page and append a timestamped update section when found; if no matching page exists, it must create a new page.
* The skill must avoid exposing secrets such as Notion tokens.
* The skill should support WeKnora's multi-module Go layout.
* The skill must offer two documentation scopes: package mode for one Go directory/package, and module mode for a cross-package feature/domain slice.
* Notion page titles should use mode-specific conventions:
  * Module mode: `WeKnora 后端模块总结 - <模块名>`
  * Package mode: `WeKnora Go Package - <package path>`

## Acceptance Criteria (evolving)

* [ ] A user can invoke the skill to document a single Go package/directory and receive a Markdown document.
* [ ] A user can invoke the skill to document a cross-package business module and receive a Markdown document.
* [ ] A user can request Notion sync and the agent publishes the Markdown to the configured Notion database.
* [ ] The skill includes a reusable documentation template with sections such as purpose, key files, data flow, public APIs, dependencies, and operational notes.
* [ ] Generated Markdown follows normalized formatting rules for headings, metadata, source citations, tables, flow diagrams, unknowns, and Notion append sections.
* [ ] The skill defines a safe inspection workflow using repository tools before asking questions.
* [ ] The skill documents Notion sync behavior clearly: search same-title page, append timestamped update if found, otherwise create a new page.
* [ ] Package-mode and module-mode docs use distinct Notion title conventions.

## Definition of Done (team quality bar)

* Tests/checks added or updated where appropriate.
* Skill validates against Pi skill frontmatter/structure expectations.
* Relevant Pi docs/examples consulted.
* Manual usage instructions included.
* No secrets committed.

## Out of Scope (explicit)

* Building a WeKnora backend/frontend feature.
* Automatically syncing on every code change unless explicitly selected later.
* Modifying the Notion extension itself unless current tools are insufficient.
* Parsing every Go AST symbol perfectly in MVP unless selected later.
* Batch documentation generation/sync for multiple packages/modules.
* Strong Notion sync guardrails beyond exact-title search plus append/create behavior, unless multiple candidates force clarification.

## Research References

* [`research/pi-skill-notion-workflow.md`](research/pi-skill-notion-workflow.md) — recommends MVP as a pure static Pi skill under `.pi/skills/go-code-doc-notion/SKILL.md`, using existing Notion tools for append-or-create publishing.
* [`research/go-code-doc-templates.md`](research/go-code-doc-templates.md) — defines WeKnora-specific package-mode and module-mode documentation templates, code inspection workflow, title conventions, and Notion append wrapper.

## Research Notes

### Feasible approaches

**Approach A: Pure static Pi skill** (Recommended)

* How it works: create `.pi/skills/go-code-doc-notion/SKILL.md` with workflow instructions, templates, and Notion sync procedure.
* Pros: simplest, safest, no runtime code, works with current `.pi/settings.json` and existing `notion_*` tools.
* Cons: relies on agent following instructions; no deterministic AST parser or replace-page upsert.

**Approach B: Static skill plus helper templates/scripts**

* How it works: add `SKILL.md` plus optional template/reference files or inspection scripts.
* Pros: more repeatable if templates grow; can reduce prompt length.
* Cons: helper scripts add maintenance and cross-platform concerns.

**Approach C: Skill plus custom Notion upsert extension**

* How it works: add TypeScript extension/tool to encapsulate search/append/create or future replace semantics.
* Pros: more deterministic publishing.
* Cons: higher complexity and not needed for current append-or-create MVP.

## Technical Approach (evolving)

* Implement MVP as a pure static Pi skill: `.pi/skills/go-code-doc-notion/SKILL.md`.
* Include package-mode and module-mode workflows in the skill.
* Include the Markdown templates directly in `SKILL.md` unless the file becomes too large.
* Use current Notion tools for publishing: verify status, search same-title page, append timestamped update if found, otherwise create page.
* Do not add a custom TypeScript extension for MVP.
* Keep MVP scope minimal: package/module modes, normalized Markdown generation guidance, and Notion exact-title append/create sync only.

## Decision (ADR-lite)

**Context**: The task needs a Pi capability for Go code documentation and Notion publishing. Pi supports both static skills and TypeScript extensions, and the existing Notion tools already support status/search/create/append operations.

**Decision**: Implement a pure static project skill under `.pi/skills/go-code-doc-notion/SKILL.md`.

**Consequences**: This is simple, safe, reviewable, and works with existing `.pi/settings.json`; the trade-off is that deterministic AST extraction, batch sync, and true Notion page replacement remain future work.

## Technical Notes

* Read Pi docs: `docs/skills.md`, `docs/extensions.md`.
* Read examples: `examples/extensions/README.md`, `examples/extensions/dynamic-resources/index.ts`, `examples/extensions/hello.ts`, `examples/skills/pdf-processing/SKILL.md`, project `examples/skills/README.md`.
* Skill locations: project `.pi/skills/` and configured `.pi/settings.json` `skills: ["./skills"]`.
* Extension examples show that custom tools require TypeScript extensions; a pure skill can simply instruct the agent to use existing tools.
* Notion tools available in this session include status/search/read/query/create/append.
* Notion database query showed existing pages named like `WeKnora 后端模块总结 - XX ...`.
* If robust upsert/replace is required, a custom extension/tool may be needed because current exposed Notion tools do not show a direct page content replacement API.
