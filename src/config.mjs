import { resolve, dirname } from "path";

export const repoUrl = "https://github.com/take4s5i/blog";

export const getPathInRepo = (metaUrl) => {
  const prefix = `file://${getRepoRoot()}`;
  if (!metaUrl.startsWith(prefix)) {
    throw new Error(`bad metaUrl: ${metaUrl}`);
  }

  return metaUrl.substring(prefix.length);
};

export const getRepoRoot = () => {
  const url = import.meta.url;

  const prefix = "file://";
  if (!url.startsWith(prefix)) {
    throw new Error(`bad metaUrl: ${url}`);
  }

  const suffix = "/src/config.mjs";
  if (!url.endsWith(suffix)) {
    return "";
    // throw new Error(`bad metaUrl: ${url}`);
  }

  return url.substring(0, url.length - suffix.length).substring(prefix.length);
};

export const getBlobUrl = (path) => {
  return `${repoUrl}/blob/main/${path}`;
};

export const getDirUrlFromMeta = (metaUrl) => {
  const path = getPathInRepo(metaUrl);
  return `${repoUrl}/blob/main/${dirname(path)}`;
};

export const getUrl = (metaUrl) => {
  const path = getPathInRepo(metaUrl);
  return `${repoUrl}/blob/main/${path}`;
};
