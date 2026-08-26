// composables/useKeyboardPulse.ts
//
// Tiny shared state that lets the terminal flash the on-screen keyboard
// in response to keystrokes it captured.
//
// Problem this solves:
//   xterm.js attaches its own document/textarea-level keydown listener
//   and calls preventDefault on the events it consumes. AdexKeyboard's
//   own `document.addEventListener('keydown', ...)` handler therefore
//   never sees keystrokes that landed on the terminal — only those
//   landing on the rest of the page (file manager, settings, etc.).
//
// Solution:
//   AdexTerminal hooks into `term.onData` (which fires AFTER xterm
//   processed the keystroke) and calls `pulseKey(data)` from this
//   module. AdexKeyboard watches `pulseId` and briefly adds the key
//   to its `physicalPressedKeys` set so the visual highlight fires.
//
//   Why a separate file (not events / globals):
//     - Module-scope Vue refs are reactive across imports automatically.
//     - No event-bus boilerplate, no addEventListener cleanup.
//     - Multiple terminal tabs can all push pulses without coupling.

import { ref } from "vue";

/**
 * Last key data the terminal observed. Set via pulseKey(data) from
 * AdexTerminal's term.onData callback. AdexKeyboard watches this and
 * derives a transient "pressed" highlight from it.
 *
 * `pulseId` increments on every pulse so the watcher fires even when
 * the same key is pressed twice in a row (otherwise Vue's value-equality
 * check would skip the second update for identical `data`).
 */
export const pulseData = ref<string>("");
export const pulseId = ref<number>(0);

/**
 * Map an xterm onData payload to the human-readable key name the
 * keyboard layout uses. xterm fires onData with the raw bytes the
 * shell would receive, which for printable chars is just the char
 * but for special keys is an escape sequence we have to decode.
 *
 * We only care about visual highlighting, not perfect mapping — if
 * we can't decode (esc sequence we don't handle), we return the raw
 * data and the keyboard's normalizeKeyName won't match anything,
 * which is a harmless miss.
 */
export function decodeKeyName(data: string): string {
  // Printable single char — return as-is, lowercased for matching.
  if (data.length === 1) {
    const c = data.charCodeAt(0);
    // Control char range — fall through to special handling below.
    if (c >= 32 && c < 127) return data.toLowerCase();
  }
  // Common xterm escape sequences → key names.
  // Reference: https://en.wikipedia.org/wiki/ANSI_escape_code
  switch (data) {
    case "\r":      return "enter";
    case "\n":      return "enter";
    case "\t":      return "tab";
    case "\x7f":    return "backspace";
    case "\x1b":    return "escape";
    case "\x1b[A":  return "arrowup";
    case "\x1b[B":  return "arrowdown";
    case "\x1b[C":  return "arrowright";
    case "\x1b[D":  return "arrowleft";
    case "\x1b[H":  return "home";
    case "\x1b[F":  return "end";
    case "\x1b[5~": return "pageup";
    case "\x1b[6~": return "pagedown";
    case "\x1b[3~": return "delete";
    case "\x1b[2~": return "insert";
    case " ":       return " ";
    default:        return data; // Unknown / paste / multi-byte → no match
  }
}

export function pulseKey(data: string): void {
  pulseData.value = decodeKeyName(data);
  pulseId.value++;
}
