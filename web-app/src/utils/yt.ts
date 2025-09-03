export function extractVideoId(input: string): string {
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
