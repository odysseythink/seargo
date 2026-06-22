import i18n from "i18next";
import { initReactI18next } from "react-i18next";

export function initI18n(locale: string) {
  return i18n
    .use(initReactI18next)
    .init({
      lng: locale,
      fallbackLng: "en",
      interpolation: {
        escapeValue: false,
      },
      resources: {},
      returnObjects: false,
    });
}

export function changeLocale(locale: string, rtl: boolean) {
  document.documentElement.dir = rtl ? "rtl" : "ltr";
  document.documentElement.lang = locale;
  return i18n.changeLanguage(locale);
}

export default i18n;
