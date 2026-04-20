import { Head } from "nextra/components";
import "nextra-theme-docs/style.css";
import "./globals.css";
import { Analytics } from "@vercel/analytics/next";
import type { Metadata } from "next";
import Script from "next/script";
import { MoAIStructuredData } from "@/components/structured-data";

export const metadata: Metadata = {
  title: {
    default: "MoAI-ADK Documentation",
    template: "%s | MoAI-ADK",
  },
  description: "MoAI-ADK 온라인 문서 - AI 기반 개발 도구",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ko" dir="ltr" suppressHydrationWarning>
      <Head>
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      </Head>
      <body>
        <MoAIStructuredData />
        <Script
          src="https://www.googletagmanager.com/gtag/js?id=G-FFRCE7KQ55"
          strategy="afterInteractive"
        />
        <Script id="google-analytics" strategy="afterInteractive">
          {`
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'G-FFRCE7KQ55');
          `}
        </Script>
        <Analytics />
        {children}
      </body>
    </html>
  );
}
