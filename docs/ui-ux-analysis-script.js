/**
 * Comprehensive UI/UX Analysis Script for MoAI-ADK Documentation
 * Analyzes: layout, Nextra theme, styling, responsive design, and overall quality
 */

const { chromium } = require('playwright');
const fs = require('fs');
const path = require('path');

class UIAnalyzer {
  constructor() {
    this.results = {
      layout: {},
      theme: {},
      styling: {},
      responsive: {},
      quality: {},
      screenshots: {},
      htmlStructure: {}
    };
  }

  async analyze(url = 'http://localhost:3000') {
    console.log('🎨 Starting comprehensive UI/UX analysis...');

    const browser = await chromium.launch();
    const context = await browser.newContext();
    const page = await context.newPage();

    try {
      // Wait for page load
      console.log('📄 Loading documentation page...');
      await page.goto(url, {
        waitUntil: 'networkidle',
        timeout: 30000
      });

      // Give extra time for dynamic content to load
      await page.waitForTimeout(3000);

      // 1. Capture screenshots at different viewport sizes
      await this.captureScreenshots(page);

      // 2. Analyze layout structure
      await this.analyzeLayout(page);

      // 3. Check Nextra theme application
      await this.checkNextraTheme(page);

      // 4. Analyze styling and fonts
      await this.analyzeStyling(page);

      // 5. Test responsive design
      await this.testResponsiveDesign(context);

      // 6. Analyze HTML structure
      await this.analyzeHTMLStructure(page);

      // 7. Evaluate design quality
      await this.evaluateDesignQuality(page);

      console.log('✅ Analysis completed successfully!');
      return this.results;

    } catch (error) {
      console.error('❌ Analysis failed:', error);
      throw error;
    } finally {
      await browser.close();
    }
  }

  async captureScreenshots(page) {
    console.log('📸 Capturing screenshots...');

    // Desktop view
    await page.setViewportSize({ width: 1920, height: 1080 });
    await this.takeScreenshot(page, 'desktop-full', { fullPage: true });

    // Tablet view
    await page.setViewportSize({ width: 768, height: 1024 });
    await this.takeScreenshot(page, 'tablet-full', { fullPage: true });

    // Mobile view
    await page.setViewportSize({ width: 375, height: 667 });
    await this.takeScreenshot(page, 'mobile-full', { fullPage: true });

    // Above the fold shots
    await page.setViewportSize({ width: 1920, height: 1080 });
    await this.takeScreenshot(page, 'desktop-above-fold');
  }

  async takeScreenshot(page, name, options = {}) {
    const screenshotPath = `screenshot-${name}.png`;
    await page.screenshot({
      path: screenshotPath,
      ...options
    });
    this.results.screenshots[name] = screenshotPath;
    console.log(`📷 Screenshot saved: ${screenshotPath}`);
  }

  async analyzeLayout(page) {
    console.log('🏗️ Analyzing layout structure...');

    // Check for key layout elements
    const layoutElements = await page.evaluate(() => {
      const elements = {};

      // Header
      elements.header = {
        exists: !!document.querySelector('header'),
        element: document.querySelector('header')?.outerHTML.substring(0, 200) + '...'
      };

      // Navigation
      elements.nav = {
        exists: !!document.querySelector('nav'),
        count: document.querySelectorAll('nav').length,
        primary: !!document.querySelector('.nextra-nav-container') || !!document.querySelector('nav[role="navigation"]')
      };

      // Sidebar
      elements.sidebar = {
        exists: !!document.querySelector('aside'),
        nextra: !!document.querySelector('.nextra-sidebar-container'),
        count: document.querySelectorAll('aside').length
      };

      // Main content
      elements.main = {
        exists: !!document.querySelector('main'),
        nextra: !!document.querySelector('.nextra-container'),
        hasContent: !!document.querySelector('main article') || !!document.querySelector('.nextra-content')
      };

      // Footer
      elements.footer = {
        exists: !!document.querySelector('footer'),
        nextra: !!document.querySelector('.nextra-footer')
      };

      // Content wrapper
      elements.wrapper = {
        nextraContainer: !!document.querySelector('.nextra-container'),
        contentWrapper: !!document.querySelector('.nextra-content'),
        markdownContent: !!document.querySelector('article, .markdown, .nextra-body')
      };

      return elements;
    });

    this.results.layout = layoutElements;
  }

  async checkNextraTheme(page) {
    console.log('🎨 Checking Nextra theme application...');

    const themeAnalysis = await page.evaluate(() => {
      const analysis = {
        stylesheets: [],
        themeClasses: [],
        nextraElements: {},
        customizations: {},
        issues: []
      };

      // Check for Nextra CSS
      document.querySelectorAll('link[rel="stylesheet"]').forEach(link => {
        if (link.href.includes('nextra') || link.href.includes('theme')) {
          analysis.stylesheets.push(link.href);
        }
      });

      // Check for Nextra CSS classes
      const nextraClasses = [
        'nextra-container', 'nextra-sidebar-container', 'nextra-nav-container',
        'nextra-content', 'nextra-footer', 'nextra-body', 'nextra-search',
        'nx-theme-document', 'nx-layout', 'nx-page'
      ];

      nextraClasses.forEach(className => {
        if (document.querySelector(`.${className}`)) {
          analysis.themeClasses.push(className);
        }
      });

      // Check for key Nextra elements
      analysis.nextraElements = {
        search: !!document.querySelector('.nextra-search, [type="search"], .search'),
        themeToggle: !!document.querySelector('.nextra-theme-toggle, .dark-mode-toggle, [data-theme-toggle]'),
        breadcrumb: !!document.querySelector('.nextra-breadcrumb, .breadcrumb'),
        toc: !!document.querySelector('.nextra-toc, .table-of-contents'),
        editLink: !!document.querySelector('.nextra-edit-link, [href*="edit"]'),
        pagination: !!document.querySelector('.nextra-pagination, .pagination')
      };

      // Check for common issues
      if (analysis.stylesheets.length === 0) {
        analysis.issues.push('No Nextra stylesheets detected');
      }
      if (analysis.themeClasses.length < 3) {
        analysis.issues.push('Missing core Nextra theme classes');
      }
      if (!analysis.nextraElements.container && !document.querySelector('.nextra-container')) {
        analysis.issues.push('Nextra container not found');
      }

      return analysis;
    });

    this.results.theme = themeAnalysis;
  }

  async analyzeStyling(page) {
    console.log('🎨 Analyzing styling and fonts...');

    const stylingAnalysis = await page.evaluate(() => {
      const analysis = {
        fonts: [],
        colors: {},
        spacing: {},
        typography: {},
        issues: []
      };

      // Analyze loaded fonts
      document.fonts.forEach(font => {
        analysis.fonts.push({
          family: font.family,
          weight: font.weight,
          style: font.style
        });
      });

      // Check common font families
      const computedBody = getComputedStyle(document.body);
      analysis.typography = {
        fontFamily: computedBody.fontFamily,
        fontSize: computedBody.fontSize,
        lineHeight: computedBody.lineHeight,
        color: computedBody.color,
        backgroundColor: computedBody.backgroundColor
      };

      // Check for CSS variables (design tokens)
      const rootStyles = getComputedStyle(document.documentElement);
      const cssVariables = {};
      for (let i = 0; i < rootStyles.length; i++) {
        const property = rootStyles[i];
        if (property.startsWith('--')) {
          cssVariables[property] = rootStyles.getPropertyValue(property);
        }
      }
      analysis.cssVariables = cssVariables;

      // Check for styling issues
      if (analysis.fonts.length === 0) {
        analysis.issues.push('No custom fonts loaded');
      }
      if (!analysis.typography.fontFamily.includes('Inter') && !analysis.typography.fontFamily.includes('system-ui')) {
        analysis.issues.push('Expected Inter or system-ui font not found');
      }

      return analysis;
    });

    this.results.styling = stylingAnalysis;
  }

  async testResponsiveDesign(context) {
    console.log('📱 Testing responsive design...');

    const viewports = [
      { name: 'Mobile', width: 375, height: 667 },
      { name: 'Tablet', width: 768, height: 1024 },
      { name: 'Desktop', width: 1920, height: 1080 }
    ];

    const responsiveResults = {};

    for (const viewport of viewports) {
      const page = await context.newPage();
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.goto('http://localhost:3000', { waitUntil: 'networkidle' });
      await page.waitForTimeout(2000);

      const viewportAnalysis = await page.evaluate((vp) => {
        const analysis = {
          sidebarVisible: false,
          navigationVisible: false,
          contentWidth: 0,
          horizontalScroll: false,
          elementsOverflow: []
        };

        // Check sidebar visibility
        const sidebar = document.querySelector('aside, .nextra-sidebar-container');
        if (sidebar) {
          const styles = getComputedStyle(sidebar);
          analysis.sidebarVisible = styles.display !== 'none' && styles.visibility !== 'hidden';
        }

        // Check navigation
        const nav = document.querySelector('nav, .nextra-nav-container');
        if (nav) {
          const styles = getComputedStyle(nav);
          analysis.navigationVisible = styles.display !== 'none';
        }

        // Check content width
        const content = document.querySelector('main, .nextra-content');
        if (content) {
          analysis.contentWidth = content.offsetWidth;
        }

        // Check for horizontal scroll
        analysis.horizontalScroll = document.body.scrollWidth > window.innerWidth;

        return analysis;
      }, { width: viewport.width, height: viewport.height });

      responsiveResults[viewport.name] = {
        viewport,
        ...viewportAnalysis
      };

      await page.close();
    }

    this.results.responsive = responsiveResults;
  }

  async analyzeHTMLStructure(page) {
    console.log('🏗️ Analyzing HTML structure...');

    const htmlAnalysis = await page.evaluate(() => {
      const analysis = {
        doctype: document.doctype?.name || 'missing',
        htmlLang: document.documentElement.lang || 'missing',
        headElements: {},
        bodyStructure: {},
        semanticHTML: {},
        issues: []
      };

      // Analyze head elements
      analysis.headElements = {
        title: document.title,
        metaDescription: document.querySelector('meta[name="description"]')?.content || 'missing',
        viewport: document.querySelector('meta[name="viewport"]')?.content || 'missing',
        favicon: !!document.querySelector('link[rel*="icon"]'),
        stylesheets: document.querySelectorAll('link[rel="stylesheet"]').length,
        scripts: document.querySelectorAll('script').length
      };

      // Analyze body structure
      analysis.bodyStructure = {
        hasHeader: !!document.querySelector('header'),
        hasNav: !!document.querySelector('nav'),
        hasMain: !!document.querySelector('main'),
        hasAside: !!document.querySelector('aside'),
        hasFooter: !!document.querySelector('footer'),
        hasArticle: !!document.querySelector('article'),
        totalElements: document.querySelectorAll('*').length
      };

      // Check semantic HTML
      analysis.semanticHTML = {
        properHeadingHierarchy: checkHeadingHierarchy(),
        usesLandmarks: checkLandmarks(),
        hasAltText: checkAltText(),
        properForms: checkForms()
      };

      // Check for issues
      if (!document.doctype) {
        analysis.issues.push('Missing DOCTYPE declaration');
      }
      if (!document.documentElement.lang) {
        analysis.issues.push('Missing html lang attribute');
      }
      if (!analysis.headElements.title) {
        analysis.issues.push('Missing page title');
      }
      if (!analysis.headElements.viewport) {
        analysis.issues.push('Missing viewport meta tag');
      }

      return analysis;

      // Helper functions
      function checkHeadingHierarchy() {
        const headings = document.querySelectorAll('h1, h2, h3, h4, h5, h6');
        if (headings.length === 0) return false;

        let lastLevel = 0;
        for (const heading of headings) {
          const level = parseInt(heading.tagName.substring(1));
          if (level > lastLevel + 1 && lastLevel !== 0) {
            return false;
          }
          lastLevel = level;
        }
        return true;
      }

      function checkLandmarks() {
        const landmarks = ['header', 'nav', 'main', 'aside', 'footer'];
        return landmarks.some(landmark => document.querySelector(landmark));
      }

      function checkAltText() {
        const images = document.querySelectorAll('img');
        return Array.from(images).every(img => img.alt || img.role === 'presentation');
      }

      function checkForms() {
        const forms = document.querySelectorAll('form');
        return Array.from(forms).every(form => {
          const inputs = form.querySelectorAll('input, textarea, select');
          return Array.from(inputs).every(input =>
            input.id && document.querySelector(`label[for="${input.id}"]`)
          );
        });
      }
    });

    this.results.htmlStructure = htmlAnalysis;
  }

  async evaluateDesignQuality(page) {
    console.log('⭐ Evaluating design quality...');

    const qualityAnalysis = await page.evaluate(() => {
      const analysis = {
        score: 0,
        strengths: [],
        weaknesses: [],
        recommendations: [],
        checks: {}
      };

      // Visual hierarchy
      analysis.checks.visualHierarchy = {
        hasClearHeadings: document.querySelectorAll('h1, h2, h3').length > 0,
        consistentSpacing: checkConsistentSpacing(),
        goodContrast: checkColorContrast()
      };

      // Content organization
      analysis.checks.contentOrganization = {
        hasNavigation: !!document.querySelector('nav'),
        hasSearch: !!document.querySelector('[type="search"], .search'),
        hasTableOfContents: !!document.querySelector('.toc, .table-of-contents'),
        hasBreadcrumb: !!document.querySelector('.breadcrumb')
      };

      // User experience
      analysis.checks.userExperience = {
        responsiveDesign: window.innerWidth >= 320, // Basic check
        fastLoad: performance.now() < 3000,
        accessible: checkAccessibility()
      };

      // Calculate score
      let score = 0;
      let totalChecks = 0;

      Object.values(analysis.checks).forEach(category => {
        Object.values(category).forEach(check => {
          totalChecks++;
          if (check) score++;
        });
      });

      analysis.score = Math.round((score / totalChecks) * 100);

      // Generate recommendations
      if (analysis.score < 70) {
        analysis.recommendations.push('Overall design quality needs significant improvement');
      }
      if (!analysis.checks.visualHierarchy.hasClearHeadings) {
        analysis.recommendations.push('Add clear heading hierarchy for better content structure');
      }
      if (!analysis.checks.contentOrganization.hasNavigation) {
        analysis.recommendations.push('Implement proper navigation structure');
      }
      if (!analysis.checks.userExperience.accessible) {
        analysis.recommendations.push('Improve accessibility features');
      }

      return analysis;

      // Helper functions
      function checkConsistentSpacing() {
        // Simple check for consistent spacing
        const elements = document.querySelectorAll('p, h1, h2, h3, div');
        return elements.length > 10; // Assume sufficient content means some organization
      }

      function checkColorContrast() {
        // Basic contrast check
        const body = document.body;
        const computed = getComputedStyle(body);
        const textColor = computed.color;
        const bgColor = computed.backgroundColor;

        // Simple heuristic - if colors are specified, assume decent contrast
        return textColor !== 'rgb(0, 0, 0)' || bgColor !== 'rgb(255, 255, 255)';
      }

      function checkAccessibility() {
        const hasAriaLabels = document.querySelectorAll('[aria-label], [aria-labelledby]').length > 0;
        const hasSemanticHTML = document.querySelectorAll('main, header, nav, footer').length > 0;
        const hasAltText = Array.from(document.querySelectorAll('img')).every(img => img.alt);

        return hasAriaLabels || hasSemanticHTML || hasAltText;
      }
    });

    this.results.quality = qualityAnalysis;
  }

  generateReport() {
    console.log('\n' + '='.repeat(80));
    console.log('🎨 MoAI-ADK Documentation UI/UX Analysis Report');
    console.log('='.repeat(80));

    // Executive Summary
    console.log('\n📊 EXECUTIVE SUMMARY');
    console.log('─'.repeat(40));
    console.log(`Overall Quality Score: ${this.results.quality.score}/100`);
    console.log(`Total Screenshots: ${Object.keys(this.results.screenshots).length}`);

    if (this.results.quality.recommendations.length > 0) {
      console.log('\n🔴 Critical Issues:');
      this.results.quality.recommendations.forEach(rec =>
        console.log(`  • ${rec}`)
      );
    }

    // Layout Analysis
    console.log('\n🏗️ LAYOUT STRUCTURE ANALYSIS');
    console.log('─'.repeat(40));
    console.log(`Header: ${this.results.layout.header?.exists ? '✅ Found' : '❌ Missing'}`);
    console.log(`Navigation: ${this.results.layout.nav?.exists ? '✅ Found' : '❌ Missing'}`);
    console.log(`Sidebar: ${this.results.layout.sidebar?.exists ? '✅ Found' : '❌ Missing'}`);
    console.log(`Main Content: ${this.results.layout.main?.exists ? '✅ Found' : '❌ Missing'}`);
    console.log(`Footer: ${this.results.layout.footer?.exists ? '✅ Found' : '❌ Missing'}`);

    // Nextra Theme Status
    console.log('\n🎨 NEXTRA THEME STATUS');
    console.log('─'.repeat(40));
    console.log(`Stylesheets Found: ${this.results.theme.stylesheets.length}`);
    console.log(`Theme Classes: ${this.results.theme.themeClasses.length}`);

    if (this.results.theme.issues.length > 0) {
      console.log('\n🔴 Theme Issues:');
      this.results.theme.issues.forEach(issue =>
        console.log(`  • ${issue}`)
      );
    }

    // Styling Analysis
    console.log('\n🎨 STYLING ANALYSIS');
    console.log('─'.repeat(40));
    console.log(`Font Family: ${this.results.styling.typography?.fontFamily || 'Not detected'}`);
    console.log(`Font Size: ${this.results.styling.typography?.fontSize || 'Not detected'}`);
    console.log(`CSS Variables: ${Object.keys(this.results.styling.cssVariables || {}).length}`);

    if (this.results.styling.issues.length > 0) {
      console.log('\n🔴 Styling Issues:');
      this.results.styling.issues.forEach(issue =>
        console.log(`  • ${issue}`)
      );
    }

    // Responsive Design
    console.log('\n📱 RESPONSIVE DESIGN ANALYSIS');
    console.log('─'.repeat(40));
    Object.entries(this.results.responsive).forEach(([device, analysis]) => {
      console.log(`\n${device} (${analysis.viewport.width}x${analysis.viewport.height}):`);
      console.log(`  Sidebar: ${analysis.sidebarVisible ? '✅ Visible' : '❌ Hidden'}`);
      console.log(`  Navigation: ${analysis.navigationVisible ? '✅ Visible' : '❌ Hidden'}`);
      console.log(`  Content Width: ${analysis.contentWidth}px`);
      console.log(`  Horizontal Scroll: ${analysis.horizontalScroll ? '❌ Yes' : '✅ No'}`);
    });

    // HTML Structure
    console.log('\n🏗️ HTML STRUCTURE ANALYSIS');
    console.log('─'.repeat(40));
    console.log(`DOCTYPE: ${this.results.htmlStructure.doctype}`);
    console.log(`HTML Lang: ${this.results.htmlStructure.htmlLang}`);
    console.log(`Page Title: ${this.results.htmlStructure.headElements?.title || 'Missing'}`);
    console.log(`Meta Description: ${this.results.htmlStructure.headElements?.metaDescription}`);

    if (this.results.htmlStructure.issues.length > 0) {
      console.log('\n🔴 HTML Issues:');
      this.results.htmlStructure.issues.forEach(issue =>
        console.log(`  • ${issue}`)
      );
    }

    // Screenshots
    console.log('\n📸 SCREENSHOTS CAPTURED');
    console.log('─'.repeat(40));
    Object.entries(this.results.screenshots).forEach(([name, path]) => {
      console.log(`${name}: ${path}`);
    });

    // Recommendations
    console.log('\n💡 RECOMMENDATIONS');
    console.log('─'.repeat(40));
    const allRecommendations = [
      ...this.results.quality.recommendations,
      ...this.results.theme.issues,
      ...this.results.styling.issues,
      ...this.results.htmlStructure.issues
    ];

    const uniqueRecommendations = [...new Set(allRecommendations)];
    if (uniqueRecommendations.length > 0) {
      uniqueRecommendations.forEach(rec =>
        console.log(`• ${rec}`)
      );
    } else {
      console.log('✅ No major issues detected!');
    }

    console.log('\n' + '='.repeat(80));

    return this.results;
  }
}

// Run analysis if script is executed directly
if (require.main === module) {
  const analyzer = new UIAnalyzer();

  analyzer.analyze()
    .then(results => {
      analyzer.generateReport();

      // Save detailed results to JSON file
      fs.writeFileSync(
        'ui-analysis-report.json',
        JSON.stringify(results, null, 2)
      );
      console.log('\n📄 Detailed report saved to: ui-analysis-report.json');
    })
    .catch(error => {
      console.error('❌ Analysis failed:', error);
      process.exit(1);
    });
}

module.exports = UIAnalyzer;