import { useEffect, useState } from "react";
import type { PluginPrefItem, AnswererPrefItem, PreferencesResponse } from "../types/search";
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

export default function SettingsPage() {
  const [prefs, setPrefs] = useState<PreferencesResponse | null>(null);
  const [pluginActive, setPluginActive] = useState<Record<string, boolean>>({});
  const [answererActive, setAnswererActive] = useState<Record<string, boolean>>({});
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    fetchPreferences().then((data) => {
      setPrefs(data);
      const pa: Record<string, boolean> = {};
      for (const p of data.plugins) pa[p.id] = p.active;
      setPluginActive(pa);
      const aa: Record<string, boolean> = {};
      for (const a of data.answerers) aa[a.id] = a.active;
      setAnswererActive(aa);
    });
  }, []);

  const togglePlugin = async (id: string) => {
    const next = { ...pluginActive, [id]: !pluginActive[id] };
    setPluginActive(next);
    setSaving(true);
    try {
      await updatePreferences({ plugins: next, answerers: answererActive });
    } finally {
      setSaving(false);
    }
  };

  const toggleAnswerer = async (id: string) => {
    const next = { ...answererActive, [id]: !answererActive[id] };
    setAnswererActive(next);
    setSaving(true);
    try {
      await updatePreferences({ plugins: pluginActive, answerers: next });
    } finally {
      setSaving(false);
    }
  };

  if (!prefs) return <div className="p-4 text-[#e5e5e5]">Loading preferences...</div>;

  const sections: Record<string, string> = {
    general: "General",
    ui: "User Interface",
    privacy: "Privacy",
    query: "Query",
  };

  const grouped = groupBySection(prefs.plugins);

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6 text-[#e5e5e5]">Preferences</h1>

      {saving && <div className="text-sm text-gray-500 mb-2">Saving...</div>}

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
                  className="flex items-center gap-3 p-3 border border-gray-700 rounded hover:bg-[#2a2a2a] cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={pluginActive[p.id] ?? false}
                    onChange={() => togglePlugin(p.id)}
                    className="w-4 h-4"
                  />
                  <div>
                    <div className="font-medium text-[#e5e5e5]">{p.name}</div>
                    <div className="text-sm text-[#9ca3af]">{p.description}</div>
                    {p.examples && p.examples.length > 0 && (
                      <div className="text-xs text-gray-500 mt-1">
                        e.g. {p.examples.join(", ")}
                      </div>
                    )}
                  </div>
                </label>
              ))}
            </div>
          </section>
        );
      })}

      {prefs.answerers.length > 0 && (
        <section className="mb-6">
          <h2 className="text-lg font-semibold mb-2 text-[#9ca3af]">Instant Answers</h2>
          <div className="space-y-2">
            {prefs.answerers.map((a) => (
              <label
                key={a.id}
                className="flex items-center gap-3 p-3 border border-gray-700 rounded hover:bg-[#2a2a2a] cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={answererActive[a.id] ?? false}
                  onChange={() => toggleAnswerer(a.id)}
                  className="w-4 h-4"
                />
                <div>
                  <div className="font-medium text-[#e5e5e5]">{a.name}</div>
                  <div className="text-sm text-[#9ca3af]">{a.description}</div>
                  <div className="text-xs text-gray-500 mt-1">
                    Keywords: {a.keywords.join(", ")}
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
