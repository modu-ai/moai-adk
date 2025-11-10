import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';

declare global {
  interface Window {
    PagefindHighlight?: any;
    pagefind?: any;
  }
}

interface SearchHighlightProviderProps {
  children: React.ReactNode;
}

export const SearchHighlightProvider: React.FC<SearchHighlightProviderProps> = ({ children }) => {
  const [isHighlightLoaded, setIsHighlightLoaded] = useState(false);
  const router = useRouter();
  const [currentPath, setCurrentPath] = useState(router.asPath);

  useEffect(() => {
    const loadHighlight = async () => {
      try {
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
            setIsHighlightLoaded(true);
            applyHighlighting();
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
      } catch (error) {
        console.error('Failed to load Pagefind Highlight:', error);
      }
    };

    loadHighlight();
  }, []);

  useEffect(() => {
    // Re-apply highlighting when the route changes
    if (currentPath !== router.asPath) {
      setCurrentPath(router.asPath);
      if (isHighlightLoaded) {
        setTimeout(applyHighlighting, 100); // Small delay to ensure DOM is ready
      }
    }
  }, [router.asPath, currentPath, isHighlightLoaded]);

  const applyHighlight = () => {
    if (typeof window !== 'undefined' && window.PagefindHighlight) {
      try {
        // Remove existing highlights
        const existingHighlights = document.querySelectorAll('.pagefind-highlight');
        existingHighlights.forEach(el => {
          el.classList.remove('pagefind-highlight');
          if (el.parentNode) {
            const parent = el.parentNode;
            while (el.firstChild) {
              parent.insertBefore(el.firstChild, el);
            }
            parent.removeChild(el);
          }
        });

        // Apply new highlights based on URL params
        const urlParams = new URLSearchParams(window.location.search);
        const highlightTerm = urlParams.get('highlight');

        if (highlightTerm) {
          new window.PagefindHighlight({
            highlightParam: "highlight",
            markOptions: {
              className: "pagefind-highlight",
              exclude: ["[data-pagefind-ignore]", "[data-pagefind-ignore] *"]
            }
          });

          // Scroll to first highlighted term
          const firstHighlight = document.querySelector('.pagefind-highlight');
          if (firstHighlight) {
            firstHighlight.scrollIntoView({
              behavior: 'smooth',
              block: 'center'
            });
          }
        }
      } catch (error) {
        console.error('Error applying search highlighting:', error);
      }
    }
  };

  const applyHighlighting = () => {
    // Small delay to ensure the page is fully rendered
    setTimeout(applyHighlight, 50);
  };

  return <>{children}</>;
};

export default SearchHighlightProvider;