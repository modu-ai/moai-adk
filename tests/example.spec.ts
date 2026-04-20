import { expect, test } from "@playwright/test";

test("homepage loads correctly", async ({ page }) => {
  await page.goto("/");

  await expect(page).toHaveTitle(/MoAI/);
});

test("navigation works", async ({ page }) => {
  await page.goto("/");

  const navigation = page.getByRole("navigation");
  await expect(navigation).toBeVisible();
});
