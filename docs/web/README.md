# PentLog Documentation

Documentation site for PentLog — Evidence-first pentest logging tool.
Source content lives in this repository under `docs/web/docs/` and is served via Zensical.

**Live Site:** https://pentlog.petruknisme.com

## Quick Start

### Using Zensical (Recommended)

```bash
# Install Zensical
pip install zensical

# Clone repository
git clone https://github.com/aancw/pentlog.git
cd pentlog/docs/web

# Serve locally
zensical serve --dev-addr 0.0.0.0:8000

# Build
zensical build
```

### Using MkDocs (Legacy)

```bash
# Install dependencies
pip install mkdocs-material mkdocs-minify-plugin

# Serve locally
mkdocs serve --dev-addr 0.0.0.0:8000

# Build
mkdocs build
```

## Deploy to Cloudflare Pages

1. Connect repository to Cloudflare Pages
2. Build settings:
   - **Build command:** `zensical build` (or `mkdocs build`)
   - **Build output:** `site`

## Structure

```
docs/web/docs/
├── index.md                    # Home page
├── getting-started/            # Getting started guides
├── guide/                      # User guides
├── advanced/                   # Advanced topics
└── reference/                  # CLI reference
```

## License

MIT License
