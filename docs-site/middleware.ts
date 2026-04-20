import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

/**
 * Supported locales for the documentation site
 */
const SUPPORTED_LOCALES = ["ko", "en", "zh", "ja"] as const;
const DEFAULT_LOCALE = "ko" as const;

type Locale = (typeof SUPPORTED_LOCALES)[number];

/**
 * i18n middleware for Nextra documentation
 *
 * Features:
 * - Detect locale from URL path prefix
 * - Fallback to Accept-Language header
 * - Store preference in cookie
 * - Redirect root to detected locale
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Skip middleware for API routes, static files, and Next.js internals
  if (
    pathname.startsWith("/api/") ||
    pathname.startsWith("/_next/") ||
    pathname.startsWith("/static/") ||
    pathname.includes(".") // static files with extensions
  ) {
    return NextResponse.next();
  }

  // Check if pathname already has a locale prefix
  const pathLocale = extractLocaleFromPath(pathname);

  if (pathLocale) {
    // Valid locale prefix exists
    // Update cookie preference to match current path
    const response = NextResponse.next();
    response.cookies.set("locale", pathLocale, {
      maxAge: 60 * 60 * 24 * 365, // 1 year
      path: "/",
      sameSite: "lax",
    });
    return response;
  }

  // No locale prefix detected - redirect to preferred locale
  const detectedLocale = detectLocale(request);
  const redirectUrl = new URL(`/${detectedLocale}${pathname}`, request.url);

  const response = NextResponse.redirect(redirectUrl);

  // Set locale preference in cookie
  response.cookies.set("locale", detectedLocale, {
    maxAge: 60 * 60 * 24 * 365, // 1 year
    path: "/",
    sameSite: "lax",
  });

  return response;
}

/**
 * Extract locale from pathname if it matches a supported locale
 */
function extractLocaleFromPath(pathname: string): Locale | null {
  const segments = pathname.split("/").filter(Boolean);

  if (segments.length === 0) {
    return null;
  }

  const firstSegment = segments[0];

  if (SUPPORTED_LOCALES.includes(firstSegment as Locale)) {
    return firstSegment as Locale;
  }

  return null;
}

/**
 * Detect user's preferred locale
 *
 * Priority order:
 * 1. Cookie preference
 * 2. Accept-Language header
 * 3. Default locale (ko)
 */
function detectLocale(request: NextRequest): Locale {
  // Check cookie first
  const cookieLocale = request.cookies.get("locale")?.value;
  if (cookieLocale && SUPPORTED_LOCALES.includes(cookieLocale as Locale)) {
    return cookieLocale as Locale;
  }

  // Check Accept-Language header
  const acceptLanguage = request.headers.get("accept-language");
  if (acceptLanguage) {
    const headerLocale = parseAcceptLanguage(acceptLanguage);
    if (headerLocale) {
      return headerLocale;
    }
  }

  // Return default locale
  return DEFAULT_LOCALE;
}

/**
 * Parse Accept-Language header and return best matching supported locale
 *
 * Format: "ko-KR,ko;q=0.9,en;q=0.8,ja;q=0.7"
 */
function parseAcceptLanguage(acceptLanguage: string): Locale | null {
  // Parse Accept-Language header into array of { locale, quality }
  const languages = acceptLanguage
    .split(",")
    .map((lang) => {
      const [locale, q] = lang.trim().split(";q=");
      const quality = q ? parseFloat(q) : 1.0;
      return { locale, quality };
    })
    .sort((a, b) => b.quality - a.quality);

  // Find first matching supported locale
  for (const { locale } of languages) {
    // Try exact match (e.g., "ko" matches "ko")
    if (SUPPORTED_LOCALES.includes(locale as Locale)) {
      return locale as Locale;
    }

    // Try prefix match (e.g., "ko-KR" matches "ko")
    const prefix = locale.split("-")[0];
    if (SUPPORTED_LOCALES.includes(prefix as Locale)) {
      return prefix as Locale;
    }
  }

  return null;
}

/**
 * Configure matcher for middleware
 *
 * Matcher excludes:
 * - API routes (/api/*)
 * - Static files (/_next/*, /static/*)
 * - Files with extensions
 */
export const config = {
  matcher: [
    /*
     * Match all paths except:
     * - api routes
     * - _next static files
     * - _vercel analytics
     * - files with extensions (images, fonts, etc.)
     */
    "/((?!api|_next|_vercel|.*\\..*).*)",
  ],
};
