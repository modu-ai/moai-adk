---
name: publishing-guide
description: Chrome Web Store submission, review process, updates, and distribution
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Publishing Guide

## Overview

Publishing a Chrome extension involves packaging the extension, creating a Chrome Web Store developer account, submitting for review, and managing updates. This module covers the complete publishing workflow from preparation to ongoing maintenance.

## Developer Account Setup

To publish on the Chrome Web Store, you need a Google account registered as a Chrome Web Store developer. Registration requires a one-time fee of $5 USD. Visit the Chrome Web Store Developer Dashboard at https://chrome.google.com/webstore/devconsole to register.

After registration, verify your email address and agree to the developer distribution agreement. You can then create and manage extensions through the dashboard.

For organizations, consider using a shared Google account dedicated to extension management. This ensures continuity if individual team members leave.

## Pre-Submission Checklist

Before submitting, verify the following:

### Manifest Validation

Ensure manifest.json is complete and valid:

- manifest_version is set to 3
- name is descriptive and under 75 characters
- version follows the dot-separated format
- description is informative and under 132 characters
- icons include all required sizes (16, 32, 48, 128)
- All declared files (scripts, HTML, icons) exist in the package
- Permissions are minimized to only what is needed
- No deprecated APIs or MV2-specific fields

### Code Quality

- All JavaScript executes without errors in the console
- No references to external scripts or remote code
- CSP is properly configured
- All features work as expected across different Chrome versions
- Error handling is implemented for all API calls
- No hardcoded API keys or secrets in source code

### Store Listing Assets

Prepare the following assets before submission:

- Extension icon at 128x128 pixels (PNG format, displayed in store)
- At least one screenshot (1280x800 or 640x400 pixels)
- Optional: Promotional images (440x280 small tile, 920x680 large tile, 1400x560 marquee)
- Detailed description for the store listing
- Category selection (Productivity, Developer Tools, etc.)
- Language selection for the listing

## Packaging the Extension

### Creating the ZIP File

Package your extension as a ZIP file containing all source files. The manifest.json must be at the root level of the ZIP archive.

```bash
# From the extension root directory
zip -r extension.zip . -x ".*" -x "__MACOSX/*" -x "node_modules/*" -x "*.map" -x "src/*"
```

Exclude development files from the package:

- Source maps (.map files)
- Node modules
- Source TypeScript/SCSS files (if using a build step)
- Test files
- Configuration files (.eslintrc, tsconfig.json, etc.)
- Version control files (.git)
- Documentation for developers

### Build Pipeline

For extensions using TypeScript, React, or other build tools, create a build script that compiles source code, bundles dependencies, copies static assets, and creates the ZIP package.

```json
{
  "scripts": {
    "build": "vite build",
    "package": "npm run build && cd dist && zip -r ../extension.zip .",
    "validate": "npx chrome-extension-validate dist/manifest.json"
  }
}
```

Recommended build tools:

- Vite with @crxjs/vite-plugin for React/Vue/Svelte extensions
- webpack with webpack-extension-reloader for complex builds
- Rollup for lightweight extensions
- esbuild for fast TypeScript compilation

### Directory Structure for Distribution

```
dist/
  manifest.json
  service-worker.js
  popup/
    popup.html
    popup.js
    popup.css
  content/
    main.js
    styles.css
  options/
    options.html
    options.js
  sidepanel/
    sidepanel.html
    sidepanel.js
  icons/
    icon16.png
    icon32.png
    icon48.png
    icon128.png
  _locales/        (if using i18n)
    en/
      messages.json
    ko/
      messages.json
```

## Privacy Policy

A privacy policy is required for extensions that handle user data, which includes most non-trivial extensions. The policy must be accessible via a public URL.

### When a Privacy Policy is Required

- Extension uses network requests
- Extension accesses browsing history or tab URLs
- Extension stores user-generated content
- Extension uses authentication
- Extension communicates with external services
- Extension uses the "tabs" permission with URL access

### Privacy Policy Content

The privacy policy should address:

- What data the extension collects
- How the data is used
- Whether data is shared with third parties
- How data is stored and secured
- User rights regarding their data
- Contact information for privacy inquiries
- How the policy is updated

### Privacy Practices in Dashboard

The Chrome Web Store Developer Dashboard requires you to declare data practices:

- Whether you collect personally identifiable information
- Whether you collect health information
- Whether you collect financial information
- Whether you collect authentication information
- Whether you collect personal communications
- Whether you collect location data
- Whether you collect web browsing history
- Whether you sell data to third parties
- Whether data is used for purposes unrelated to extension functionality

Answer these questions accurately. Misrepresentation can lead to extension removal.

## Submission Process

### Uploading to the Dashboard

1. Navigate to the Chrome Web Store Developer Dashboard
2. Click "New Item" to create a new extension listing
3. Upload the ZIP file
4. Fill in the store listing details:
   - Detailed description (can include formatting)
   - Category
   - Language
   - Screenshots and promotional images
5. Configure distribution options:
   - Visibility (public, unlisted, or private)
   - Regions where the extension is available
   - Group publishing (for organization accounts)
6. Set the privacy tab:
   - Link to privacy policy URL
   - Declare data usage practices
   - Explain permission justifications
7. Submit for review

### Permission Justifications

For each permission declared in the manifest, you must provide a justification explaining why the extension needs it. Reviewers will verify these justifications. Be specific and accurate:

- "storage" - "Stores user preferences for theme and notification settings"
- "activeTab" - "Needs temporary access to the current tab to extract page content when the user clicks the toolbar button"
- "scripting" - "Injects content script to highlight search terms on the current page"

Vague justifications like "needed for functionality" will likely result in rejection.

### Review Process

After submission, the extension enters the review queue. Review times vary:

- New extensions: typically 1-3 business days
- Updates to existing extensions: typically 1-2 business days
- Extensions using sensitive permissions: may take longer

If the extension is rejected, the dashboard provides rejection reasons. Common rejection reasons include:

- Insufficient privacy policy
- Requesting unnecessary permissions
- Deceptive functionality (extension does not match description)
- Code quality issues (errors, deprecated APIs)
- Missing or inadequate permission justifications
- Violation of Chrome Web Store policies

Address the rejection reasons and resubmit. You can also appeal rejections through the dashboard.

## Managing Updates

### Automatic Updates via Chrome Web Store

When you upload a new version to the Chrome Web Store, Chrome automatically updates installed extensions. The update check typically occurs every few hours.

To publish an update:

1. Increment the version number in manifest.json
2. Build and package the new version
3. Upload to the Developer Dashboard
4. Submit for review (updates also require review)

### Version Numbering

Follow a consistent versioning strategy:

- Major version: Breaking changes or significant feature additions
- Minor version: New features, backward compatible
- Patch version: Bug fixes and minor improvements
- Build version (optional fourth number): Internal builds

```json
{
  "version": "2.1.3"
}
```

### Staged Rollouts

The Chrome Web Store supports staged rollouts where updates are gradually deployed to a percentage of users. This helps catch issues before they affect all users.

Configure rollout percentage in the Dashboard when publishing an update. Monitor feedback and crash reports during rollout before increasing to 100%.

### Handling Update Lifecycle

```javascript
// service-worker.js - Handle extension updates
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'update') {
    const previousVersion = details.previousVersion;
    const currentVersion = chrome.runtime.getManifest().version;

    console.log(`Updated from ${previousVersion} to ${currentVersion}`);

    // Perform data migrations
    migrateData(previousVersion, currentVersion);

    // Optionally notify users of new features
    if (shouldShowUpdateNotification(previousVersion)) {
      chrome.notifications.create('update-notification', {
        type: 'basic',
        iconUrl: 'icons/icon128.png',
        title: 'Extension Updated',
        message: `Updated to version ${currentVersion}. Click to see what is new.`
      });
    }
  }
});

async function migrateData(fromVersion, toVersion) {
  const { settings } = await chrome.storage.local.get('settings');
  if (!settings) return;

  // Example: Add new setting field introduced in v2.0
  if (compareVersions(fromVersion, '2.0.0') < 0) {
    settings.newFeatureEnabled = false;
    await chrome.storage.local.set({ settings });
  }
}
```

## Self-Hosted Distribution

For enterprise or private distribution, extensions can be self-hosted outside the Chrome Web Store.

### Packaging as CRX

Create a CRX file by going to chrome://extensions, enabling Developer mode, clicking "Pack extension", selecting the extension directory, and specifying a private key (generated on first pack).

### Update Manifest

Host an update manifest XML file on your server:

```xml
<?xml version='1.0' encoding='UTF-8'?>
<gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
  <app appid='YOUR_EXTENSION_ID'>
    <updatecheck crid='YOUR_EXTENSION_ID' version='1.0.1' prodversionmin='88.0'
      src='https://your-server.com/extension-1.0.1.crx' />
  </app>
</gupdate>
```

Reference the update URL in manifest.json:

```json
{
  "update_url": "https://your-server.com/updates.xml"
}
```

Note: Self-hosted extensions can only be installed in enterprise environments via policy or in developer mode. Regular users can only install from the Chrome Web Store.

## Internationalization

For global distribution, localize your extension using the Chrome i18n system.

### Messages Files

Create _locales directory with subdirectories for each language:

```json
// _locales/en/messages.json
{
  "extName": {
    "message": "My Extension",
    "description": "Extension display name"
  },
  "extDescription": {
    "message": "Enhances your browsing experience",
    "description": "Extension description for Web Store"
  },
  "popupTitle": {
    "message": "Dashboard",
    "description": "Title shown in popup header"
  }
}
```

```json
// _locales/ko/messages.json
{
  "extName": {
    "message": "My Extension",
    "description": "Extension display name"
  },
  "extDescription": {
    "message": "브라우징 경험을 향상시킵니다",
    "description": "Extension description for Web Store"
  },
  "popupTitle": {
    "message": "대시보드",
    "description": "Title shown in popup header"
  }
}
```

### Using i18n in Manifest

```json
{
  "name": "__MSG_extName__",
  "description": "__MSG_extDescription__",
  "default_locale": "en"
}
```

### Using i18n in JavaScript

```javascript
const title = chrome.i18n.getMessage('popupTitle');
document.getElementById('title').textContent = title;
```

### Using i18n in HTML

```html
<span data-i18n="popupTitle"></span>
```

```javascript
// Auto-translate elements with data-i18n attribute
document.querySelectorAll('[data-i18n]').forEach(el => {
  const key = el.getAttribute('data-i18n');
  el.textContent = chrome.i18n.getMessage(key);
});
```

## Analytics and Monitoring

### Chrome Web Store Statistics

The Developer Dashboard provides:

- Install and uninstall counts
- Active user counts
- Rating and review summaries
- Geographic distribution
- Chrome version distribution

### Extension Error Monitoring

Monitor extension errors through the Chrome Developer Dashboard error reports. You can also implement custom error tracking:

```javascript
// service-worker.js - Global error handler
self.addEventListener('error', (event) => {
  reportError({
    type: 'uncaught',
    message: event.message,
    filename: event.filename,
    line: event.lineno,
    col: event.colno
  });
});

self.addEventListener('unhandledrejection', (event) => {
  reportError({
    type: 'unhandled-rejection',
    message: event.reason?.message || String(event.reason)
  });
});

async function reportError(errorData) {
  // Store errors locally
  const { errors = [] } = await chrome.storage.local.get('errors');
  errors.push({ ...errorData, timestamp: Date.now() });
  // Keep only last 100 errors
  if (errors.length > 100) errors.splice(0, errors.length - 100);
  await chrome.storage.local.set({ errors });
}
```

## Policy Compliance

Ensure ongoing compliance with Chrome Web Store Developer Program Policies:

- Extensions must have a single, clear purpose
- Do not collect data beyond what is necessary
- Do not use deceptive installation tactics
- Provide clear uninstall behavior
- Do not modify browser settings without user consent
- Do not inject advertisements unless that is the declared purpose
- Keep permissions minimal and justified
- Respond promptly to policy violation notices
