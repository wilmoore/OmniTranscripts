# Plan: Add Clipboard-Friendly Make Targets

## Status: COMPLETE

## Problem

`make transcribe URL='...'` produces verbose output; users want clean transcript on clipboard.

## Solution Implemented

### 1. Added cross-platform clipboard detection to Makefile

```makefile
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    CLIP_CMD = pbcopy
else
    # Linux: prefer xclip, fallback to xsel
    ...
endif
```

### 2. Added `transcribe-clip` target

Extracts transcript between `--- Transcript ---` and `--- Segments ---` markers, pipes to clipboard, confirms with "Transcript copied to clipboard".

### 3. Documented CLI pipeline patterns

Added to `docs/development.md`:
- `make transcribe-clip` usage
- Manual pipeline patterns for custom filtering
- Examples for transcript, segments, file output

## Files Modified

1. `makefile` - Added CLIP_CMD detection + transcribe-clip target
2. `docs/development.md` - Added clipboard and pipeline documentation
3. `CLAUDE.md` - Added transcribe-clip to command reference

## Verification

- [x] `make help` shows transcribe-clip target
- [ ] Actual transcription test (requires whisper model)
