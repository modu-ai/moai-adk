// Language Selector for MkDocs Material
document.addEventListener('DOMContentLoaded', function() {
  const header = document.querySelector('.md-header__inner');
  if (header && !document.querySelector('.md-select')) {
    // Detect current language from URL
    const path = window.location.pathname;
    let currentLang = 'ko'; // default
    if (path.startsWith('/en/')) currentLang = 'en';
    else if (path.startsWith('/ja/')) currentLang = 'ja';
    else if (path.startsWith('/zh/')) currentLang = 'zh';

    const languageNames = {
      'ko': '한국어',
      'en': 'English',
      'ja': '日本語',
      'zh': '中文'
    };

    const languages = [
      { name: '한국어', link: '/', lang: 'ko' },
      { name: 'English', link: '/en/', lang: 'en' },
      { name: '日本語', link: '/ja/', lang: 'ja' },
      { name: '中文', link: '/zh/', lang: 'zh' }
    ];

    const selector = document.createElement('div');
    selector.className = 'md-select';

    let listHTML = '';
    languages.forEach(lang => {
      const activeClass = lang.lang === currentLang ? ' md-select__item--active' : '';
      listHTML += `<a href="${lang.link}" class="md-select__item${activeClass}" hreflang="${lang.lang}">
        <span>${lang.name}</span>
      </a>`;
    });

    selector.innerHTML = `
      <button class="md-select__inner" aria-label="Select language">
        <span>${languageNames[currentLang]}</span>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" style="vertical-align: middle; margin-left: 0.25rem;">
          <path d="M7 10l5 5 5-5z" fill="currentColor"/>
        </svg>
      </button>
      <div class="md-select__list">
        ${listHTML}
      </div>
    `;

    // Insert before palette toggle
    const paletteForm = header.querySelector('form[data-md-component="palette"]');
    if (paletteForm) {
      header.insertBefore(selector, paletteForm);
    } else {
      header.appendChild(selector);
    }
  }
});
