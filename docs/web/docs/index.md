---
template: home.html
title: PentLog — Evidence-First Pentest Logger
hide:
  - navigation
  - toc
hero:
  title: PentLog
  subtitle: Evidence-first penetration testing logger.
  description: Capture every command, find anything, prove everything. High-fidelity terminal logging with AI analysis, searchable content, and compliance-ready reports.
  install_button: Getting Started
features:
  - title: High-Fidelity Recording
    description: Full-fidelity ttyrec capture with ANSI color, cursor moves, and redraws preserved.
  - title: Powerful Search
    description: Regex + boolean search across logs and notes with fast incremental results.
  - title: Compliance Reports
    description: Markdown and HTML reports with hashes, audit trails, and AES-256 archives.
  - title: AI Analysis
    description: Summaries and vulnerability insights from Gemini or local Ollama models.
  - title: Live Sharing
    description: Stream sessions in real time with an xterm.js viewer and full history.
  - title: Crash Recovery
    description: Heartbeats and stale-session detection protect evidence from crashes.
  - title: Timeline Browser
    description: Browse chronological command timelines, view output, and export JSON.
  - title: Context-Aware Sessions
    description: Organize by client, engagement, and phase with auto metadata tracking.
  - title: Vulnerability Management
    description: Track findings with severity, remediation notes, and session traceability.
---

<div class="mdx-split">
  <div class="mdx-split__content">
    <h2>Start in minutes</h2>
    <p>Install once, create a context, and begin recording. </p>
    <p>The same workflow scales from labs to full client engagements.</p>
    <div class="mdx-inline-links">
      <a href="getting-started/quickstart/" class="md-button md-button--primary">Quickstart</a>
      <a href="getting-started/installation/" class="md-button">Installation</a>
    </div>
  </div>
  <div class="mdx-code-card">
    <h4 class="mdx-code-card__label">Install + first session</h4>
    <pre><code class="language-bash">curl -sSf https://raw.githubusercontent.com/aancw/pentlog/main/install.sh | sh
pentlog setup
pentlog create
pentlog shell</code></pre>
  </div>
</div>

<div class="mdx-signal-grid">
  <div class="mdx-signal-card">
    <h3>Evidence Chain</h3>
    <p>Full-fidelity ttyrec recordings, context metadata, and integrity hashes keep your chain of custody intact.</p>
  </div>
  <div class="mdx-signal-card">
    <h3>Session Intelligence</h3>
    <p>Search across logs and notes, extract timelines, and pinpoint exactly what happened and when.</p>
  </div>
  <div class="mdx-signal-card">
    <h3>Delivery Ready</h3>
    <p>Generate Markdown or HTML reports, archive with AES-256 encryption, and ship in client-ready format.</p>
  </div>
</div>

<div class="mdx-resource-grid">
  <a class="mdx-resource-card" href="getting-started/quickstart/">
    <h3>Getting Started</h3>
    <p>Learn the core workflow and create your first context.</p>
  </a>
  <a class="mdx-resource-card" href="guide/sessions/">
    <h3>User Guide</h3>
    <p>Sessions, search, notes, timeline, export, and AI analysis.</p>
  </a>
  <a class="mdx-resource-card" href="advanced/archiving/">
    <h3>Advanced</h3>
    <p>Archiving, crash recovery, configuration, and storage layout.</p>
  </a>
  <a class="mdx-resource-card" href="reference/commands/">
    <h3>Reference</h3>
    <p>CLI commands, flags, and tool comparisons.</p>
  </a>
</div>

<div class="mdx-footnote">
  <p>Found a bug or want a feature? <a href="https://github.com/aancw/pentlog/issues">Open an issue</a>. PentLog is licensed under the MIT License.</p>
</div>
