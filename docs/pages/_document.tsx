import { Html, Head, Main, NextScript } from 'next/document'
import type { DocumentProps } from 'next/document'

export default function Document(props: DocumentProps) {
  return (
    <Html lang={props.locale || 'ko'}>
      <Head>
        {/* Preconnect to external domains for performance */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link rel="preconnect" href="https://github.com" />

        {/* Fonts - Pretendard for Korean, Source Han Sans for CJK, Hack for code */}
        <link
          href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.6/dist/web/static/pretendard.css"
          rel="stylesheet"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Source+Han+Sans+JP:wght@300;400;500;600;700;900&display=swap"
          rel="stylesheet"
        />
        <link
          href="https://fonts.googleapis.com/css2?family=Source+Han+Sans+CN:wght@300;400;500;600;700;900&display=swap"
          rel="stylesheet"
        />
        <link
          href="https://cdn.jsdelivr.net/npm/hack-font@3.3.0/build/web/hack.css"
          rel="stylesheet"
        />

        {/* Skip link for accessibility */}
        <style dangerouslySetInnerHTML={{
          __html: `
            .skip-link {
              position: absolute;
              top: -40px;
              left: 6px;
              background: #000000;
              color: white;
              padding: 8px;
              text-decoration: none;
              border-radius: 4px;
              z-index: 1000;
              transition: top 0.3s ease;
            }
            .skip-link:focus {
              top: 6px;
            }
          `
        }} />
      </Head>
      <body>
        {/* Skip link for keyboard navigation */}
        <a href="#main-content" className="skip-link">
          메인 콘텐츠로 바로가기
        </a>
        <Main />
        <NextScript />
      </body>
    </Html>
  )
}