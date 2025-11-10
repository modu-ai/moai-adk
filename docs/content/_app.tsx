import type { AppProps } from 'next/app'
import '@/styles/globals.css'

/**
 * MoAI-ADK Documentation App Component
 *
 * Global styling setup
 */
export default function App({ Component, pageProps }: AppProps) {
  return <Component {...pageProps} />
}
