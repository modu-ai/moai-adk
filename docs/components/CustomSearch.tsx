import React, { useEffect, useRef, useState } from 'react';

interface CustomSearchProps {
  placeholder?: string;
  className?: string;
}

declare global {
  interface Window {
    PagefindUI?: any;
    pagefind?: any;
  }
}

export const CustomSearch: React.FC<CustomSearchProps> = ({
  placeholder = "검색...",
  className = ""
}) => {
  const searchContainerRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [currentLocale, setCurrentLocale] = useState('ko');

  // Language-specific translations
  const translations: Record<string, Record<string, string>> = {
    ko: {
      placeholder: "검색...",
      clear_search: "검색 지우기",
      load_more: "더 보기",
      search_label: "검색",
      filters: "필터",
      zero_results: "검색 결과가 없습니다",
      many_results: "[SEARCH_TERM]에 대한 [COUNT]개의 결과",
      one_result: "[SEARCH_TERM]에 대한 1개의 결과",
      alt_search: "[SEARCH_TERM]에 대한 결과가 없습니다. [DIFFERENT_TERM]에 대한 결과를 표시합니다",
      search_suggestion: "[SEARCH_TERM]에 대한 결과가 없습니다. 다음 중 하나를 시도해 보세요:",
      searching: "검색 중..."
    },
    en: {
      placeholder: "Search...",
      clear_search: "Clear search",
      load_more: "Load more",
      search_label: "Search",
      filters: "Filters",
      zero_results: "No results for [SEARCH_TERM]",
      many_results: "[COUNT] results for [SEARCH_TERM]",
      one_result: "1 result for [SEARCH_TERM]",
      alt_search: "No results for [SEARCH_TERM]. Showing results for [DIFFERENT_TERM]",
      search_suggestion: "No results for [SEARCH_TERM]. Try one of the following:",
      searching: "Searching..."
    },
    ja: {
      placeholder: "検索...",
      clear_search: "検索をクリア",
      load_more: "もっと見る",
      search_label: "検索",
      filters: "フィルター",
      zero_results: "[SEARCH_TERM]の結果がありません",
      many_results: "[SEARCH_TERM]の[COUNT]件の結果",
      one_result: "[SEARCH_TERM]の1件の結果",
      alt_search: "[SEARCH_TERM]の結果がありません。[DIFFERENT_TERM]の結果を表示しています",
      search_suggestion: "[SEARCH_TERM]の結果がありません。次のいずれかを試してください：",
      searching: "検索中..."
    },
    zh: {
      placeholder: "搜索...",
      clear_search: "清除搜索",
      load_more: "加载更多",
      search_label: "搜索",
      filters: "筛选器",
      zero_results: "没有找到[SEARCH_TERM]的结果",
      many_results: "[SEARCH_TERM]的[COUNT]个结果",
      one_result: "[SEARCH_TERM]的1个结果",
      alt_search: "没有找到[SEARCH_TERM]的结果。显示[DIFFERENT_TERM]的结果",
      search_suggestion: "没有找到[SEARCH_TERM]的结果。请尝试以下之一：",
      searching: "搜索中..."
    }
  };

  useEffect(() => {
    // Detect current locale from URL path
    const path = window.location.pathname;
    const localeMatch = path.match(/^\/([a-z]{2})\//);
    const detectedLocale = localeMatch ? localeMatch[1] : 'ko';

    if (detectedLocale !== currentLocale) {
      setCurrentLocale(detectedLocale);
    }
  }, [currentLocale]);

  useEffect(() => {
    const loadPagefind = async () => {
      if (!searchContainerRef.current) return;

      try {
        // Destroy existing Pagefind instance if any
        if (window.PagefindUI) {
          const existingInstance = searchContainerRef.current.querySelector('.pagefind-ui');
          if (existingInstance) {
            searchContainerRef.current.innerHTML = '';
          }
        }

        // Load Pagefind CSS
        const cssLink = document.createElement('link');
        cssLink.rel = 'stylesheet';
        cssLink.href = `/${currentLocale}/pagefind/pagefind-ui.css`;
        document.head.appendChild(cssLink);

        // Load Pagefind UI JavaScript
        const script = document.createElement('script');
        script.src = `/${currentLocale}/pagefind/pagefind-ui.js`;
        script.async = true;

        script.onload = () => {
          if (window.PagefindUI && searchContainerRef.current) {
            // Initialize PagefindUI with language-specific settings
            new window.PagefindUI({
              element: searchContainerRef.current,
              bundlePath: `/${currentLocale}/pagefind/`,
              showSubResults: true,
              showImages: false,
              excerptLength: 30,
              translations: translations[currentLocale],
              highlightParam: "highlight",
              baseUrl: `/${currentLocale}/`,
              processResult: (result: any) => {
                // Process result to fix URLs and metadata
                if (result.url && !result.url.startsWith(`/${currentLocale}/`)) {
                  result.url = `/${currentLocale}${result.url}`;
                }
                return result;
              }
            });
            setIsLoaded(true);
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
        console.error('Failed to load Pagefind:', error);
      }
    };

    loadPagefind();
  }, [currentLocale]);

  return (
    <div className={`nx-search nx-relative ${className}`}>
      <div
        ref={searchContainerRef}
        className="pagefind-search-input"
        style={{
          width: '100%',
          height: '40px'
        }}
      />
      {!isLoaded && (
        <div className="nx-absolute nx-inset-y-0 nx-right-0 nx-flex nx-items-center nx-pr-3">
          <svg
            className="nx-animate-spin nx-h-4 nx-w-4"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="nx-opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            ></circle>
            <path
              className="nx-opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
        </div>
      )}
    </div>
  );
};

export default CustomSearch;