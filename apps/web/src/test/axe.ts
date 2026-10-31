import axe from "axe-core";
import { expect } from "vitest";

/**
 * axe-core で、要素の中にアクセシビリティの違反がないことを確かめる。
 * jsdom は配置と色を計算しないため、color-contrast は目視で確かめる（quickstart.md）。
 */
export async function expectNoAxeViolations(container: Element) {
  const results = await axe.run(container, {
    rules: { "color-contrast": { enabled: false } },
  });
  const violations = results.violations.map(
    (v) => `${v.id}: ${v.help}\n  ${v.nodes.map((n) => n.html).join("\n  ")}`,
  );
  expect(violations).toEqual([]);
}
