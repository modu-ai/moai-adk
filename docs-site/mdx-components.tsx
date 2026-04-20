import { useMDXComponents as getDocsMDXComponents } from "nextra-theme-docs";

const docsComponents = getDocsMDXComponents();

export const useMDXComponents = (
  components?: Record<string, React.ComponentType>,
) => ({
  ...docsComponents,
  ...components,
});
