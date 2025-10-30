// @CODE:NEXTRA-I18N-004
/**
 * i18n Routing Configuration
 *
 * Defines locale settings for next-intl:
 * - Supported locales: Korean (ko), English (en)
 * - Default locale: Korean (ko)
 * - Locale prefix: as-needed (shows /en/ but not /ko/)
 */

import { defineRouting } from 'next-intl/routing';

export const routing = defineRouting({
  // Supported locales
  locales: ['ko', 'en'],

  // Default locale (fallback)
  defaultLocale: 'ko',

  // Locale prefix strategy
  // 'as-needed': default locale (/ko/) path prefix is optional
  // 'always': all locales require prefix (/ko/, /en/)
  localePrefix: 'as-needed',
});
