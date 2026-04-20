import nextra from "nextra";

const withNextra = nextra({
	latex: true,
});

// Nextra 4 uses file-based i18n with content/ko, content/en, etc.
// Do NOT configure Next.js i18n when using Nextra's file-based i18n
export default withNextra({});
