import React from 'react'

/**
 * Locale Layout for App Router
 *
 * This layout handles locale-based routing and internationalization.
 * It receives the locale parameter from the dynamic route.
 */
export default function LocaleLayout({ children, params }) {
  const { locale } = params

  return (
    <html lang={locale}>
      <body>
        <LocaleProvider locale={locale}>
          {children}
        </LocaleProvider>
      </body>
    </html>
  )
}

/**
 * Locale Provider Component
 *
 * Wraps content with locale-specific providers for internationalization.
 */
function LocaleProvider({ children, locale }) {
  return (
    <div data-locale={locale}>
      {children}
    </div>
  )
}