# Git Workflow

## Branches

- Create a feature branch for every change: `feature/<task-name-kebab-case>`.
- Never commit directly to `main`.

## What not to commit

Never commit build artifacts (`dist/`, `bin/`), binaries, temporary or
backup files (e.g. `*.bak`), or `.claude/`. Don't force-add anything that
`.gitignore` excludes.

## Commit messages

- Subject: at most 50 characters, capitalized, imperative mood
  ("Add", "Fix", "Update"), no trailing period.
- Leave a blank line between the subject and the body.
- Wrap the body at 72 characters.
