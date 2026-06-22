import type { PreferencesResponse, PreferencesUpdate } from "../types/search";

export async function fetchPreferences(): Promise<PreferencesResponse> {
  const res = await fetch("/api/preferences");
  if (!res.ok) throw new Error(`Failed to fetch preferences: ${res.status}`);
  return res.json();
}

export async function updatePreferences(update: PreferencesUpdate): Promise<PreferencesResponse> {
  const res = await fetch("/api/preferences", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(update),
  });
  if (!res.ok) throw new Error(`Failed to update preferences: ${res.status}`);
  return res.json();
}
