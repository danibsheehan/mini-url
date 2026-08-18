---
name: doc-writer
description: >
  Generates high-quality documentation for JavaScript/TypeScript and Go codebases,
  and checks existing docs for drift against the code. Use this skill whenever a
  user asks to write, generate, update, improve, or create documentation of any
  kind — including README files, API docs (JSDoc/GoDoc), inline code comments, or
  function-level docstrings. Trigger even for vague requests like "document this",
  "add docs to my code", "write a README", "explain this function", or "make this
  repo easier to understand". Also trigger for "check if the docs are up to date",
  "is the README accurate", "audit the documentation", or before a release/PR that
  touches routes, config, or project layout. When in doubt, use this skill.
---

# Doc Writer Skill

Generates clear, consistent, production-quality documentation for JS/TS and Go projects.
Covers three output types: **README files**, **API/function docs**, and **inline code comments**.
All output is written in Markdown (`.md`) unless writing inline source annotations.

---

## Step 1: Classify the Request

Determine which doc type(s) are needed:

| Request | Doc Type |
|---|---|
| "Write a README", "document this repo" | → README |
| "Document this function/class/interface", "add JSDoc/GoDoc" | → API Docs |
| "Add comments", "explain what this code does inline" | → Inline Comments |
| "Is the README accurate", "check docs for drift", "audit documentation" | → Drift Check |
| Mixed / ambiguous | → Ask, or default to README + API Docs |

---

## Step 2: Gather Context

Before writing, read the relevant files:

- **README**: Scan repo structure, `package.json` / `go.mod`, existing README if any, entry points, exported symbols
- **API Docs**: Read the specific file(s) containing the functions/types to document
- **Inline Comments**: Read the specific functions or blocks to annotate

Use `bash_tool` to explore if needed:
```bash
# JS/TS: find exported functions/types
grep -rn "^export " src/ --include="*.ts" | head -40

# Go: find exported symbols
grep -rn "^func \|^type \|^var \|^const " *.go | grep -v "_test.go" | head -40

# Repo overview
find . -maxdepth 2 -name "*.md" -o -name "package.json" -o -name "go.mod" | head -20
```

---

## Step 3: Write the Documentation

Read the appropriate reference file for the doc type before writing:

- **README** → read `references/readme.md`
- **API Docs (JS/TS)** → read `references/jsdoc.md`
- **API Docs (Go)** → read `references/godoc.md`
- **Inline Comments** → read `references/inline-comments.md`

Then produce the output following those guidelines exactly.

---

## Mode: Drift Check

Triggered separately from Steps 1–4. Compares what the README *claims* against
what the code *does*, and reports gaps — it does not silently rewrite anything.

1. Extract from the README: documented routes/endpoints, the config table
   (flags, env vars, defaults, ports, file paths), and any project-layout /
   file-tree description.
2. Extract the same facts from the code:
   - Go: route registration (`http.HandleFunc`, router `.Handle`/`.Get` calls),
     the literal defaults passed into config/init functions (e.g. a hardcoded
     port or DB path string), and the actual package directory structure.
   - JS/TS: exported route handlers / API definitions, `package.json` scripts,
     config object defaults, and the actual `src/` directory structure.
3. Diff the two lists and classify each mismatch:
   - **Missing from docs** — exists in code, not mentioned in README.
   - **Stale in docs** — documented but no longer exists, or the documented
     value (port, path, status code, default) no longer matches the code.
   - **Drifted layout** — README's project-layout section lists directories/
     files that don't match what's actually on disk.
4. Report findings as a plain list — file/line references for both the doc
   claim and the code fact — with nothing fixed yet.
5. If the user confirms they want it fixed, proceed into Step 3's README flow
   for the affected sections only (don't regenerate the whole file).

---

## Step 4: Deliver Output

Write directly into the repo using the Edit or Write tool — do not stage output elsewhere.

- **README**: Write/update `README.md` at the repo root.
- **API Docs**: Edit the source file(s) in place, inserting doc comments (JSDoc/GoDoc) directly above the documented symbol.
- **Inline Comments**: Edit the source file(s) in place, adding comments inline at the relevant lines.
- If multiple files are touched, edit each one.

Always tell the user:
1. What was generated or changed, with file paths
2. Any gaps (e.g., "I couldn't find a description for `X` — you may want to fill that in")
