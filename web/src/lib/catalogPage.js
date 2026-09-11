export const CATALOG_PAGE_SIZE = 20;

export function sliceCatalogPage(items, page, size = CATALOG_PAGE_SIZE) {
  const list = items || [];
  const total = list.length;
  const pages = Math.max(1, Math.ceil(total / size) || 1);
  const current = Math.min(Math.max(1, Number(page) || 1), pages);
  const start = (current - 1) * size;
  return {
    items: list.slice(start, start + size),
    page: current,
    pages,
    total,
  };
}
