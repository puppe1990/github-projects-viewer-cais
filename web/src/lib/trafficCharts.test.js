import { describe, expect, test } from "vitest";
import {
  catalogTrafficTotals,
  cutUnsettledDays,
  dailyCatalogPulse,
  pulseBars,
  quietStarsRank,
  rankTraffic,
  trafficRepos,
  uniqueVisitorsByLanguage,
} from "./trafficCharts.js";

const catalog = [
  {
    name: "alpha",
    language: "Go",
    stargazers_count: 100,
    traffic: {
      available: true,
      views: 50,
      view_uniques: 40,
      clones: 10,
      clone_uniques: 8,
      views_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 7, uniques: 3 }],
      clones_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 2, uniques: 1 }],
    },
  },
  {
    name: "beta",
    language: "Go",
    stargazers_count: 2,
    traffic: {
      available: true,
      views: 80,
      view_uniques: 70,
      clones: 5,
      clone_uniques: 4,
      views_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 10, uniques: 8 }],
      clones_by_day: [{ timestamp: "2026-08-28T00:00:00Z", count: 5, uniques: 4 }],
    },
  },
  {
    name: "gamma",
    language: "Python",
    stargazers_count: 10,
    traffic: {
      available: true,
      views: 20,
      view_uniques: 5,
      clones: 90,
      clone_uniques: 60,
    },
  },
  { name: "delta", language: "Rust", stargazers_count: 1 },
];

describe("trafficRepos", () => {
  test("keeps only repositories with a traffic snapshot", () => {
    expect(trafficRepos(catalog).map((repo) => repo.name)).toEqual(["alpha", "beta", "gamma"]);
  });
});

describe("rankTraffic", () => {
  test("ranks top unique visitors and unique cloners", () => {
    expect(rankTraffic(catalog, "view_uniques", 10).map((row) => row.name)).toEqual(["beta", "alpha", "gamma"]);
    expect(rankTraffic(catalog, "clone_uniques", 2).map((row) => [row.name, row.value])).toEqual([
      ["gamma", 60],
      ["alpha", 8],
    ]);
  });

  test("keeps the repository html url on ranked rows", () => {
    const withUrls = catalog.map((repo, i) => ({ ...repo, html_url: `https://github.com/octocat/${repo.name}` }));
    expect(rankTraffic(withUrls, "view_uniques", 1)[0].href).toBe("https://github.com/octocat/beta");
    expect(quietStarsRank(withUrls)[0].href).toBe("https://github.com/octocat/beta");
  });

  test("scales bar share against the peak value", () => {
    const [lead, second] = rankTraffic(catalog, "view_uniques", 10);
    expect(lead.share).toBe(100);
    expect(second.share).toBe(57);
  });
});

describe("catalogTrafficTotals", () => {
  test("sums GitHub windows across tracked repositories", () => {
    expect(catalogTrafficTotals(catalog)).toEqual({
      tracked: 3,
      loaded: 4,
      views: 150,
      viewUniques: 115,
      clones: 105,
      cloneUniques: 72,
    });
  });
});

describe("dailyCatalogPulse", () => {
  test("merges daily unique views and clones by date", () => {
    expect(dailyCatalogPulse(catalog)).toEqual([
      { date: "2026-08-27", views: 17, viewUniques: 11, clones: 2, cloneUniques: 1 },
      { date: "2026-08-28", views: 0, viewUniques: 0, clones: 5, cloneUniques: 4 },
    ]);
  });

  test("fills missing dates between the first and last day", () => {
    const sparse = [
      {
        name: "alpha",
        traffic: {
          available: true,
          views_by_day: [
            { timestamp: "2026-08-27T00:00:00Z", count: 1, uniques: 1 },
            { timestamp: "2026-08-29T00:00:00Z", count: 2, uniques: 2 },
          ],
        },
      },
    ];
    expect(dailyCatalogPulse(sparse).map((day) => day.date)).toEqual(["2026-08-27", "2026-08-28", "2026-08-29"]);
  });
});

describe("cutUnsettledDays", () => {
  const pulse = [
    { date: "2026-09-16", viewUniques: 5, cloneUniques: 2 },
    { date: "2026-09-17", viewUniques: 3, cloneUniques: 1 },
    { date: "2026-09-18", viewUniques: 0, cloneUniques: 0 },
    { date: "2026-09-19", viewUniques: 0, cloneUniques: 0 },
  ];

  test("drops trailing days GitHub has not settled yet", () => {
    const now = new Date("2026-09-19T03:00:00Z");
    expect(cutUnsettledDays(pulse, now).map((day) => day.date)).toEqual(["2026-09-16", "2026-09-17"]);
  });

  test("keeps a day once it is older than the settle window", () => {
    const now = new Date("2026-09-19T13:00:00Z");
    expect(cutUnsettledDays(pulse, now).map((day) => day.date)).toEqual([
      "2026-09-16",
      "2026-09-17",
      "2026-09-18",
    ]);
  });

  test("returns an empty list when every day is unsettled", () => {
    const fresh = [{ date: "2026-09-19", viewUniques: 1, cloneUniques: 0 }];
    expect(cutUnsettledDays(fresh, new Date("2026-09-19T03:00:00Z"))).toEqual([]);
  });
});

describe("pulseBars", () => {
  test("labels UTC days and scales both series to the same peak", () => {
    const bars = pulseBars(dailyCatalogPulse(catalog));
    expect(bars[0].label).toBe("Aug 27");
    expect(bars[0].viewHeight).toBe("100%");
    expect(bars[1].cloneHeight).toBe(`${Math.round((4 / 11) * 100)}%`);
  });
});

describe("uniqueVisitorsByLanguage", () => {
  test("groups unique visitors by language", () => {
    expect(uniqueVisitorsByLanguage(catalog).map((row) => [row.language, row.value])).toEqual([
      ["Go", 110],
      ["Python", 5],
    ]);
  });
});

describe("quietStarsRank", () => {
  test("ranks repositories with the most unique visitors per star", () => {
    expect(quietStarsRank(catalog).map((row) => row.name)).toEqual(["beta", "gamma", "alpha"]);
    expect(quietStarsRank(catalog)[0].uniques).toBe(70);
    expect(quietStarsRank(catalog)[0].stars).toBe(2);
  });
});
