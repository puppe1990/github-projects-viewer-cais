import { fireEvent, render, screen, within } from "@testing-library/svelte";
import { describe, expect, test, vi } from "vitest";
import RepoCard from "./RepoCard.svelte";

const repo = {
  name: "cais",
  html_url: "https://github.com/puppe1990/cais",
  owner_login: "puppe1990",
  stargazers_count: 1,
  forks_count: 0,
  language: "Go",
  description: "Go on Cais",
  updated_at: "2026-08-28T00:00:00Z",
  traffic: {
    available: true,
    views: 17,
    view_uniques: 3,
    clones: 342,
    clone_uniques: 104,
    views_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 7, uniques: 3 }],
    clones_by_day: [{ timestamp: "2026-08-27T00:00:00Z", count: 5, uniques: 2 }],
  },
};

describe("RepoCard", () => {
  test("opens a traffic dialog from the card button", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false }));
    render(RepoCard, { props: { repo, index: 0 } });
    await fireEvent.click(screen.getByRole("button", { name: "View traffic" }));
    const dialog = screen.getByRole("dialog", { name: "Traffic · cais" });
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText("17")).toBeInTheDocument();
    expect(within(dialog).getByText("342")).toBeInTheDocument();
  });

  test("shows unique views and unique clones on the card", () => {
    render(RepoCard, { props: { repo, index: 0 } });
    expect(screen.getByTitle("17 views · 3 unique")).toHaveTextContent("3 unique");
    expect(screen.getByTitle("342 clones · 104 unique")).toHaveTextContent("104 unique");
  });

  test("shows a Private badge on private repositories", () => {
    render(RepoCard, { props: { repo: { ...repo, private: true, traffic: {} }, index: 0 } });
    expect(screen.getByText("Private")).toBeInTheDocument();
  });
});
