// Stub for Nuxt's `#imports` virtual module so Vitest can resolve composables
// that rely on Nuxt auto-imports without booting Nuxt.

export const useState = <T>(_key: string, init: () => T) => {
  const value = { value: init() };
  return value;
};
export const useRuntimeConfig = () => ({ public: {} });
export const navigateTo = async () => {};
export const defineNuxtPlugin = (fn: unknown) => fn;
export const useNuxtApp = () => ({ $config: { public: {} } });
