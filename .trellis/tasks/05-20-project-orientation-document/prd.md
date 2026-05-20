# Project Orientation Document

## Goal

Create a Chinese Markdown orientation document that follows the recommended reading order and records concrete findings about WeKnora's project purpose, architecture, modules, startup paths, development workflow, deployment entry points, and Trellis context.

## What I already know

* The user wants a step-by-step, concrete walkthrough following the previously recommended reading order.
* The user wants the results recorded in a Markdown file.
* WeKnora is an enterprise knowledge management framework centered on RAG, Agent reasoning, MCP tooling, and Wiki generation.
* Main components include Go backend, Vue frontend, Python DocReader, Python MCP server, Go CLI, Go SDK, miniprogram, Docker/Helm deployment, and Trellis workflow.
* Current Trellis task directory: `.trellis/tasks/05-20-project-orientation-document/`.

## Assumptions (temporary)

* The document should be written in Chinese.
* The document should be useful for future onboarding and not just a short summary.
* The document should avoid changing application behavior.

## Open Questions

* None.

## Requirements (evolving)

* Follow the recommended reading order item by item.
* For each item, explain what the files are for and what concrete project facts they reveal.
* Include key file paths and practical next-reading guidance.
* Generate one Markdown document at `docs/项目梳理.md` that records the final orientation results.

## Acceptance Criteria (evolving)

* [x] `docs/项目梳理.md` exists.
* [x] The document is in Chinese.
* [x] The document covers README/project positioning, development commands, backend startup chain, frontend startup chain, DocReader, MCP server, Trellis workflow/specs, and CLI conventions.
* [x] The document includes concrete file references.
* [x] No application code behavior is changed.

## Definition of Done

* Markdown document created or updated.
* Relevant files inspected read-only before writing the document.
* Trellis task was started for implementation after PRD confirmation.

## Out of Scope

* Running the app end-to-end.
* Changing source code.
* Completing the existing bootstrap guidelines task.
* Adding or modifying project specs unless separately requested.

## Technical Notes

* Key files from prior inspection: `README_CN.md`, `Makefile`, `cmd/server/main.go`, `frontend/package.json`, `.trellis/workflow.md`, `cli/AGENTS.md`.
* Need to inspect the remaining recommended files before writing the final document.
