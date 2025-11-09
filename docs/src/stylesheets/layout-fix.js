// Layout Fix: Force 30:70 sidebar/content ratio for MoAI-ADK Documentation
(function() {
  'use strict';

  window.layoutFixApplied = false;

  function applyLayoutFix() {
    if (window.layoutFixApplied) return;

    const mainInner = document.querySelector('.md-main__inner');
    const sidebar = document.querySelector('.md-sidebar--primary');
    const main = document.querySelector('.md-main');

    if (!mainInner || !sidebar || !main) {
      // Elements not ready yet, retry
      setTimeout(applyLayoutFix, 100);
      return;
    }

    // Apply to sidebar (30% width)
    sidebar.setAttribute('style',
      'position: relative !important; ' +
      'top: auto !important; ' +
      'left: auto !important; ' +
      'width: 30% !important; ' +
      'flex: 0 0 30% !important; ' +
      'min-width: 30% !important; ' +
      'max-width: 30% !important; ' +
      'margin-right: 0 !important; ' +
      'margin-bottom: 0 !important;'
    );

    // Apply to main (70% width)
    main.setAttribute('style',
      'width: 70% !important; ' +
      'flex: 0 0 70% !important; ' +
      'min-width: 70% !important; ' +
      'max-width: 70% !important; ' +
      'margin: 0 !important;'
    );

    // Apply to mainInner (flex container)
    mainInner.setAttribute('style',
      'display: flex !important; ' +
      'flex-direction: row !important; ' +
      'gap: 0 !important; ' +
      'align-items: flex-start !important;'
    );

    window.layoutFixApplied = true;

    // Watch for Material re-rendering and reapply if needed
    const observer = new MutationObserver(function() {
      window.layoutFixApplied = false;
      applyLayoutFix();
    });

    observer.observe(mainInner, {
      attributes: true,
      childList: true,
      subtree: true,
      attributeFilter: ['style', 'class']
    });
  }

  // Apply immediately if DOM is ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', applyLayoutFix);
  } else {
    applyLayoutFix();
  }

  // Also ensure it's applied after delays to catch various timing issues
  setTimeout(applyLayoutFix, 50);
  setTimeout(applyLayoutFix, 200);
  setTimeout(applyLayoutFix, 500);
})();
