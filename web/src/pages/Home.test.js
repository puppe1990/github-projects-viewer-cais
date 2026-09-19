import { fireEvent, render, screen, within } from "@testing-library/svelte";
import { router } from "@inertiajs/svelte";
import { describe, expect, test, vi } from "vitest";
import Home from "./Home.svelte";

vi.mock("@inertiajs/svelte", () => ({
  router: { get: vi.fn() },
  inertia: () => {},
}));

describe("Home", () => {
  test("renders the catalog masthead", () => {
    render(Home, { props: { populated: false, repos: [], orgs: [], profile: {}, source: {}, error: "" } });
    expect(screen.getByRole("heading", { name: /Projects/ })).toBeInTheDocument();
    expect(screen.getByLabelText("GitHub username")).toBeInTheDocument();
  });

  test("shows a catalog spinner while looking up a username", async () => {
    render(Home, { props: { populated: false, repos: [], orgs: [], profile: {}, source: {}, error: "" } });
    await fireEvent.input(screen.getByLabelText("GitHub username"), { target: { value: "octocat" } });
    await fireEvent.click(screen.getByRole("button", { name: "Look up" }));
    expect(screen.getByRole("status", { name: /Reading the public graph/ })).toBeInTheDocument();
    expect(document.querySelector(".atlas-spinner")).toBeInTheDocument();
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
    expect(screen.queryByRole("tab", { name: "Charts" })).not.toBeInTheDocument();
    expect(within(screen.getByLabelText("Sort by")).queryByRole("option", { name: "Unique visitors" })).not.toBeInTheDocument();
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

  const trafficHomeProps = {
    populated: true,
    lookup: "octocat",
    profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 1, followers: 1, following: 1 },
    orgs: [],
    source: { type: "user", login: "octocat" },
    repos: [
      {
        name: "hello-world",
        html_url: "https://github.com/octocat/hello-world",
        stargazers_count: 10,
        forks_count: 2,
        language: "Go",
        description: "demo",
        updated_at: "2026-08-01T00:00:00Z",
        traffic: { available: true, views: 40, view_uniques: 22, clones: 9, clone_uniques: 6 },
      },
    ],
    error: "",
  };

  test("links catalog and charts to their own routes", () => {
    render(Home, { props: trafficHomeProps });
    const catalog = screen.getByRole("tab", { name: "Catalog" });
    const charts = screen.getByRole("tab", { name: "Charts" });
    expect(catalog).toHaveAttribute("href", "/u/octocat");
    expect(charts).toHaveAttribute("href", "/u/octocat/charts");
    expect(catalog).toHaveAttribute("aria-selected", "true");
    expect(charts).toHaveAttribute("aria-selected", "false");
    expect(screen.getByRole("tabpanel", { name: "Catalog" })).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Top unique visitors" })).not.toBeInTheDocument();
  });

  test("renders the charts panel on its own route", () => {
    render(Home, { props: { ...trafficHomeProps, view: "charts" } });
    expect(screen.getByRole("tab", { name: "Charts" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("region", { name: "Top unique visitors" })).toBeInTheDocument();
    expect(screen.queryByRole("tabpanel", { name: "Catalog" })).not.toBeInTheDocument();
  });

  test("links the charts route for the combined catalog", () => {
    render(Home, {
      props: { ...trafficHomeProps, profile: { ...trafficHomeProps.profile, login: "puppe1990" }, source: { type: "all", login: "puppe1990" } },
    });
    expect(screen.getByRole("tab", { name: "Charts" })).toHaveAttribute("href", "/u/puppe1990/all/charts");
  });

  test("links the charts route for an org catalog", () => {
    render(Home, {
      props: { ...trafficHomeProps, source: { type: "org", login: "github" } },
    });
    expect(screen.getByRole("tab", { name: "Charts" })).toHaveAttribute("href", "/u/octocat/orgs/github/charts");
  });

  test("paginates the catalog twenty repositories at a time", async () => {
    const repos = Array.from({ length: 21 }, (_, i) => ({
      name: `repo-${i + 1}`,
      html_url: `https://github.com/octocat/repo-${i + 1}`,
      stargazers_count: 21 - i,
      forks_count: 0,
      language: "Go",
      description: "demo",
      updated_at: "2026-08-01T00:00:00Z",
    }));
    render(Home, {
      props: {
        populated: true,
        lookup: "octocat",
        profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 21, followers: 1, following: 1 },
        orgs: [],
        source: { type: "user", login: "octocat" },
        repos,
        error: "",
      },
    });
    expect(screen.getByRole("link", { name: "repo-1" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "repo-20" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "repo-21" })).not.toBeInTheDocument();
    await fireEvent.click(screen.getByRole("button", { name: "Next page" }));
    expect(screen.getByRole("link", { name: "repo-21" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "repo-1" })).not.toBeInTheDocument();
  });

  test("lists public repositories by default and offers visibility options", async () => {
    render(Home, {
      props: {
        populated: true,
        lookup: "octocat",
        profile: { login: "octocat", name: "The Octocat", html_url: "https://github.com/octocat", avatar_url: "https://example.com/a.png", public_repos: 2, followers: 1, following: 1 },
        orgs: [],
        source: { type: "user", login: "octocat" },
        repos: [
          { name: "hello-world", html_url: "https://github.com/octocat/hello-world", stargazers_count: 10, forks_count: 0, language: "Go", description: "public", updated_at: "2026-08-01T00:00:00Z", private: false },
          { name: "secret", html_url: "https://github.com/octocat/secret", stargazers_count: 1, forks_count: 0, language: "Go", description: "private", updated_at: "2026-08-01T00:00:00Z", private: true },
        ],
        error: "",
      },
    });
    expect(screen.getByRole("link", { name: "hello-world" })).toBeInTheDocument();
    expect(screen.queryByRole("link", { name: "secret" })).not.toBeInTheDocument();
    const visibility = screen.getByLabelText("Visibility");
    expect(within(visibility).getByRole("option", { name: "Public" })).toBeInTheDocument();
    expect(within(visibility).getByRole("option", { name: "Private" })).toBeInTheDocument();
    expect(within(visibility).getByRole("option", { name: "All" })).toBeInTheDocument();
    await fireEvent.change(visibility, { target: { value: "all" } });
    expect(screen.getByRole("link", { name: "secret" })).toBeInTheDocument();
  });
});


