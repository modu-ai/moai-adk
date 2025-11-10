import React from 'react'

/**
 * Dynamic Page Component for App Router
 *
 * This handles all page routing within locales using catch-all dynamic routes.
 * Supports both homepage (/) and nested pages (/guide/example).
 */
export default function Page({ params }) {
  const { locale, slug = [] } = params

  // Join slug array to create path string
  const path = slug.length > 0 ? `/${slug.join('/')}` : ''
  const fullPath = `${locale}${path}`

  return (
    <div>
      <h1>Nextra Page</h1>
      <p>Locale: {locale}</p>
      <p>Path: {fullPath}</p>
      <p>Slug: {JSON.stringify(slug)}</p>
    </div>
  )
}