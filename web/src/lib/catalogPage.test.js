import { describe, expect, test } from "vitest";
import { CATALOG_PAGE_SIZE, sliceCatalogPage } from "./catalogPage.js";

const items = Array.from({ length: 45 }, (_, i) => ({ name: `repo-${i + 1}` }));

describe("sliceCatalogPage", () => {
  test("keeps twenty repositories on the first page", () => {
    expect(CATALOG_PAGE_SIZE).toBe(20);
    const page = sliceCatalogPage(items, 1);
    expect(page.items.map((repo) => repo.name)).toEqual(items.slice(0, 20).map((repo) => repo.name));
    expect(page.page).toBe(1);
    expect(page.pages).toBe(3);
    expect(page.total).toBe(45);
  });

  test("returns the remainder on the last page", () => {
    const page = sliceCatalogPage(items, 3);
    expect(page.items.map((repo) => repo.name)).toEqual(["repo-41", "repo-42", "repo-43", "repo-44", "repo-45"]);
    expect(page.page).toBe(3);
  });

  test("clamps a page past the end back onto the last page", () => {
    expect(sliceCatalogPage(items, 99).page).toBe(3);
    expect(sliceCatalogPage(items, 0).page).toBe(1);
  });
});
