// @CODE:NEXTRA-I18N-004
/**
 * i18n Request Configuration
 *
 * Handles message loading for each locale:
 * - Validates locale against supported locales
 * - Falls back to default locale for unsupported locales
 * - Loads appropriate message files
 */

import { getRequestConfig } from 'next-intl/server';
import { routing } from './routing';

export default getRequestConfig(async ({ locale }) => {
  // Validate locale and fallback to default if unsupported
  const validLocale = routing.locales.includes(locale as 'ko' | 'en')
    ? locale
    : routing.defaultLocale;

  return {
    messages: (await import(`../messages/${validLocale}.json`)).default,
  };
});
