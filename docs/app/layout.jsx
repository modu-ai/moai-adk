import React from 'react'

/**
 * Root Layout for App Router
 *
 * This is the main layout component that wraps the entire application.
 * It handles HTML structure and global providers for the App Router.
 */
export default function RootLayout({ children }) {
  return (
    <html lang="ko">
      <body>
        <div id="root">
          {children}
        </div>
      </body>
    </html>
  )
}