// @CODE:NEXTRA-I18N-008
/**
 * Language Switcher Component
 *
 * Provides UI for switching between supported locales:
 * - Displays current locale
 * - Allows selection of Korean or English
 * - Preserves current path when switching
 */

'use client';

import React from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { useRouter, usePathname } from 'next/navigation';

export default function LanguageSwitcher() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();
  const t = useTranslations('common');

  const handleLanguageChange = (newLocale: string) => {
    // Remove current locale prefix from pathname
    const pathWithoutLocale = pathname.replace(`/${locale}`, '');

    // Build new path with new locale prefix
    const newPath = `/${newLocale}${pathWithoutLocale || '/'}`;

    // Navigate to new path
    router.push(newPath);
  };

  return (
    <select
      value={locale}
      onChange={(e) => handleLanguageChange(e.target.value)}
      className="border rounded px-2 py-1 text-sm bg-white dark:bg-gray-800"
      aria-label={t('language')}
    >
      <option value="ko">한국어</option>
      <option value="en">English</option>
    </select>
  );
}
