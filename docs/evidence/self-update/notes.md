# self-update

- OS:           linux 7.2 (Arch, Hyprland, Wayland)
- Theme:        apollo
- Wails:        v3.1.0
- Date:         2026-11-03
- App SHA:      4081d95 plus the uncommitted updater, translation and AppImage changes
- Verifier:     manual — screenshots of the running AppImage (cropped to the dialog)

## What this proves

An AppImage built at 0.9.0 finds a signed 1.0.0 in a local manifest, downloads
it with live progress, verifies its signature, stages it, and relaunches as
1.0.0. An artifact altered after signing is refused and the AppImage on disk is
left untouched. A copy that is not running from an AppImage is told about the
release but only offered the release page. The dialog follows the interface
language (Japanese shown).

## How to reproduce

Follow "Trying an update locally" in `docs/cross-platform-packaging.md`
(0.9.0 and 1.0.0 AppImages, manifest signed with `adex.key`, served on
127.0.0.1:8491, app run with `ADEX_UPDATE_MANIFEST`). The server was throttled
to about 6 MB/s so the download stays on screen.

1. Launch the 0.9.0 AppImage: the dialog opens with INSTALL focused.
2. INSTALL: download progress, then UPDATE READY.
3. RESTART & APPLY: the app relaunches.
4. Tampered: restore the 0.9.0 AppImage, flip one byte of the served 1.0.0
   artifact (same size), INSTALL.
5. Notify-only: run the extracted 0.9.0 (`--appimage-extract`, then `AppRun`
   without `APPIMAGE`).

## Observed

- `available.png` — CURRENT 0.9.0, LATEST 1.0.0; INSTALL has focus.
- `downloading.png` — progress bar with "21.3 MB / 164.7 MB · 4.3 MB/s".
- Verifying and installing took under a second together, so no frame caught
  them; the dialog went straight from downloading to `ready.png`.
- After RESTART & APPLY the AppImage on disk matched the served 1.0.0 artifact
  (md5 748f1336…), and the relaunched process requested
  `update-manifest.json?…&version=1.0.0` and showed no update.
- `error.png` — VERIFY FAILED, "updater: digest mismatch"; the AppImage's md5
  was unchanged (c2e9ed5b…) and no staging directory was left behind.
- `notify-only.png` — "not running from an AppImage", OPEN RELEASE PAGE in
  place of INSTALL.
- `notify-only-ja.png` — the same dialog with the interface set to Japanese;
  the backend's reason text stays as sent.
- In these captures the release-notes box is clipped to one line. That was a
  layout bug (a zero flex basis), fixed after the captures were taken.
