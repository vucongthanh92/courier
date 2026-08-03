import type { UserProfile } from "../types";

const USER_PROFILE_CACHE_KEY = "conversa.user_profiles";

export function readUserProfileCache() {
  const raw = localStorage.getItem(USER_PROFILE_CACHE_KEY);
  if (!raw) return new Map<string, UserProfile>();
  try {
    const entries = JSON.parse(raw) as Array<[string, UserProfile]>;
    return new Map(entries);
  } catch {
    localStorage.removeItem(USER_PROFILE_CACHE_KEY);
    return new Map<string, UserProfile>();
  }
}

export function writeUserProfileCache(profiles: Map<string, UserProfile>) {
  localStorage.setItem(USER_PROFILE_CACHE_KEY, JSON.stringify([...profiles.entries()]));
}

export function cacheUserProfiles(profiles: UserProfile[]) {
  if (profiles.length === 0) return readUserProfileCache();
  const current = readUserProfileCache();
  for (const profile of profiles) {
    if (profile.user_id) current.set(profile.user_id, profile);
  }
  writeUserProfileCache(current);
  return current;
}
