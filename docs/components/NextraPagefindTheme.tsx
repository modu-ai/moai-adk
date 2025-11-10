import React from 'react';
import { useThemeConfig } from 'nextra-theme-docs';
import PagefindSearch from './PagefindSearch';

interface NextraPagefindThemeProps {
  children: React.ReactNode;
}

const NextraPagefindTheme: React.FC<NextraPagefindThemeProps> = ({ children }) => {
  const themeConfig = useThemeConfig();
  const { locale } = React.useContext(React.createContext({ locale: 'ko' }));

  return (
    <div className="nextra-container">
      {/* Override default search with Pagefind */}
      <div className="nextra-search-container">
        <PagefindSearch
          placeholder={themeConfig.search?.placeholder || "검색..."}
          locale={locale}
        />
      </div>

      {/* Render the rest of the theme */}
      {children}
    </div>
  );
};

export default NextraPagefindTheme;