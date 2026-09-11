export function trafficRepos(repos) {
  return (repos || []).filter((repo) => repo.traffic?.available);
}

function withShare(rows, valueOf) {
  const peak = Math.max(1, ...rows.map(valueOf));
  return rows.map((row) => ({ ...row, share: Math.round((valueOf(row) / peak) * 100) }));
}

export function rankTraffic(repos, key, limit = 10) {
  const rows = trafficRepos(repos)
    .map((repo) => ({
      name: repo.name,
      href: repo.html_url || "",
      language: repo.language || "Not specified",
      stars: repo.stargazers_count || 0,
      value: Number(repo.traffic?.[key]) || 0,
    }))
    .filter((row) => row.value > 0)
    .sort((a, b) => b.value - a.value || a.name.localeCompare(b.name))
    .slice(0, limit);
  return withShare(rows, (row) => row.value);
}

function sumField(repos, key) {
  return trafficRepos(repos).reduce((total, repo) => total + (Number(repo.traffic?.[key]) || 0), 0);
}

export function catalogTrafficTotals(repos) {
  return {
    tracked: trafficRepos(repos).length,
    loaded: (repos || []).length,
    views: sumField(repos, "views"),
    viewUniques: sumField(repos, "view_uniques"),
    clones: sumField(repos, "clones"),
    cloneUniques: sumField(repos, "clone_uniques"),
  };
}

function emptyDay(date) {
  return { date, views: 0, viewUniques: 0, clones: 0, cloneUniques: 0 };
}

function addDayPoints(byDay, points, countKey, uniqueKey) {
  for (const point of points || []) {
    const date = (point.timestamp || "").slice(0, 10);
    if (!date) continue;
    const row = byDay.get(date) || emptyDay(date);
    row[countKey] += point.count || 0;
    row[uniqueKey] += point.uniques || 0;
    byDay.set(date, row);
  }
}

function fillDaySpan(rows) {
  if (rows.length < 2) return rows;
  const byDate = new Map(rows.map((row) => [row.date, row]));
  const start = Date.parse(`${rows[0].date}T00:00:00Z`);
  const end = Date.parse(`${rows[rows.length - 1].date}T00:00:00Z`);
  const out = [];
  for (let ts = start; ts <= end; ts += 864e5) {
    const date = new Date(ts).toISOString().slice(0, 10);
    out.push(byDate.get(date) || emptyDay(date));
  }
  return out;
}

export function dailyCatalogPulse(repos) {
  const byDay = new Map();
  for (const repo of trafficRepos(repos)) {
    addDayPoints(byDay, repo.traffic.views_by_day, "views", "viewUniques");
    addDayPoints(byDay, repo.traffic.clones_by_day, "clones", "cloneUniques");
  }
  const rows = [...byDay.values()].sort((a, b) => a.date.localeCompare(b.date));
  return fillDaySpan(rows);
}

function pulseHeight(value, peak) {
  if (value <= 0) return "0%";
  return `${Math.max(6, Math.round((value / peak) * 100))}%`;
}

export function pulseBars(days) {
  const peak = Math.max(1, ...(days || []).flatMap((day) => [day.viewUniques, day.cloneUniques]));
  return (days || []).map((day) => {
    const utc = new Date(`${day.date}T00:00:00Z`);
    return {
      ...day,
      label: utc.toLocaleDateString("en-US", { month: "short", day: "numeric", timeZone: "UTC" }),
      viewHeight: pulseHeight(day.viewUniques, peak),
      cloneHeight: pulseHeight(day.cloneUniques, peak),
    };
  });
}

export function uniqueVisitorsByLanguage(repos, limit = 8) {
  const totals = new Map();
  for (const repo of trafficRepos(repos)) {
    const language = repo.language || "Not specified";
    totals.set(language, (totals.get(language) || 0) + (repo.traffic.view_uniques || 0));
  }
  const ranked = [...totals.entries()]
    .map(([language, value]) => ({ language, value }))
    .filter((row) => row.value > 0)
    .sort((a, b) => b.value - a.value);
  const head = ranked.slice(0, limit);
  const rest = ranked.slice(limit).reduce((total, row) => total + row.value, 0);
  if (rest > 0) head.push({ language: "Other", value: rest });
  return withShare(head, (row) => row.value);
}

export function quietStarsRank(repos, limit = 10) {
  const rows = trafficRepos(repos)
    .map((repo) => {
      const uniques = repo.traffic.view_uniques || 0;
      const stars = repo.stargazers_count || 0;
      return { name: repo.name, href: repo.html_url || "", uniques, stars, score: uniques / Math.max(stars, 1) };
    })
    .filter((row) => row.uniques > 0)
    .sort((a, b) => b.score - a.score || b.uniques - a.uniques)
    .slice(0, limit);
  return withShare(rows, (row) => row.score);
}
