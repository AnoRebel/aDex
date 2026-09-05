# audio-cue-system

- OS:           linux 6.19.13-zen1-1-zen
- Theme:        n/a
- Wails:        v3.0.0-beta.16
- Date:         2026-05-05
- App SHA:      e02dbba (branch migrate-wails-v3)
- Verifier:     automated — frontend/tests/composables/use-adex-audio.spec.ts

## What this proves

Validates the V2 audio cue system end-to-end:
- All 13 canonical eDEX cues (`alarm, denied, error, expand, folder,
  granted, info, keyboard, panels, scan, stdin, stdout, theme`) are
  registered in `frontend/app/assets/audio/soundpacks.json` under the
  `edex` pack
- All 13 WAVs physically copied into `frontend/app/assets/audio/edex/`
- Playback is performed by the GO BACKEND, not Howler. The webview refuses
  to start audio before a user gesture, which silently dropped every
  boot-splash cue; Go plays through the OS audio stack where no such policy
  applies. Howler remains only as a webview fallback.
- Web-Audio synthesizer fallback kicks in on Howler load failure or when
  the user opts into the `synth` pack
- Legacy event names (`button_click`, `notification`, `system_alert`, ...)
  alias to canonical cues so existing call sites continue to work
- Global mute, per-cue volume, and 4 Hz rate-limit on stdin/stdout all
  covered
- Background-mute signal (set by the backend audio service when another
  app takes audio focus) suppresses playback

## How to reproduce

```bash
bun scripts/copy-audio-cues.mjs
cd frontend
bunx vitest run tests/composables/use-adex-audio.spec.ts
```

Expected: `Tests 9 passed (9)`.

## Observed

- 9/9 tests pass (~140 ms in jsdom)
- WAVs total ~110 KB across 13 files

## Production capture

Run `task evidence:capture -- audio-cue-system <theme>` while clicking
through the on-screen keyboard / opening modals / switching themes;
attach the resulting screenshot here. Audio itself is captured by
recording the dev console (Howler logs each play call when DEBUG is set):

```bash
DEBUG=howler:* bun run dev | tee docs/evidence/audio-cue-system/play-log.txt
```


## Update — 2026-09-05

Re-verified on Wails v3.0.0-beta.16. See `verification.log` in this directory
for the generated artifact: 23 cue WAVs embedded in the Go binary, the Go
audio service tests, and the frontend cue tests.

The cue vocabulary has grown beyond the original 13: aDex adds `keypress`,
`destructive`, `gunshot` and `click`, and `button_click` now resolves to the
new click tone rather than `keyboard` (which is reserved for the mechanical
typewriter click).
