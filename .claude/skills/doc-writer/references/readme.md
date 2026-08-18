# README Writing Guide

## Structure (in order)

```markdown
# Project Name
> One-line tagline — what it does and who it's for

## Overview
2–4 sentences. What problem does this solve? What does it do?

## Features
- Bullet list of key capabilities (5–8 max)

## Installation
\`\`\`bash
npm install my-package   # JS/TS
# or
go get github.com/org/repo   # Go
\`\`\`

## Quick Start
Minimal working example. Copy-pasteable. No setup steps buried in prose.

\`\`\`ts
// JS/TS example
import { thing } from 'my-package'
thing.doSomething()
\`\`\`

\`\`\`go
// Go example
import "github.com/org/repo"
repo.DoSomething()
\`\`\`

## API Reference
Link to generated docs or summarize key exports here.

## Configuration (if applicable)
Table of options: | Option | Type | Default | Description |

## Contributing (optional)
Brief guide or link to CONTRIBUTING.md

## License
MIT / Apache 2.0 / etc.
```

## Tone & Style Rules

- **Active voice**: "Fetches the user record" not "The user record is fetched"
- **Present tense**: "Returns a string" not "Will return a string"
- **No fluff**: No "This amazing library...", no "Easy to use!"
- **Code examples first**: Show before you tell. Readers want to see it work.
- **Concrete over abstract**: "Retries failed requests up to 3 times" > "Handles errors gracefully"
- Tailor length to project complexity — a utility library needs less than a platform

## Inferring Missing Info

If the repo doesn't have a description, infer from:
1. Package name + entry point exports
2. Folder structure (e.g., `handlers/`, `models/`, `api/`)
3. Test file names and assertions
4. Import statements in example files

Flag anything you've inferred so the user can verify.
