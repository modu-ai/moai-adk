import React, { useEffect } from 'react';

declare global {
  interface Window {
    PagefindHighlight?: any;
  }
}

interface PagefindHighlightProps {
  children: React.ReactNode;
}

export const PagefindHighlight: React.FC<PagefindHighlightProps> = ({ children }) => {
  useEffect(() => {
    // Load Pagefind Highlight CSS
    const cssLink = document.createElement('link');
    cssLink.rel = 'stylesheet';
    cssLink.href = '/pagefind/pagefind-highlight.css';
    document.head.appendChild(cssLink);

    // Load Pagefind Highlight JavaScript
    const script = document.createElement('script');
    script.src = '/pagefind/pagefind-highlight.js';
    script.async = true;

    script.onload = () => {
      if (window.PagefindHighlight) {
        // Initialize highlighting with current URL search params
        const urlParams = new URLSearchParams(window.location.search);
        const highlightTerm = urlParams.get('highlight');

        if (highlightTerm) {
          new window.PagefindHighlight({
            highlightParam: "highlight"
          });
        }
      }
    };

    document.body.appendChild(script);

    return () => {
      // Cleanup
      if (document.head.contains(cssLink)) {
        document.head.removeChild(cssLink);
      }
      if (document.body.contains(script)) {
        document.body.removeChild(script);
      }
    };
  }, []);

  return <>{children}</>;
};

export default PagefindHighlight;