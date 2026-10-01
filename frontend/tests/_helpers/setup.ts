// tests/frontend/_setup.ts
// Vitest global setup. Runs once before every test file.
// Per-test mocks live in _helpers.ts and are installed in beforeEach.

import { afterEach } from "vitest";
import { uninstallWailsMocks } from "./_helpers";

afterEach(() => {
  uninstallWailsMocks();
});
