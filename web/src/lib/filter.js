export const SORT_FIELDS = [
  { value: "stars", group: "catalog", label: "Stars" },
  { value: "forks", group: "catalog", label: "Forks" },
  { value: "updated", group: "catalog", label: "Updated" },
  { value: "name", group: "catalog", label: "Name" },
  { value: "traffic", group: "traffic", label: "Traffic" },
  { value: "views", group: "traffic", label: "Views" },
  { value: "view_uniques", group: "traffic", label: "Unique visitors" },
  { value: "clones", group: "traffic", label: "Clones" },
  { value: "clone_uniques", group: "traffic", label: "Unique cloners" },
];

export function sortFieldLabel(value) {
  return SORT_FIELDS.find((field) => field.value === value)?.label || "Stars";
}

function metric(repo, key) {
  const traffic = repo.traffic || {};
  switch (key) {
    case "forks":
      return repo.forks_count || 0;
    case "updated":
      return new Date(repo.updated_at).getTime() || 0;
    case "name":
      return (repo.name || "").toLowerCase();
    case "traffic":
    case "views":
      return traffic.views || 0;
    case "view_uniques":
      return traffic.view_uniques || 0;
    case "clones":
      return traffic.clones || 0;
    case "clone_uniques":
      return traffic.clone_uniques || 0;
    default:
      return repo.stargazers_count || 0;
  }
}

function compareRepos(a, b, key, multiplier) {
  if (key === "name") {
    return a.name.localeCompare(b.name) * multiplier;
  }
  return (metric(a, key) - metric(b, key)) * multiplier;
}

function direction(order) {
  return order === "asc" ? 1 : -1;
}

export function applyFilters(repos, filters) {
  const language = (filters.language || "").trim();
  const minStars = Number.isNaN(parseInt(filters.minStars, 10)) ? 0 : parseInt(filters.minStars, 10);
  const query = (filters.query || "").toLowerCase();
  const includeForks = !!filters.includeForks;
  const hasHomepage = !!filters.hasHomepage;
  const visibility = filters.visibility || "public";
  const sortBy = filters.sortBy || "stars";
  const sortOrder = filters.sortOrder || "desc";
  const sortThen = (filters.sortThen || "").trim();
  const sortThenOrder = filters.sortThenOrder || "desc";

  let out = (repos || []).filter((repo) => {
    const repoLanguage = repo.language || "Not specified";
    if ((repo.stargazers_count || 0) < minStars) return false;
    if (language && repoLanguage !== language) return false;
    const text = `${repo.name || ""} ${repo.description || ""}`.toLowerCase();
    if (query && !text.includes(query)) return false;
    if (!includeForks && repo.fork) return false;
    if (visibility === "public" && repo.private) return false;
    if (visibility === "private" && !repo.private) return false;
    const homepage = (repo.homepage || "").trim();
    if (hasHomepage && !homepage) return false;
    return true;
  });

  const primaryDir = direction(sortOrder);
  const thenDir = direction(sortThenOrder);
  out = out.slice().sort((a, b) => {
    const primary = compareRepos(a, b, sortBy, primaryDir);
    if (primary !== 0 || !sortThen || sortThen === sortBy) return primary;
    return compareRepos(a, b, sortThen, thenDir);
  });
  return out;
}

export function languageOptions(repos) {
  const set = new Set();
  (repos || []).forEach((repo) => set.add(repo.language || "Not specified"));
  return Array.from(set).sort((a, b) => a.localeCompare(b));
}
