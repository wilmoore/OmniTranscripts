---
description: "Create & check out a new feature branch, entering plan mode to define its implementation"
allowed-tools: ["Bash"]
---

## Sync main with remote

!git fetch origin main && git checkout main && git pull origin main

## Context

Let's plan the implementation for: $ARGUMENTS

## Your Task

1. Enter **plan mode** (announce this to the user).
2. Confirm and document the requirements and scope.
3. Ask clarifying questions until mutual clarity is reached on the design and approach.
4. Generate a clear, descriptive feature branch name based on the agreed work.
5. Create and switch to the new branch.
6. **Planning Documentation**:
   - Create a directory under `docs/.plan` named after the branch (e.g., `fix/pattern-matcher-tests-static-rule` → `docs/.plan/fix-pattern-matcher-tests-static-rule`).
   - Store all planning notes, todos, and related documentation here.
7. Outline detailed implementation steps.
8. Implement the feature and document changes.
9. **Definition of Done**:
   - All features meet the agreed specification.
   - Verified by both user and assistant.
   - No errors, bugs, or warnings.
10. Declare success or commit changes only after full verification.
