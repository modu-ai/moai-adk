"use client";

import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import { ChevronDown } from "lucide-react";
import { usePathname } from "next/navigation";
import * as React from "react";

import { cn } from "@/lib/utils";

// Language configuration
const LANGUAGES = [
  { code: "ko", name: "한국어", flag: "🇰🇷" },
  { code: "en", name: "English", flag: "🇺🇸" },
  { code: "ja", name: "日本語", flag: "🇯🇵" },
  { code: "zh", name: "中文", flag: "🇨🇳" },
] as const;

type LanguageCode = (typeof LANGUAGES)[number]["code"];

/**
 * Extract language code from pathname
 * Handles both prefixed (/ko/...) and non-prefixed paths
 */
function getLocaleFromPathname(pathname: string): LanguageCode {
  // Check if pathname starts with a known locale prefix
  const segments = pathname.split("/").filter(Boolean);
  const firstSegment = segments[0];

  if (firstSegment && LANGUAGES.some((lang) => lang.code === firstSegment)) {
    return firstSegment as LanguageCode;
  }

  // Default to Korean for root paths
  return "ko";
}

/**
 * Build localized pathname by replacing or adding locale prefix
 */
function buildLocalizedPathname(
  pathname: string,
  targetLocale: LanguageCode,
): string {
  const _currentLocale = getLocaleFromPathname(pathname);
  const segments = pathname.split("/").filter(Boolean);

  let pathWithoutLocale: string[];

  if (LANGUAGES.some((lang) => lang.code === segments[0])) {
    // Remove current locale prefix
    pathWithoutLocale = segments.slice(1);
  } else {
    // No locale prefix, keep all segments
    pathWithoutLocale = segments;
  }

  // For default locale (ko), we can omit the prefix if desired
  // For now, always include the prefix for consistency
  if (pathWithoutLocale.length === 0) {
    return `/${targetLocale}`;
  }

  return `/${targetLocale}/${pathWithoutLocale.join("/")}`;
}

/**
 * LanguageSelector dropdown component for Nextra documentation
 *
 * Provides language switching functionality with flag emoji display
 * and preserves current page path when switching languages.
 *
 * @example
 * ```tsx
 * import LanguageSelector from '@/components/LanguageSelector'
 *
 * // In theme.config.tsx navbar
 * <LanguageSelector />
 * ```
 */
export default function LanguageSelector() {
  const pathname = usePathname();
  const [currentLocale, setCurrentLocale] = React.useState<LanguageCode>("ko");

  // Update current locale when pathname changes
  React.useEffect(() => {
    setCurrentLocale(getLocaleFromPathname(pathname));
  }, [pathname]);

  const currentLanguage = LANGUAGES.find((lang) => lang.code === currentLocale);

  const handleLanguageChange = (targetLocale: LanguageCode) => {
    const newPathname = buildLocalizedPathname(pathname, targetLocale);
    window.location.href = newPathname;
  };

  // Safely get language properties with fallbacks
  const flag = currentLanguage?.flag ?? "🌐";
  const _name = currentLanguage?.name ?? "Language";

  return (
    <DropdownMenuPrimitive.Root modal={false}>
      <DropdownMenuPrimitive.Trigger
        className={cn(
          "x:inline-flex x:items-center x:gap-1.5 x:shrink-0 x:whitespace-nowrap",
          "x:rounded-md x:p-2 x:text-sm x:font-medium",
          "x:transition-colors x:cursor-pointer",
          "x:text-gray-600 x:hover:text-black x:dark:text-gray-400 x:dark:hover:text-gray-200",
        )}
        aria-label="Select language"
      >
        <span className="x:text-base x:leading-none">{flag}</span>
        <ChevronDown className="x:size-3 x:opacity-50" aria-hidden="true" />
      </DropdownMenuPrimitive.Trigger>

      <DropdownMenuPrimitive.Content
        align="end"
        sideOffset={8}
        className={cn(
          "x:z-50 x:min-w-[180px] x:overflow-hidden x:rounded-md x:p-1",
          "x:border x:border-gray-200 x:dark:border-neutral-700",
          "x:bg-white x:dark:bg-neutral-900",
          "x:text-gray-900 x:dark:text-gray-100",
          "x:shadow-lg",
        )}
      >
        <div className="x:px-2 x:py-1.5 x:text-xs x:font-semibold x:text-gray-500 x:dark:text-gray-400">
          Language / 언어
        </div>

        {LANGUAGES.map((language) => {
          const isSelected = language.code === currentLocale;

          return (
            <DropdownMenuPrimitive.Item
              key={language.code}
              onClick={() => handleLanguageChange(language.code)}
              className={cn(
                "x:relative x:flex x:cursor-pointer x:select-none x:items-center x:gap-3",
                "x:rounded-sm x:px-2 x:py-2 x:text-sm x:outline-none",
                "x:transition-colors",
                "x:hover:bg-gray-100 x:dark:hover:bg-neutral-800",
                "focus:x:bg-gray-100 dark:focus:x:bg-neutral-800",
                isSelected &&
                  "x:bg-gray-100 x:dark:bg-neutral-800 x:font-medium",
              )}
            >
              <span className="x:text-lg" aria-hidden="true">
                {language.flag}
              </span>
              <span className="x:flex-1">{language.name}</span>
              {isSelected && (
                <span className="x:text-xs x:text-gray-400 x:dark:text-gray-500">
                  ✓
                </span>
              )}
            </DropdownMenuPrimitive.Item>
          );
        })}
      </DropdownMenuPrimitive.Content>
    </DropdownMenuPrimitive.Root>
  );
}
