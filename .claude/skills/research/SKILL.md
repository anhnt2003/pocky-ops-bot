---
name: research
description: >
  Research online documents from URLs and produce high-quality markdown reference files
  that AI Agents can use as reliable context for decision-making and action.
  Use this skill whenever the user provides a URL and wants to extract, organize, or
  document its content — including technical docs, blog posts, API references, tutorials,
  guides, release notes, or any web page. Also use when the user says things like
  "research this", "read this page", "document this URL", "create a reference from this",
  "fetch and summarize", or "I need context from this link".
---

# Research Skill

Turn web URLs into structured, accurate markdown reference documents that AI Agents can rely on for context and decision-making.

## Why this skill exists

AI Agents perform best when given clean, well-structured context documents. Raw web pages are noisy — ads, navigation, sidebars, inconsistent formatting. This skill bridges the gap: it fetches web content and transforms it into reference-grade markdown that preserves accuracy while maximizing readability and usefulness for downstream AI consumption.

## Workflow

### 1. Fetch and analyze the source

Use `WebFetch` to retrieve the page content. Read through it carefully and assess:

- **Content type**: Is this an API reference? A tutorial? A conceptual guide? A changelog? A blog post?
- **Scope**: How much material is there? Is it a single focused topic or a broad multi-section document?
- **Depth**: Does it contain code examples, configuration details, data tables, or step-by-step procedures?
- **Cross-references**: Does it reference other pages that are essential context? (If so, note them — the user may want those fetched too.)

### 2. Decide on file structure

**Single file** when:
- The content covers one focused topic (e.g., a single API endpoint, one tutorial, a blog post)
- Total content fits comfortably under ~500 lines of markdown

**Multiple files** when:
- The source covers distinct domains or sections (e.g., an API reference with auth, endpoints, webhooks, errors)
- The content is large enough that a single file would be hard to navigate
- Different sections would naturally be consumed independently by an agent

When splitting, create:
- An `index.md` or `README.md` that provides an overview and links to each sub-file
- Individual files named descriptively (e.g., `authentication.md`, `endpoints.md`, `error-codes.md`)

### 3. Extract and structure the content

This is the critical step. The goal is not to summarize — it is to **accurately capture** the information in a format optimized for AI consumption.

#### Content principles

- **Accuracy over brevity**: Never paraphrase in a way that loses precision. If a parameter is described as "a string of 1-128 characters matching `[a-zA-Z0-9_-]+`", preserve that exact specification — don't simplify to "a short string".
- **Preserve code examples**: Copy code blocks exactly as they appear. They are often the most valuable part of a technical document for an AI Agent that needs to generate or modify code.
- **Preserve numbers, versions, and constraints**: These details matter for correctness. "Supports up to 100 concurrent connections" is meaningful; "supports many connections" is not.
- **Remove noise, keep signal**: Strip navigation elements, ads, promotional content, cookie banners, and other non-informational content. But keep warnings, deprecation notices, version compatibility notes, and caveats — these are critical context.
- **Translate implicit knowledge into explicit statements**: If a tutorial assumes you've already set up a database but doesn't state it, add a "Prerequisites" section that makes this explicit.

#### Structural patterns

Use these markdown patterns to maximize readability for both humans and AI:

**Metadata header** — Every output file starts with this:

```markdown
# [Document Title]

> Source: [original URL]
> Fetched: [date]
> Content type: [API Reference | Tutorial | Guide | Blog Post | Documentation | etc.]

[1-3 sentence overview of what this document covers and when an agent should reference it]
```

**For API references**, organize by:
```markdown
## Authentication
## Endpoints
### [HTTP Method] /path
- **Description**: ...
- **Parameters**: (use tables)
- **Request example**: (code block)
- **Response example**: (code block)
- **Error codes**: (table or list)
## Rate Limits
## Error Handling
```

**For tutorials/guides**, preserve the sequential flow:
```markdown
## Prerequisites
## Step 1: [Action]
## Step 2: [Action]
## Common Issues
## Key Takeaways
```

**For conceptual docs**, organize by topic:
```markdown
## Overview
## Core Concepts
### [Concept A]
### [Concept B]
## How It Works
## Configuration
## Limitations
```

**For changelogs/release notes**:
```markdown
## [Version] — [Date]
### Breaking Changes
### New Features
### Bug Fixes
### Deprecations
```

Use **tables** for structured data (parameters, options, comparisons). Use **code blocks** with language tags for all code. Use **blockquotes** (`>`) for important warnings or notes from the original source.

### 4. Quality checklist

Before saving, verify:

- [ ] All code examples are preserved exactly (not paraphrased or truncated)
- [ ] All numbers, version constraints, and limits are accurate
- [ ] No content was hallucinated — everything traces back to the source
- [ ] Warnings, deprecation notices, and caveats are included
- [ ] The document has a clear metadata header with source URL and fetch date
- [ ] The structure matches the content type (API → endpoint-focused, tutorial → sequential, etc.)
- [ ] If split into multiple files, the index file has a clear overview and links

### 5. Save the output

Save markdown files to the location the user specifies (or the current working directory if none specified). Use descriptive filenames:

- Single file: `{topic-name}.md` (e.g., `stripe-webhooks.md`, `react-server-components.md`)
- Multi-file: `{topic-name}/index.md` + `{topic-name}/{section}.md`

After saving, report to the user:
- What files were created and where
- A brief summary of what was captured
- Any important caveats (e.g., "The page had dynamic content that couldn't be fully fetched", "Referenced pages X and Y may also be worth researching")

## Handling edge cases

- **Dynamic/JS-rendered pages**: WebFetch may not capture JS-rendered content. If the result looks incomplete or empty, tell the user and suggest they provide the content another way (paste it, use a different URL, or provide a cached/static version).
- **Very long pages**: If content exceeds what can be reasonably fetched in one call, focus on the most information-dense sections and note what was omitted.
- **Multiple URLs**: If the user provides several URLs on the same topic, fetch all of them and produce a unified reference that cross-references the sources rather than creating isolated documents for each.
- **Non-English content**: Preserve the original language unless the user asks for translation. Add the language to the metadata header.
- **Paywalled or access-restricted content**: If the fetch fails or returns a login page, inform the user immediately rather than producing a document from the error page.
