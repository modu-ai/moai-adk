// Language Selector for MkDocs Material with Icon Support
document.addEventListener('DOMContentLoaded', function() {
  const header = document.querySelector('.md-header__inner');
  if (header && !document.querySelector('.md-select')) {
    // Detect current language from URL
    const path = window.location.pathname;
    let currentLang = 'ko'; // default
    if (path.startsWith('/en/')) currentLang = 'en';

    const languageNames = {
      'ko': '한국어',
      'en': 'English'
    };

    const languageIcons = {
      'ko': '🇰🇷', // Korean flag
      'en': '🇺🇸'  // US flag
    };

    const languages = [
      { name: '한국어', icon: '🇰🇷', link: '/', lang: 'ko' },
      { name: 'English', icon: '🇺🇸', link: '/en/', lang: 'en' }
    ];

    const selector = document.createElement('div');
    selector.className = 'md-select';

    let listHTML = '';
    languages.forEach(lang => {
      const activeClass = lang.lang === currentLang ? ' md-select__item--active' : '';
      listHTML += `<a href="${lang.link}" class="md-select__item${activeClass}" hreflang="${lang.lang}" title="${lang.name}">
        <span class="md-select__icon">${lang.icon}</span>
        <span class="md-select__label">${lang.name}</span>
      </a>`;
    });

    selector.innerHTML = `
      <button class="md-select__inner" aria-label="Select language" title="Language: ${languageNames[currentLang]}">
        <span class="md-select__current-icon">${languageIcons[currentLang]}</span>
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" width="16" height="16" class="md-select__arrow">
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

    // Handle dropdown toggle
    const button = selector.querySelector('.md-select__inner');
    button.addEventListener('click', function(e) {
      e.stopPropagation();
      selector.classList.toggle('md-select--active');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function() {
      selector.classList.remove('md-select--active');
    });

    // Prevent dropdown close when clicking inside
    selector.addEventListener('click', function(e) {
      e.stopPropagation();
    });
  }
});
