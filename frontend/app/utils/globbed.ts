/**
 * Look up a module in an `import.meta.glob` map by its trailing path.
 *
 * The keys these maps produce depend on how the `~` alias resolves, and that
 * differs between builds:
 *
 *   Nuxt build   ~ -> app/    keys look like "/assets/data/themes/tron.json"
 *   Vitest       ~ -> .       keys look like "/app/assets/data/themes/tron.json"
 *
 * Matching on a fixed prefix therefore worked in the application and failed
 * in every test that loaded a theme or a keyboard layout. Matching on the
 * suffix is stable under both, since the part after the alias is identical.
 *
 * @param modules  the glob map
 * @param file     path relative to the globbed data directory,
 *                 e.g. "themes/tron.json"
 */
export function resolveGlobbed<T>(
  modules: Record<string, () => Promise<T>>,
  file: string,
): (() => Promise<T>) | undefined {
  const suffix = `/${file}`.replace(/\/+/g, "/");
  const key = Object.keys(modules).find((k) => k.endsWith(suffix));
  return key ? modules[key] : undefined;
}
