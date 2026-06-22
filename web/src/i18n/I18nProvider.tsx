import { useEffect, useState, createContext, useContext } from "react";
import i18n from "./config";
import type { Config } from "../types/config";

interface I18nContextValue {
  locale: string;
  rtl: boolean;
}

const I18nContext = createContext<I18nContextValue>({ locale: "en", rtl: false });

export function useI18nContext() {
  return useContext(I18nContext);
}

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(false);
  const [ctx, setCtx] = useState<I18nContextValue>({ locale: "en", rtl: false });

  useEffect(() => {
    let cancelled = false;

    async function init() {
      try {
        const res = await fetch("/api/config");
        const config: Config = await res.json();
        const locale = config.ui?.default_locale || "en";
        const rtl = config.ui?.rtl || false;

        const bundleRes = await fetch(`/locales/${locale}.json`);
        const bundle = bundleRes.ok ? await bundleRes.json() : null;

        const enRes = await fetch("/locales/en.json");
        const enBundle = enRes.ok ? await enRes.json() : null;

        if (!cancelled) {
          if (bundle) {
            i18n.addResourceBundle(locale, "translation", bundle, true, true);
          }
          if (enBundle) {
            i18n.addResourceBundle("en", "translation", enBundle, true, true);
          }

          await i18n.changeLanguage(locale);
          document.documentElement.dir = rtl ? "rtl" : "ltr";
          document.documentElement.lang = locale;

          setCtx({ locale, rtl });
          setReady(true);
        }
      } catch {
        document.documentElement.dir = "ltr";
        document.documentElement.lang = "en";

        try {
          const enRes = await fetch("/locales/en.json");
          const enBundle = enRes.ok ? await enRes.json() : null;
          if (enBundle) {
            i18n.addResourceBundle("en", "translation", enBundle, true, true);
          }
        } catch {
          // No bundles available
        }

        if (!cancelled) {
          setReady(true);
        }
      }
    }

    init();
    return () => { cancelled = true; };
  }, []);

  if (!ready) {
    return (
      <div className="min-h-screen bg-[#0f0f0f] flex items-center justify-center">
        <div className="inline-block w-8 h-8 border-2 border-[#3b82f6]/30 border-t-[#3b82f6] rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <I18nContext.Provider value={ctx}>
      {children}
    </I18nContext.Provider>
  );
}
