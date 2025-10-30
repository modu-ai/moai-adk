// @CODE:NEXTRA-I18N-006
/**
 * i18n Middleware
 *
 * Handles locale detection and routing:
 * - Detects locale from Accept-Language header
 * - Redirects root path to appropriate locale
 * - Manages locale-prefixed paths
 */

import createMiddleware from 'next-intl/middleware';
import { routing } from './i18n/routing';

export default createMiddleware(routing);

export const config = {
  // Matcher configuration for middleware execution
  matcher: [
    // Match all paths except:
    // - Next.js internals (_next, _vercel)
    // - Static files (contains a dot)
    '/((?!_next|_vercel|.*\\..*).*)',
  ],
};
