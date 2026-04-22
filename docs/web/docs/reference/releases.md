# Release Notes

## v0.18.0 (2026-04-22)

PentLog v0.18.0 introduces a major browser **Web Dashboard** refresh focused on operational clarity, dark-mode readability, and direct workflow actions.

### Added

- **Web Dashboard Mission Control**
  - At-a-glance layout for active context, session health, findings, and artifact readiness
  - Quick links to Sessions, Search, Reports, and Archives
- **Unified Dashboard API**
  - New `GET /api/dashboard/overview` endpoint
  - Consolidates dashboard stats, activity, context, and artifact metadata
- **Scoped Web Flow Prefill**
  - Sessions, Search, and Reports pages now accept context-prefill query parameters

### Changed

- Dashboard information architecture prioritized around operational status first
- Improved keyboard accessibility and focus visibility in web UI
- Theme token refinements for better dark/light consistency

### Fixed

- Dark-mode contrast issues in dashboard state blocks (including live-share empty state)
- Session hydration reliability for `state`, `last_sync_at`, `target`, and `target_ip`

---

## v0.17.0 (2026-04-04)

PentLog v0.17.0 introduced multi-target engagement management.

### Added

- New `pentlog target` command:
  - Add, list, switch, remove, and clear target entries
  - Target storage in `~/.pentlog/targets.json`

### Changed

- Target context now flows through:
  - Active context
  - Session metadata (JSON + database)
  - Shell prompt and session naming
