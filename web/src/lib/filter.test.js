import { describe, expect, test } from "vitest";
import { applyFilters, languageOptions } from "./filter.js";

const repos = [
  { name: "alpha", description: "first", language: "Go", stargazers_count: 10, forks_count: 1, updated_at: "2026-08-01T00:00:00Z" },
  { name: "beta", description: "second site", language: "Python", stargazers_count: 3, forks_count: 8, homepage: "https://example.com", fork: true, updated_at: "2026-08-20T00:00:00Z" },
  { name: "gamma", description: "tooling", language: "Go", stargazers_count: 30, forks_count: 2, updated_at: "2026-07-01T00:00:00Z" },
];

describe("applyFilters", () => {
  test("filters by min stars and sorts desc", () => {
    const got = applyFilters(repos, { minStars: 10, includeForks: true, sortBy: "stars", sortOrder: "desc" });
    expect(got.map((r) => r.name)).toEqual(["gamma", "alpha"]);
  });

  test("filters language and query", () => {
    const got = applyFilters(repos, { language: "Go", query: "tool", includeForks: true, sortBy: "name", sortOrder: "asc" });
    expect(got.map((r) => r.name)).toEqual(["gamma"]);
  });

  test("hides forks unless included", () => {
    const got = applyFilters(repos, { includeForks: false, hasHomepage: true, sortBy: "stars", sortOrder: "desc" });
    expect(got).toHaveLength(0);
  });

  test("lists public repositories by default and can show private or all", () => {
    const mixed = [
      { name: "open", private: false, stargazers_count: 1 },
      { name: "secret", private: true, stargazers_count: 1 },
    ];
    expect(applyFilters(mixed, { includeForks: true, sortBy: "name", sortOrder: "asc" }).map((r) => r.name)).toEqual(["open"]);
    expect(applyFilters(mixed, { includeForks: true, visibility: "private", sortBy: "name", sortOrder: "asc" }).map((r) => r.name)).toEqual([
      "secret",
    ]);
    expect(applyFilters(mixed, { includeForks: true, visibility: "all", sortBy: "name", sortOrder: "asc" }).map((r) => r.name)).toEqual([
      "open",
      "secret",
    ]);
  });

  test("sorts by traffic views and clones", () => {
    const withTraffic = [
      { name: "quiet", traffic: { views: 2, clones: 80, view_uniques: 2, clone_uniques: 10 } },
      { name: "busy", traffic: { views: 40, clones: 3, view_uniques: 20, clone_uniques: 3 } },
      { name: "cloned", traffic: { views: 8, clones: 90, view_uniques: 4, clone_uniques: 40 } },
    ];
    expect(applyFilters(withTraffic, { includeForks: true, sortBy: "traffic", sortOrder: "desc" }).map((r) => r.name)).toEqual([
      "busy",
      "cloned",
      "quiet",
    ]);
    expect(applyFilters(withTraffic, { includeForks: true, sortBy: "clones", sortOrder: "desc" }).map((r) => r.name)).toEqual([
      "cloned",
      "quiet",
      "busy",
    ]);
    expect(applyFilters(withTraffic, { includeForks: true, sortBy: "clone_uniques", sortOrder: "desc" }).map((r) => r.name)).toEqual([
      "cloned",
      "quiet",
      "busy",
    ]);
  });

  test("uses then-by when primary values tie", () => {
    const tied = [
      { name: "alpha", stargazers_count: 10, traffic: { views: 1 } },
      { name: "beta", stargazers_count: 10, traffic: { views: 9 } },
    ];
    expect(
      applyFilters(tied, { includeForks: true, sortBy: "stars", sortOrder: "desc", sortThen: "views", sortThenOrder: "desc" }).map(
        (r) => r.name,
      ),
    ).toEqual(["beta", "alpha"]);
  });
});

test("languageOptions lists unique languages", () => {
  expect(languageOptions(repos)).toEqual(["Go", "Python"]);
});
