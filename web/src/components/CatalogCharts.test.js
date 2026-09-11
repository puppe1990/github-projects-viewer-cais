import { render, screen, within } from "@testing-library/svelte";
import { describe, expect, test } from "vitest";
import CatalogCharts from "./CatalogCharts.svelte";

const repos = [
  {
    name: "cais",
    html_url: "https://github.com/puppe1990/cais",
    language: "Go",
    stargazers_count: 2,
    traffic: {
      available: true,
      views: 80,
      view_uniques: 70,
      clones: 5,
      clone_uniques: 4,
      views_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 10, uniques: 8 }],
      clones_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 2, uniques: 1 }],
    },
  },
  {
    name: "atlas",
    html_url: "https://github.com/puppe1990/atlas",
    language: "Python",
    stargazers_count: 40,
    traffic: {
      available: true,
      views: 20,
      view_uniques: 12,
      clones: 90,
      clone_uniques: 60,
    },
  },
  { name: "silent", language: "Rust", stargazers_count: 1 },
];

describe("CatalogCharts", () => {
  test("explains when the catalog has no traffic snapshots", () => {
    render(CatalogCharts, { props: { repos: [{ name: "silent" }] } });
    expect(screen.getByText(/No traffic snapshots/)).toBeInTheDocument();
  });

  test("shows top unique visitors and unique cloners", () => {
    render(CatalogCharts, { props: { repos } });
    const visitors = screen.getByRole("region", { name: "Top unique visitors" });
    expect(within(visitors).getByRole("link", { name: "cais" })).toHaveAttribute("href", "https://github.com/puppe1990/cais");
    expect(within(visitors).getByText("70")).toBeInTheDocument();
    const cloners = screen.getByRole("region", { name: "Top unique cloners" });
    expect(within(cloners).getByRole("link", { name: "atlas" })).toHaveAttribute("href", "https://github.com/puppe1990/atlas");
    expect(within(cloners).getByText("60")).toBeInTheDocument();
  });

  test("shows catalog totals, language mix, pulse, and quiet stars", () => {
    render(CatalogCharts, { props: { repos } });
    expect(screen.getByText("82")).toBeInTheDocument();
    expect(screen.getByText("64")).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Unique visitors by language" })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Last 14 days" })).toBeInTheDocument();
    const quiet = screen.getByRole("region", { name: "Quiet stars" });
    expect(within(quiet).getByRole("link", { name: "cais" })).toHaveAttribute("href", "https://github.com/puppe1990/cais");
    expect(within(screen.getByRole("region", { name: "Unique visitors by language" })).queryByRole("link")).not.toBeInTheDocument();
  });
});
