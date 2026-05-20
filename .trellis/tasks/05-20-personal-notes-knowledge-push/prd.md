# Personal Notes Knowledge Push

## Goal

Add a personal notes capability to WeKnora so users can privately create, organize, manage, and search their own Markdown notes, then manually push selected notes to one or more target knowledge bases so the pushed content participates in existing RAG retrieval and question answering.

## Product Positioning

The MVP is **private personal notes first**, not a knowledge-base authoring surface first.

Users should feel that notes are their own private workspace. Pushing to a knowledge base is an explicit action, not an automatic consequence of editing or saving a note.

## Decisions

### MVP positioning

**Decision**: Private personal notes + manual push.

**Reason**: The user chose private notes first.

**Consequences**:

* Notes are managed independently from knowledge bases.
* Users explicitly select one or more knowledge bases when they want to push a note.
* Saving or editing a note does not automatically mutate knowledge-base content.

### Management scope

**Decision**: Include complete note management in MVP.

**Reason**: The user chose full management rather than minimal CRUD.

**Consequences**:

* MVP should include creation, editing, deletion, search/filtering, notebooks/categories, tags, batch operations, archive/recycle-bin style lifecycle, favorites/recents, and push status.
* Implementation should likely be split into backend foundation, frontend management UI, and knowledge-push integration.

### Push synchronization

**Decision**: Manual re-push semantics with stale marking for all previously successful targets, and push allowed only for active notes.

**Reason**: The user wants explicit control after publication, chose to mark all successfully pushed targets stale when the source note changes, and wants archived/recycle-bin notes excluded from publishing actions.

**Consequences**:

* Editing a pushed note only updates the private note.
* Only active notes can be pushed or re-pushed.
* Archived notes and recycle-bin/deleted notes cannot be pushed or re-pushed unless restored to active first.
* If a pushed note's title or content changes, all previously successful publication targets should be marked stale/outdated.
* Changes to note tags, notebook/category, favorite state, or other private organization metadata should not mark publications stale.
* Failed or never-pushed targets should keep their existing failed/never-pushed status when the note changes.
* Users must click re-push to update target knowledge-base content.
* Push/re-push should use the existing manual knowledge asynchronous ingestion flow rather than waiting for all indexing work in the request.
* Publication status should transition through pushing and then pushed/failed based on the result of creating/updating the target knowledge item and dispatching/processing the existing manual ingestion flow where practical.
* Successful re-push to a target should clear stale status for that target.
* If the previously mapped target knowledge item no longer exists during re-push, create a new knowledge item through the manual knowledge flow and update the publication mapping to the new `knowledge_id`.
* If the target knowledge base is deleted/unavailable or the user loses write permission, keep the publication record and mark it failed/unavailable with a user-visible reason.
* Deleting or archiving a note does not automatically delete knowledge items already created in knowledge bases.

### Push target cardinality and status display

**Decision**: Support multiple target knowledge bases per note, with aggregated list status and per-target detail status.

**Reason**: The user chose multi-knowledge-base push and prefers an aggregated status on note lists to avoid clutter.

**Consequences**:

* A separate publication mapping is required.
* Each note-target knowledge base pair needs independent status tracking.
* Note lists should show an aggregated push status such as never pushed, pushed, stale, pushing, failed, or partially failed.
* Note detail, push dialog, or expandable status view should show per-target knowledge-base status and error details.
* UI should allow re-pushing selected targets rather than forcing all targets to re-push.

### Privacy

**Decision**: Strictly private notes.

**Reason**: The user chose strict privacy.

**Consequences**:

* Only the note owner can read or mutate note content via note APIs.
* Tenant Owner/Admin should not be able to read private note bodies through note APIs.
* Tenant Owner/Admin can still manage knowledge items created from pushed notes through existing knowledge-base permissions.

### Information architecture

**Decision**: Use notebooks/categories + lightweight tags, with one default notebook per user.

**Reason**: The user chose full organization with both primary grouping and flexible labels, wants a default notebook so notes always have a safe home, and prefers lightweight tag management in MVP.

**Consequences**:

* Each note should belong to one notebook/category.
* Each user should have a default notebook within each tenant context.
* The default notebook can be renamed but cannot be deleted.
* The default notebook should be pinned/placed first in notebook lists.
* New notes without an explicit notebook should be placed in the default notebook.
* Each note can have multiple tags.
* Users can input/select tags while editing notes; missing tags should be created automatically for the current user/tenant.
* Deleting a tag should remove tag bindings only and should not delete notes.
* MVP does not require a separate full tag management page for rename/merge workflows.
* List pages should support filtering by notebook, tag, keyword, lifecycle status, favorite/recent, and push status.
* Keyword search should match both note title and Markdown body content in MVP; it may start with ordinary text matching before any later full-text-search optimization.

### Notebook deletion behavior

**Decision**: Deleting a notebook moves that notebook and its notes into recycle bin after explicit confirmation.

**Reason**: The user wants default-notebook support and prefers notebook deletion to cascade notes into recycle bin, but only after a confirmation dialog to reduce accidental bulk deletion.

**Consequences**:

* Notebook deletion must require an explicit confirmation dialog in the frontend.
* Deleting a notebook should soft-delete/archive-to-recycle-bin the notebook and all notes currently inside it.
* Notes deleted through notebook deletion should follow the same 30-day restore and purge retention behavior as individually deleted notes.
* Restoring a deleted notebook should restore its notes where practical; restoring an individual note whose notebook remains deleted should move the note to the default notebook or require selecting a target notebook.
* Previously pushed knowledge items are still not automatically deleted when notebooks or notes enter recycle bin.

### Favorites and recents

**Decision**: Reuse the existing generic favorites table/pattern for note favorites and the existing local recent-resource pattern for note recents.

**Reason**: `user_resource_favorites` is explicitly extensible by resource type, and `useResourcePins.ts` already separates DB-backed favorites from localStorage recents.

**Consequences**:

* Add note/notebook resource type constants and frontend support rather than creating note-specific favorite tables.
* Keep recents localStorage-based for MVP, scoped by user and tenant, matching the existing KB/agent pattern.
* Update note recents when the user opens/views a note detail.
* Do not update recents merely because a note appears in list/search results.

### Recycle-bin retention and destructive confirmations

**Decision**: Deleted notes enter recycle bin and are automatically purged after 30 days; destructive delete actions require confirmation.

**Reason**: The user chose 30-day automatic cleanup and wants deletion actions confirmed while keeping archive operations lightweight.

**Consequences**:

* Soft-deleted notes should remain restorable for 30 days.
* Archive/unarchive actions do not require confirmation.
* Single delete, batch delete, notebook delete, and permanent delete actions should require explicit confirmation.
* The UI should show remaining retention or deletion time where practical.
* Backend should add a cleanup runner or scheduled sweep similar in spirit to existing retention patterns such as audit-log retention / datasource sync-log cleanup.
* Manual permanent deletion can still be exposed for users who want immediate cleanup.

### Pushed knowledge title

**Decision**: Preserve the original note title without adding a visible prefix.

**Reason**: The user chose clean titles.

**Consequences**:

* Pushed knowledge should use the note title as-is.
* Note origin should be represented through `channel`, metadata, and publication records rather than title decoration.

### Pushed knowledge source metadata and visibility

**Decision**: Store internal source IDs in pushed knowledge metadata, and show only the source type in knowledge-base UI.

**Reason**: The user chose to preserve internal IDs for re-push/debugging and wants pushed knowledge to show that it came from a personal note without exposing private note identifiers or links.

**Consequences**:

* Pushed knowledge metadata should include internal fields such as `source_note_id`, `source_publication_id`, and `source_type=personal_note`.
* These fields are for internal linkage, re-push, audit, and debugging.
* Knowledge-base UI may display a human-readable source type such as "来源：个人笔记".
* Knowledge-base UI should not display private note IDs or source note links to users who cannot access the source note.
* MVP should not add a clickable "open source note" link from knowledge-base pages unless a later permission-aware design is added.

### Editor and import scope

**Decision**: Use a Markdown basic editor for MVP, reuse/generalize the existing manual knowledge editor, and support note import only.

**Reason**: The user chose the Markdown basic version, the project already has `frontend/src/components/manual-knowledge-editor.vue` for online Markdown knowledge editing, and the user wants to reference existing import logic while only adding import in MVP.

**Consequences**:

* Reuse or generalize `frontend/src/components/manual-knowledge-editor.vue` as the note Markdown editor foundation where practical.
* Prefer extracting a shared Markdown editor core if the existing component is tightly coupled to manual knowledge APIs or global UI store state.
* The note editor wrapper should call note create/update APIs, not manual knowledge create/update APIs directly.
* Support Markdown editing, preview, basic toolbar actions, and manual save.
* Support importing note content, initially focused on Markdown/text files.
* Imported note title should use the source file name by default.
* Imported note content should use the source file content.
* Imported notes should be created in the currently selected notebook when available, otherwise in the user's default notebook.
* Import UI should allow selecting one or more tags to apply to all notes in that import operation.
* MVP import should not parse tags from Markdown frontmatter or file content.
* Notes with the same title should be allowed within the same notebook; use creation/update time and note IDs for distinction rather than enforcing title uniqueness.
* Reuse existing frontend upload/import patterns from knowledge file import and existing backend parsing behavior where practical, but imported content should create private notes rather than knowledge-base knowledge items.
* Do not include note export in MVP.
* Exclude auto-save, attachments/images, and AI editing helpers from MVP unless added later.

### Batch operation failure behavior

**Decision**: Keep partial successes and track failures independently for both batch push and batch import.

**Reason**: The user chose partial-success semantics for batch push and batch import.

**Consequences**:

* Each note-target publication should be processed independently.
* Successful publications should remain successful even if other targets fail.
* Failed publications should store error details and support retry.
* Note-level push status should be derived from per-target publication records for list filtering/display.
* Batch note import should preserve successfully imported notes when other files fail.
* Failed imported files should show per-file failure reasons and support retry where practical.
* The UI should summarize success/failure counts and show per-target or per-file failure reasons.

## Requirements

### Backend requirements

* Add user-owned, tenant-scoped notes.
* Notes must derive `tenant_id` and `user_id` from authenticated context, not request body.
* Notes are strictly private: all note CRUD queries must filter by both `tenant_id` and `user_id`.
* Add notebooks/categories for primary note organization.
* Ensure each user has a default notebook per tenant, and create notes in the default notebook when no notebook is selected.
* The default notebook can be renamed but cannot be deleted.
* Add lightweight note tags for flexible classification.
* Users can create tags by entering them during note edit/import flows.
* Deleting a tag removes tag bindings and does not delete notes.
* Support note lifecycle states such as active, archived, and deleted/recycle-bin.
* Deleting a notebook should require confirmation and move the notebook plus contained notes into recycle bin.
* Notes in recycle bin should be restorable for 30 days and then automatically purged.
* Support note favorites and recents.
* Note recents should update when the user opens/views note detail, not when a note merely appears in a list.
* Support list/search/filter/pagination.
* Note lists should default to recently updated first (`updated_at DESC`).
* Notebook lists should place the default notebook first.
* Keyword search should cover note titles and Markdown body content.
* Support batch operations, including at minimum batch archive/delete/restore and batch push.
* Delete, batch delete, notebook delete, and permanent delete actions should require confirmation; archive/unarchive actions do not require confirmation.
* Support importing Markdown/text files as private notes, reusing existing import/parsing patterns where practical.
* Batch note import should preserve successfully imported notes when other files fail, and expose per-file failure reasons.
* For imported notes, use the file name as the default title, the file content as the note body, and the selected/default notebook as the destination.
* Import UI should allow selecting tags to apply to all imported notes in that operation; do not parse tags from file content in MVP.
* Allow duplicate note titles within the same notebook; do not block create/import or auto-rename solely because of title collision.
* Support manual push of one active note or multiple active notes to one or more target knowledge bases.
* Archived and recycle-bin/deleted notes should not be pushable or re-pushable until restored to active.
* Batch push should preserve partial successes and record per-note-target failures with retryable status.
* Push must verify both:
  * the caller owns the note; and
  * the caller has write access to each target knowledge base.
* If a target knowledge base is unavailable/deleted or write access is lost, preserve the note-target publication record and mark it failed/unavailable with an explanatory error.
* Pushed notes should become normal knowledge-base knowledge items.
* Push should reuse the existing manual knowledge ingestion path instead of writing chunks, embeddings, or vector indexes directly.
* Push/re-push should follow the existing asynchronous manual knowledge processing model and should not block the request until all indexing work is complete.
* Re-push should update the corresponding knowledge item for that note-target pair where possible.
* If the mapped knowledge item no longer exists, re-push should create a new target knowledge item and update the publication mapping.
* Editing a pushed note title or content should mark all previously successful publication records as stale/outdated.
* Editing note organization metadata such as tags, notebook/category, or favorite state should not mark publication records stale.
* Failed or never-pushed publication records should keep their status when the source note changes.
* Deleting/archiving a note should not automatically delete pushed knowledge items.

### Frontend requirements

* Add a personal notes entry under the platform area.
* Add routes such as `/platform/notes` and notebook/note detail routes or query-driven detail mode.
* Provide notebook/category management, including a visible default notebook that can be renamed but not deleted.
* Notebook deletion must show an explicit confirmation dialog before moving the notebook and contained notes into recycle bin.
* Provide lightweight tag selection/creation in note edit/import flows.
* Provide tag filtering and tag deletion behavior that removes bindings without deleting notes.
* Provide note list with search across title/body, filters, tags, lifecycle status, favorite/recent views, push status, and recently-updated-first default sorting.
* Provide a Markdown basic note editor, reusing or generalizing the existing manual knowledge editor where practical.
* The MVP editor supports Markdown editing, preview, basic toolbar actions, and manual save only.
* Provide note detail/preview.
* Provide batch selection and batch actions.
* Show confirmation dialogs for delete, batch delete, notebook delete, and permanent delete; archive/unarchive can proceed without confirmation.
* Provide note import UI for Markdown/text files, referencing existing knowledge upload/import interaction patterns.
* Import UI should allow selecting destination notebook and tags for the current import batch.
* Batch import UI should summarize success/failure counts and show per-file failure reasons.
* Provide push dialog for selecting one or more target knowledge bases.
* Show aggregated push status on note list rows, including never pushed, pushed, stale/outdated, pushing, failed, and partially failed states.
* Show per-target push status in note detail, push dialog, or expandable status view.
* Provide manual re-push action for selected target knowledge bases, enabled only when the source note is active.

### Permission requirements

* Note APIs must not expose private note content to other tenant members or tenant admins.
* Push APIs must reuse existing target-KB write permission behavior.
* For target knowledge bases, existing guards such as `OwnedKBOrAdmin()` and `KBAccessWrite("id")` should be reused where possible.
* Shared-KB push must distinguish caller tenant/user for reading private notes from effective target-KB tenant for creating knowledge.

## Acceptance Criteria

* [ ] A user can create, edit, view, archive, restore, and delete their own notes.
* [ ] A user cannot view or mutate another user's private notes in the same tenant.
* [ ] Notes can be organized by notebook/category and lightweight tags.
* [ ] Users can create tags while editing/importing notes, and deleting tags removes bindings without deleting notes.
* [ ] New notes can be created in a default notebook when no notebook is selected.
* [ ] The default notebook can be renamed but cannot be deleted.
* [ ] The default notebook appears first in notebook lists, and note lists default to recently updated first.
* [ ] Deleting a non-default notebook requires confirmation and moves the notebook plus contained notes into recycle bin.
* [ ] The note editor supports Markdown editing, preview, basic toolbar actions, and manual save.
* [ ] Notes can be searched by title and Markdown body content, and filtered by keyword, notebook, tag, lifecycle status, favorite/recent, and push status.
* [ ] Favorites use the existing generic user-resource favorite pattern; recents are stored locally per user and tenant.
* [ ] Note recents update when users open/view note detail and not from list/search exposure alone.
* [ ] Deleted notes enter recycle bin, can be restored within 30 days, and are automatically purged after 30 days.
* [ ] Delete, batch delete, notebook delete, and permanent delete actions require confirmation; archive/unarchive actions do not.
* [ ] A user can select one or more notes and batch archive/delete/restore/push them.
* [ ] A user can import Markdown/text files as private notes.
* [ ] Batch note import preserves successful imports when other files fail and shows per-file failure reasons.
* [ ] Imported notes use the file name as the default title and the file content as the note body.
* [ ] Import UI can apply selected notebook and tags to all notes in the current import batch.
* [ ] Duplicate note titles are allowed within the same notebook for create/import flows.
* [ ] Note import creates private notes and does not automatically create or update knowledge-base content.
* [ ] A user can manually push an active note to multiple knowledge bases.
* [ ] Archived and recycle-bin/deleted notes cannot be pushed or re-pushed until restored to active.
* [ ] Batch push preserves successful note-target publications when other targets fail.
* [ ] Failed note-target publications show error details and can be retried.
* [ ] Push fails for target knowledge bases where the user lacks write permission.
* [ ] If target knowledge bases are deleted/unavailable or write access is lost, existing publication records are preserved and marked failed/unavailable with a visible reason.
* [ ] Successful push creates or updates normal knowledge-base knowledge items using the existing asynchronous manual knowledge ingestion flow.
* [ ] Push/re-push requests do not directly write chunks, embeddings, or vector indexes and do not wait for all indexing work to finish before returning.
* [ ] Pushed knowledge uses the original note title without a visible source prefix.
* [ ] Pushed knowledge metadata preserves internal source note/publication IDs for re-push and debugging.
* [ ] Knowledge-base UI can show pushed-note source type without exposing private note IDs or source-note links to unauthorized users.
* [ ] Editing a pushed note title or content marks all previously successful publications stale/outdated but does not auto-update the knowledge base.
* [ ] Editing note tags, notebook/category, or favorite state does not mark publications stale.
* [ ] Re-push updates the corresponding target knowledge item and clears stale status for that target.
* [ ] If the mapped target knowledge item has been deleted, re-push creates a replacement knowledge item and updates the publication mapping.
* [ ] Deleting or archiving a note does not automatically delete previously pushed knowledge items.
* [ ] The frontend displays aggregated push status on note lists and per-target push status in detail/push surfaces.

## Recommended Technical Approach

### Existing backend reuse

Relevant existing files:

* `internal/router/router.go`
* `internal/handler/knowledge.go`
* `internal/application/service/knowledge_create.go`
* `internal/application/service/knowledge_process.go`
* `internal/application/service/knowledge_post_process.go`
* `internal/types/knowledge.go`
* `internal/types/task.go`
* `internal/middleware/auth.go`
* `internal/middleware/rbac.go`
* `internal/middleware/kb_access.go`
* `internal/router/rbac.go`
* `internal/application/service/audit_log*.go`
* `internal/application/repository/datasource_repo.go`
* `internal/container/container.go`

Recommended reuse:

* Use `KnowledgeService.CreateKnowledgeFromManual` for first push.
* Use `KnowledgeService.UpdateManualKnowledge` for re-push to an existing target knowledge item where possible.
* Do not directly write `knowledges`, `chunks`, `embeddings`, or retriever indexes from the note feature.
* Let the existing manual knowledge flow handle Markdown cleaning, chunking, indexing, and post-processing.

### Suggested backend modules

Potential new files:

* `internal/types/user_note.go`
* `internal/types/user_note_notebook.go`
* `internal/types/user_note_tag.go`
* `internal/types/user_note_publication.go`
* `internal/types/interfaces/user_note.go`
* `internal/application/repository/user_note.go`
* `internal/application/service/user_note.go`
* `internal/handler/user_note.go`
* `internal/router/user_note.go`

Potential migrations:

* `migrations/versioned/000052_user_notes.up.sql`
* `migrations/versioned/000052_user_notes.down.sql`

Exact migration number should be adjusted to the next available version at implementation time.

### Suggested data model

Core tables:

* `user_note_notebooks`
* `user_notes`
* `user_note_tags`
* `user_note_tag_bindings`
* `user_note_publications`

`user_notes` likely fields:

* `id`
* `tenant_id`
* `user_id`
* `notebook_id`
* `title`
* `content`
* `status` (`active`, `archived`, possibly `deleted` with soft delete)
* `metadata`
* `created_at`
* `updated_at`
* `deleted_at`

`user_note_publications` likely fields:

* `id`
* `tenant_id`
* `user_id`
* `note_id`
* `knowledge_base_id`
* `knowledge_id`
* `target_tenant_id`
* `status` (`never_pushed`, `pushing`, `pushed`, `stale`, `failed`)
* `last_pushed_note_updated_at`
* `published_at`
* `last_error`
* `created_at`
* `updated_at`

### Existing frontend reuse

Relevant files:

* `frontend/src/router/index.ts`
* `frontend/src/components/manual-knowledge-editor.vue`
* `frontend/src/views/knowledge/KnowledgeBaseList.vue`
* `frontend/src/views/knowledge/KnowledgeBase.vue`
* `frontend/src/views/knowledge/components/DocumentListView.vue`
* `frontend/src/api/knowledge-base/index.ts`
* `frontend/src/hooks/useKnowledgeBase.ts`
* `frontend/src/stores/ui.ts`
* `frontend/src/composables/useListUrlState.ts`
* `frontend/src/composables/useResourcePins.ts`
* `internal/types/user_resource_favorite.go`
* `frontend/src/api/user-favorites.ts`
* `frontend/src/components/GlobalCommandPalette.vue`
* `frontend/src/components/MentionSelector.vue`

Recommended frontend direction:

* Add `frontend/src/api/notes/index.ts`.
* Add a lightweight notes store similar to `frontend/src/stores/knowledge.ts`.
* Reuse/generalize `frontend/src/components/manual-knowledge-editor.vue` for Markdown editing; if needed, split it into a shared Markdown editor core plus knowledge/note-specific wrappers.
* Keep note editor persistence separate from manual knowledge persistence: note saves update private notes only, and push/re-push remains a separate explicit action.
* Treat other existing `EditorDialog`/`EditorModal` components as configuration dialogs, not content-editor foundations.
* Reuse knowledge import/upload interaction patterns from `frontend/src/views/knowledge/KnowledgeBase.vue`, `frontend/src/api/knowledge-base/index.ts`, and `frontend/src/hooks/useKnowledgeBase.ts` for note import UI where practical.
* Note import should target note APIs and private note storage, not knowledge-base import endpoints.
* Reuse `DocumentListView.vue` patterns for note list, row actions, and batch selection.
* Reuse `useListUrlState.ts` for filters and search state.
* Extend `user_resource_favorites` valid resource types plus `useResourcePins.ts` for note/notebook favorites; keep recents localStorage-based for MVP.

## Out of Scope

* Browser extension clipping.
* Mobile/miniprogram note editing.
* Collaborative shared notes.
* Sharing private notes directly with other users.
* Full offline note sync.
* Editor auto-save and conflict resolution.
* Note export.
* Note attachments/image uploads in the editor.
* AI-assisted note editing helpers.
* Automatic sync on every note save.
* Auto-deleting pushed knowledge when a note is deleted.
* A separate note-only retrieval/vector pipeline.
* Direct retrieval over private notes unless separately designed.

## Implementation Decomposition

Recommended implementation phases:

1. Backend data model and private note CRUD.
2. Notebook/tag/favorite/recent/lifecycle and batch APIs.
3. Publication model and push/re-push APIs using existing manual knowledge service.
4. Frontend notes shell, list, filters, and editor.
5. Frontend push dialog and per-target publication status.
6. Quality pass: permission tests, push permission tests, stale-status tests, frontend manual verification.

## Open Questions

* None.
