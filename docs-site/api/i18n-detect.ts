/**
 * i18n-detect.ts — Vercel Edge Function for locale detection
 * SPEC-DOCS-SITE-001 Phase 5, REQ-DS-13/14
 *
 * Ported from moai-docs/middleware.ts (164 LOC, Next.js) to Vercel Edge Function.
 * Compatible with Hugo static build + Vercel Edge runtime.
 *
 * Priority order:
 *   1. URL path already has locale prefix -> pass through, update cookie
 *   2. Cookie (locale) -> use stored preference
 *   3. Accept-Language header -> parse best matching locale
 *   4. Default -> ko
 *
 * Bypass paths (no redirect):
 *   - /api/*
 *   - /_next/*
 *   - /static/*
 *   - Any path containing a file extension (e.g. /robots.txt, /og.png)
 */

// @MX:ANCHOR: [AUTO] Edge Function entry point — fan_in includes Vercel rewrite + direct access
// @MX:REASON: Vercel routes all unmatched requests here; changes affect all 4-locale redirect flows

export const config = { runtime: "edge" };

const SUPPORTED_LOCALES = ["ko", "en", "ja", "zh"] as const;
const DEFAULT_LOCALE = "ko" as const;
const COOKIE_NAME = "locale";
const COOKIE_MAX_AGE = 60 * 60 * 24 * 365; // 1 year

type Locale = (typeof SUPPORTED_LOCALES)[number];

/**
 * Parse Accept-Language header and return best matching supported locale.
 *
 * Format: "ko-KR,ko;q=0.9,en;q=0.8"
 */
function parseAcceptLanguage(header: string | null): Locale | null {
  if (!header) return null;

  const languages = header
    .split(",")
    .map((lang) => {
      const parts = lang.trim().split(";q=");
      const locale = parts[0].trim();
      const quality = parts[1] ? parseFloat(parts[1]) : 1.0;
      return { locale, quality };
    })
    .sort((a, b) => b.quality - a.quality);

  for (const { locale } of languages) {
    // Exact match (e.g. "ko")
    if (SUPPORTED_LOCALES.includes(locale as Locale)) {
      return locale as Locale;
    }
    // Prefix match (e.g. "ko-KR" -> "ko")
    const prefix = locale.split("-")[0];
    if (SUPPORTED_LOCALES.includes(prefix as Locale)) {
      return prefix as Locale;
    }
  }

  return null;
}

/**
 * Extract locale from pathname if it starts with a supported locale prefix.
 * e.g. "/ko/getting-started" -> "ko"
 *      "/unknown/page"       -> null
 */
function extractLocaleFromPath(pathname: string): Locale | null {
  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) return null;

  const first = segments[0];
  if (SUPPORTED_LOCALES.includes(first as Locale)) {
    return first as Locale;
  }
  return null;
}

/**
 * Parse locale from cookie header string.
 * e.g. "locale=ja; other=value" -> "ja"
 */
function parseCookieLocale(cookieHeader: string | null): Locale | null {
  if (!cookieHeader) return null;

  const match = cookieHeader.match(
    new RegExp(`(?:^|;\\s*)${COOKIE_NAME}=([a-z]{2})(?:;|$)`)
  );
  if (!match) return null;

  const value = match[1];
  return SUPPORTED_LOCALES.includes(value as Locale) ? (value as Locale) : null;
}

/**
 * Build Set-Cookie header value.
 */
function buildSetCookie(locale: Locale): string {
  return `${COOKIE_NAME}=${locale}; Path=/; Max-Age=${COOKIE_MAX_AGE}; SameSite=Lax`;
}

// @MX:WARN: [AUTO] Edge runtime constraint — no Node.js fs/path/crypto APIs allowed
// @MX:REASON: Vercel Edge executes in V8 isolate; only Web Platform APIs (URL, Headers, Request, Response) available

export default function handler(req: Request): Response {
  const url = new URL(req.url);
  const { pathname } = url;

  // Bypass: API routes, Next.js internals, static assets, and paths with file extensions
  if (
    pathname.startsWith("/api/") ||
    pathname.startsWith("/_next/") ||
    pathname.startsWith("/_vercel/") ||
    pathname.startsWith("/static/") ||
    /\.[a-zA-Z0-9]+$/.test(pathname)
  ) {
    // Pass through — do not redirect
    return new Response(null, { status: 200 });
  }

  // Check if URL path already has a supported locale prefix
  const pathLocale = extractLocaleFromPath(pathname);
  if (pathLocale) {
    // Already localized — update cookie preference and pass through
    return new Response(null, {
      status: 200,
      headers: {
        "Set-Cookie": buildSetCookie(pathLocale),
      },
    });
  }

  // Detect locale: cookie > Accept-Language > default
  const cookieLocale = parseCookieLocale(req.headers.get("cookie"));
  const acceptLocale = parseAcceptLanguage(req.headers.get("accept-language"));
  const locale: Locale = cookieLocale ?? acceptLocale ?? DEFAULT_LOCALE;

  // Build redirect URL: /{locale}{original-path}{search}
  const targetPath =
    pathname === "/" ? `/${locale}/` : `/${locale}${pathname}`;
  const redirectUrl = new URL(targetPath + url.search, req.url);

  return new Response(null, {
    status: 307,
    headers: {
      Location: redirectUrl.toString(),
      "Set-Cookie": buildSetCookie(locale),
    },
  });
}
