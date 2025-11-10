import React, { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/router';

interface PagefindSearchProps {
  placeholder?: string;
  locale?: string;
}

declare global {
  interface Window {
    PagefindUI?: any;
    pagefind?: any;
  }
}

export const PagefindSearch: React.FC<PagefindSearchProps> = ({
  placeholder = "검색...",
  locale = "ko"
}) => {
  const searchContainerRef = useRef<HTMLDivElement>(null);
  const [isLoaded, setIsLoaded] = useState(false);
  const [currentLocale, setCurrentLocale] = useState(locale);
  const router = useRouter();

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
    const loadPagefind = async () => {
      try {
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

    if (searchContainerRef.current && currentLocale) {
      loadPagefind();
    }
  }, [currentLocale]);

  useEffect(() => {
    // Update locale when router changes
    if (router.locale && router.locale !== currentLocale) {
      setCurrentLocale(router.locale);
    }
  }, [router.locale, currentLocale]);

  return (
    <div className="pagefind-search-container">
      <div
        ref={searchContainerRef}
        className="pagefind-ui__search-input"
        data-pagefind-ui-input
        style={{
          width: '100%',
          maxWidth: '400px'
        }}
      />
      {!isLoaded && (
        <div className="pagefind-loading">
          <span>{translations[currentLocale]?.searching || "Loading search..."}</span>
        </div>
      )}
    </div>
  );
};

export default PagefindSearch;