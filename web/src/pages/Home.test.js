import { fireEvent, render, screen, within } from "@testing-library/svelte";
import { router } from "@inertiajs/svelte";
import { describe, expect, test, vi } from "vitest";
import Home from "./Home.svelte";

vi.mock("@inertiajs/svelte", () => ({
  router: { get: vi.fn() },
}));

describe("Home", () => {
  test("renders the catalog masthead", () => {
    render(Home, { props: { populated: false, repos: [], orgs: [], profile: {}, source: {}, error: "" } });
    expect(screen.getByRole("heading", { name: /Projects/ })).toBeInTheDocument();
    expect(screen.getByLabelText("GitHub username")).toBeInTheDocument();
  });

  test("renders a populated profile and repo card", () => {
    render(Home, {
      props: {
        populated: true,
        lookup: "octocat",
        profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 8, followers: 1, following: 1 },
        orgs: [],
        source: { type: "user", login: "octocat" },
        repos: [{ name: "hello-world", html_url: "https://github.com/octocat/hello-world", stargazers_count: 10, forks_count: 2, language: "Go", description: "demo", updated_at: "2026-08-01T00:00:00Z" }],
        error: "",
      },
    });
    expect(screen.getByRole("heading", { name: "The Octocat" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "hello-world" })).toBeInTheDocument();
  });

  test("offers unique sorts in the quick dropdown", () => {
    render(Home, {
      props: {
        populated: true,
        lookup: "octocat",
        profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 1, followers: 1, following: 1 },
        orgs: [],
        source: { type: "user", login: "octocat" },
        repos: [{ name: "hello-world", html_url: "https://github.com/octocat/hello-world", stargazers_count: 10, forks_count: 2, language: "Go", description: "demo", updated_at: "2026-08-01T00:00:00Z", traffic: { available: true } }],
        error: "",
      },
    });
    const sortSelect = screen.getByLabelText("Sort by");
    expect(within(sortSelect).getByRole("option", { name: "Unique visitors" })).toBeInTheDocument();
    expect(within(sortSelect).getByRole("option", { name: "Unique cloners" })).toBeInTheDocument();
  });

  test("opens a complete sort dialog with traffic fields", async () => {
    render(Home, {
      props: {
        populated: true,
        lookup: "octocat",
        profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 1, followers: 1, following: 1 },
        orgs: [],
        source: { type: "user", login: "octocat" },
        repos: [{ name: "hello-world", html_url: "https://github.com/octocat/hello-world", stargazers_count: 10, forks_count: 2, language: "Go", description: "demo", updated_at: "2026-08-01T00:00:00Z", traffic: { available: true, views: 12, clones: 4 } }],
        error: "",
      },
    });
    await fireEvent.click(screen.getByRole("button", { name: "More sort options" }));
    const dialog = screen.getByRole("dialog", { name: "Arrange this catalog" });
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getAllByRole("button", { name: "Traffic" }).length).toBeGreaterThan(0);
    expect(within(dialog).getAllByRole("button", { name: "Unique visitors" }).length).toBeGreaterThan(0);
    expect(within(dialog).getAllByRole("button", { name: "Clones" }).length).toBeGreaterThan(0);
  });

  test("shows an All catalog chip that opens the combined catalog", async () => {
    render(Home, {
      props: {
        populated: true,
        lookup: "puppe1990",
        profile: { login: "puppe1990", name: "Matheus", html_url: "https://github.com/puppe1990", avatar_url: "https://example.com/a.png", public_repos: 8, followers: 1, following: 1 },
        orgs: [
          { login: "hidden-org", avatar_url: "https://example.com/h.png" },
          { login: "purchasestore", avatar_url: "https://example.com/p.png" },
        ],
        source: { type: "all", login: "puppe1990" },
        repos: [
          { name: "cais", html_url: "https://github.com/puppe1990/cais", stargazers_count: 1, forks_count: 0, language: "Go", description: "demo", updated_at: "2026-08-01T00:00:00Z" },
          { name: "private-app", html_url: "https://github.com/hidden-org/private-app", stargazers_count: 0, forks_count: 0, language: "Go", description: "org", updated_at: "2026-08-01T00:00:00Z" },
        ],
        error: "",
      },
    });
    const allChip = screen.getByRole("button", { name: /^All\b/ });
    expect(allChip).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "hidden-org" })).toBeInTheDocument();
    expect(screen.getByText(/across personal and orgs/)).toBeInTheDocument();
    await fireEvent.click(allChip);
    expect(router.get).toHaveBeenCalledWith("/u/puppe1990/all", {}, expect.any(Object));
  });
});
