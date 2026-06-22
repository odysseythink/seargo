import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import type { PluginPrefItem, PreferencesResponse, UserPreferences } from "../types/search";
import { fetchPreferences, updatePreferences } from "../api/client";

function groupBySection(plugins: PluginPrefItem[]): Record<string, PluginPrefItem[]> {
  const groups: Record<string, PluginPrefItem[]> = {};
  for (const p of plugins) {
    const section = p.preference_section || "general";
    if (!groups[section]) groups[section] = [];
    groups[section].push(p);
  }
  return groups;
}

function useDebouncedSave(
  deps: unknown[],
  buildPayload: () => Record<string, unknown>,
  delay: number = 500
) {
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    // Skip initial render
    if (timerRef.current === null) {
      timerRef.current = 0 as unknown as ReturnType<typeof setTimeout>;
      return;
    }

    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(async () => {
      setSaving(true);
      try {
        await updatePreferences(buildPayload());
      } catch {
        // silent fail
      } finally {
        setSaving(false);
      }
    }, delay);

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return saving;
}

function ToggleField({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <label className="flex items-center gap-3 p-3 border border-[rgba(255,255,255,0.08)] rounded-lg cursor-pointer hover:bg-[#2a2a2a] transition-colors">
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="w-4 h-4 accent-[#3b82f6]"
      />
      <span className="text-sm text-[#e5e5e5]">{label}</span>
    </label>
  );
}

function SelectField({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <label className="flex flex-col gap-1">
      <span className="text-sm text-[#e5e5e5]">{label}</span>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-64 px-3 py-2 bg-[#1a1a1a] border border-[rgba(255,255,255,0.08)]
                   rounded-lg text-[#e5e5e5] outline-none focus:border-[#3b82f6]"
      >
        {options.map((opt) => (
          <option key={opt.value} value={opt.value}>{opt.label}</option>
        ))}
      </select>
    </label>
  );
}

export default function SettingsPage() {
  const { t } = useTranslation();
  const [prefs, setPrefs] = useState<PreferencesResponse | null>(null);
  const [settings, setSettings] = useState<UserPreferences | null>(null);
  const [pluginActive, setPluginActive] = useState<Record<string, boolean>>({});
  const [answererActive, setAnswererActive] = useState<Record<string, boolean>>({});

  useEffect(() => {
    fetchPreferences().then((data) => {
      setPrefs(data);
      setSettings(data.settings);
      const pa: Record<string, boolean> = {};
      for (const p of data.plugins) pa[p.id] = p.active;
      setPluginActive(pa);
      const aa: Record<string, boolean> = {};
      for (const a of data.answerers) aa[a.id] = a.active;
      setAnswererActive(aa);
    });
  }, []);

  const saving = useDebouncedSave(
    [settings, pluginActive, answererActive],
    () => ({
      ...settings,
      plugins: pluginActive,
      answerers: answererActive,
    }),
    500
  );

  const updateSetting = <K extends keyof UserPreferences>(key: K, value: UserPreferences[K]) => {
    if (!settings) return;
    setSettings({ ...settings, [key]: value });
  };

  if (!prefs || !settings) {
    return <div className="p-4 text-[#e5e5e5]">{t("preferences.loading")}</div>;
  }

  const sections: Record<string, string> = {
    general: t("preferences.section_general"),
    ui: t("preferences.section_ui"),
    privacy: t("preferences.section_privacy"),
    query: t("preferences.section_query"),
  };

  const grouped = groupBySection(prefs.plugins);

  return (
    <div className="max-w-2xl mx-auto p-6 pt-20">
      <h1 className="text-2xl font-bold mb-6 text-[#e5e5e5]">{t("preferences.title")}</h1>

      {/* Search Section */}
      <section className="mb-6">
        <h2 className="text-lg font-semibold mb-3 text-[#9ca3af]">{t("preferences.search_section")}</h2>
        <div className="space-y-3">
          <SelectField
            label={t("preferences.autocomplete_label")}
            value={settings.autocomplete}
            options={[
              { value: "google", label: "Google" },
              { value: "bing", label: "Bing" },
              { value: "duckduckgo", label: "DuckDuckGo" },
              { value: "brave", label: "Brave" },
              { value: "qwant", label: "Qwant" },
              { value: "startpage", label: "Startpage" },
              { value: "wikipedia", label: "Wikipedia" },
              { value: "dbpedia", label: "DBpedia" },
              { value: "swisscows", label: "Swisscows" },
              { value: "baidu", label: "Baidu" },
              { value: "360search", label: "360 Search" },
              { value: "naver", label: "Naver" },
              { value: "yandex", label: "Yandex" },
              { value: "seznam", label: "Seznam" },
              { value: "sogou", label: "Sogou" },
              { value: "mwmbl", label: "Mwmbl" },
              { value: "privacywall", label: "PrivacyWall" },
              { value: "quark", label: "Quark" },
            ]}
            onChange={(v) => updateSetting("autocomplete", v)}
          />

          <SelectField
            label={t("preferences.language_label")}
            value={settings.language}
            options={[
              { value: "en", label: "English" },
              { value: "zh", label: "中文" },
              { value: "de", label: "Deutsch" },
              { value: "fr", label: "Français" },
              { value: "es", label: "Español" },
              { value: "ja", label: "日本語" },
              { value: "ko", label: "한국어" },
            ]}
            onChange={(v) => updateSetting("language", v)}
          />

          <SelectField
            label={t("preferences.locale_label")}
            value={settings.locale}
            options={prefs.locales.map((l) => ({ value: l.tag, label: `${l.name} (${l.tag})` }))}
            onChange={(v) => updateSetting("locale", v)}
          />

          <SelectField
            label={t("preferences.theme_label")}
            value={settings.theme}
            options={prefs.themes.map((t) => ({ value: t, label: t.charAt(0).toUpperCase() + t.slice(1) }))}
            onChange={(v) => updateSetting("theme", v)}
          />

          <SelectField
            label={t("preferences.safesearch_label")}
            value={String(settings.safesearch)}
            options={[
              { value: "0", label: t("preferences.safesearch_off") },
              { value: "1", label: t("preferences.safesearch_moderate") },
              { value: "2", label: t("preferences.safesearch_strict") },
            ]}
            onChange={(v) => updateSetting("safesearch", Number(v))}
          />

          <SelectField
            label={t("preferences.method_label")}
            value={settings.method}
            options={[
              { value: "GET", label: "GET" },
              { value: "POST", label: "POST" },
            ]}
            onChange={(v) => updateSetting("method", v)}
          />
        </div>
      </section>

      {/* Display Section */}
      <section className="mb-6">
        <h2 className="text-lg font-semibold mb-3 text-[#9ca3af]">{t("preferences.display_section")}</h2>
        <div className="space-y-2">
          <ToggleField
            label={t("preferences.results_on_new_tab_label")}
            checked={settings.results_on_new_tab}
            onChange={(v) => updateSetting("results_on_new_tab", v)}
          />
          <ToggleField
            label={t("preferences.center_alignment_label")}
            checked={settings.center_alignment}
            onChange={(v) => updateSetting("center_alignment", v)}
          />
          <ToggleField
            label={t("preferences.query_in_title_label")}
            checked={settings.query_in_title}
            onChange={(v) => updateSetting("query_in_title", v)}
          />
          <ToggleField
            label={t("preferences.search_on_category_select_label")}
            checked={settings.search_on_category_select}
            onChange={(v) => updateSetting("search_on_category_select", v)}
          />
          <ToggleField
            label={t("preferences.image_proxy_label")}
            checked={settings.image_proxy}
            onChange={(v) => updateSetting("image_proxy", v)}
          />
        </div>
      </section>

      {/* Saving indicator */}
      {saving && (
        <div className="text-sm text-gray-500 mb-2">{t("preferences.saving")}</div>
      )}

      {/* Plugin Sections */}
      {Object.entries(sections).map(([key, label]) => {
        const plugins = grouped[key];
        if (!plugins || plugins.length === 0) return null;
        return (
          <section key={key} className="mb-6">
            <h2 className="text-lg font-semibold mb-2 text-[#9ca3af]">{label}</h2>
            <div className="space-y-2">
              {plugins.map((p) => (
                <label
                  key={p.id}
                  className="flex items-center gap-3 p-3 border border-[rgba(255,255,255,0.08)] rounded-lg hover:bg-[#2a2a2a] cursor-pointer transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={pluginActive[p.id] ?? false}
                    onChange={() => {
                      const next = { ...pluginActive, [p.id]: !pluginActive[p.id] };
                      setPluginActive(next);
                    }}
                    className="w-4 h-4 accent-[#3b82f6]"
                  />
                  <div>
                    <div className="font-medium text-[#e5e5e5]">{p.name}</div>
                    <div className="text-sm text-[#9ca3af]">{p.description}</div>
                    {p.examples && p.examples.length > 0 && (
                      <div className="text-xs text-gray-500 mt-1">
                        {t("preferences.examples")}: {p.examples.join(", ")}
                      </div>
                    )}
                  </div>
                </label>
              ))}
            </div>
          </section>
        );
      })}

      {/* Answerers Section */}
      {prefs.answerers.length > 0 && (
        <section className="mb-6">
          <h2 className="text-lg font-semibold mb-2 text-[#9ca3af]">{t("preferences.instant_answers")}</h2>
          <div className="space-y-2">
            {prefs.answerers.map((a) => (
              <label
                key={a.id}
                className="flex items-center gap-3 p-3 border border-[rgba(255,255,255,0.08)] rounded-lg hover:bg-[#2a2a2a] cursor-pointer transition-colors"
              >
                <input
                  type="checkbox"
                  checked={answererActive[a.id] ?? false}
                  onChange={() => {
                    const next = { ...answererActive, [a.id]: !answererActive[a.id] };
                    setAnswererActive(next);
                  }}
                  className="w-4 h-4 accent-[#3b82f6]"
                />
                <div>
                  <div className="font-medium text-[#e5e5e5]">{a.name}</div>
                  <div className="text-sm text-[#9ca3af]">{a.description}</div>
                  <div className="text-xs text-gray-500 mt-1">
                    {t("preferences.keywords")}: {a.keywords.join(", ")}
                  </div>
                </div>
              </label>
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
