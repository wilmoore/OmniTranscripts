# 002. Examples Directory for Usage Patterns

Date: 2026-01-22

## Status

Accepted

## Context

The README was growing with usage examples but lacked depth for real-world scenarios. Users needed to see:
- Short-form video transcription (YouTube Shorts, TikTok, Instagram Reels)
- Docker-based workflows with actual transcription commands
- Async job polling patterns
- Local file handling

Adding all these examples to the README would make it too long and hard to navigate. The README should be scannable, not exhaustive.

## Decision

Create an `examples/` directory at the repository root with:
- Subdirectories organized by use case (short-form, local-files, docker, production)
- Executable shell scripts with parameterized inputs
- README files in each subdirectory explaining the examples
- A top-level README.md linking to all example categories

The main README links to `examples/` for users who want depth, keeping the README itself concise.

## Consequences

**Positive:**
- README stays scannable (~450 lines)
- Examples are discoverable via directory browsing
- Shell scripts are copy-paste ready and testable
- Each example category can grow independently
- Progressive disclosure: README for overview, examples/ for depth

**Negative:**
- Two places to maintain (README examples and examples/ directory)
- Users must navigate to a separate directory for full examples

## Alternatives Considered

1. **All examples in README**: Rejected - makes README too long, hurts scannability
2. **Wiki-based examples**: Rejected - not version-controlled with code
3. **External docs site**: Rejected - overkill for current project size

## Related

- Planning: `.plan/.done/feat-readme-local-files-positioning/`
