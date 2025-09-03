export function extractYTVideoId(input: string): string {
  const u = input.trim();

  const mShort = u.match(/shorts\/([a-zA-Z0-9_-]{6,})/);
  if (mShort) {
    return mShort[1];
  }

  const mWatch = u.match(/[?&]v=([a-zA-Z0-9_-]{6,})/);
  if (mWatch) {
    return mWatch[1];
  }

  const mEmbed = u.match(/embed\/([a-zA-Z0-9_-]{6,})/);
  if (mEmbed) {
    return mEmbed[1];
  }

  return u;
}

export function extractTikTokVideoId(urlOrId: string): string {
  const v = urlOrId.trim();
  const m = v.match(/\/video\/(\d+)/) || v.match(/^(\d{8,})$/);
  return m ? m[1] : "";
}

export type VideoPlatform = "youtube" | "tiktok" | "unknown";

export function detectVideoPlatform(url: string): VideoPlatform {
  try {
    const parsed = new URL(url.trim());

    const host = parsed.hostname.toLowerCase();
    if (host.includes("youtube.com") || host === "youtu.be") {
      return "youtube";
    }

    if (host.includes("tiktok.com")) {
      return "tiktok";
    }

    return "unknown";
  } catch {
    return "unknown";
  }
}
