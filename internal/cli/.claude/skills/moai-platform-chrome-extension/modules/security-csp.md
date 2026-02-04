---
name: security-csp
description: Content Security Policy, permissions model, input validation, and secure coding
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Security and Content Security Policy

## Overview

Manifest V3 significantly strengthens the security model for Chrome extensions. Remote code execution is prohibited, Content Security Policy enforcement is stricter, and the permissions model provides granular control over API and website access. This module covers all security aspects of Chrome extension development.

## Content Security Policy

### MV3 Default Policy

Manifest V3 enforces a baseline CSP that cannot be relaxed. The default policy is:

script-src 'self' 'wasm-unsafe-eval'

This means:

- Only scripts bundled within the extension package can execute
- No inline scripts (event handlers in HTML attributes, script tags with content)
- No eval(), new Function(), or other dynamic code evaluation
- No remote scripts loaded from CDNs or external URLs
- WebAssembly modules are permitted with wasm-unsafe-eval

### Custom CSP Configuration

Extensions can only make the CSP more restrictive, not less. Configure via manifest.json:

```json
{
  "content_security_policy": {
    "extension_pages": "script-src 'self' 'wasm-unsafe-eval'; object-src 'self'; style-src 'self'"
  }
}
```

The extension_pages policy applies to popup, options page, side panel, and other extension HTML pages. Content scripts are not affected by extension CSP because they execute in the web page context.

### CSP Implications for Development

All JavaScript must be in separate .js files referenced by script tags with src attributes. Move inline event handlers to addEventListener calls in JavaScript files. Move inline styles to external CSS files or use JavaScript DOM manipulation.

Template engines that generate code at runtime (such as eval-based templating) cannot be used. Choose template libraries that support pre-compilation or DOM-based rendering.

```html
<!-- WRONG: Inline script -->
<button onclick="handleClick()">Click</button>
<script>var x = 1;</script>

<!-- CORRECT: External script with addEventListener -->
<button id="my-btn">Click</button>
<script src="popup.js"></script>
```

```javascript
// popup.js
document.getElementById('my-btn').addEventListener('click', handleClick);

function handleClick() {
  // Handle the click
}
```

### Remote Code Prohibition

Extensions cannot load or execute code from external sources. All of the following are prohibited:

- Loading scripts from CDN URLs in HTML
- Fetching JavaScript and executing with eval or Function
- Creating script elements with external src attributes
- Using document.write to inject external scripts
- Loading remote worker scripts

All code must be bundled within the extension package. Use build tools like webpack, Vite, or Rollup to bundle dependencies into the extension.

## Permissions Model

### Principle of Minimum Privilege

Request only the permissions your extension actively needs. Each permission increases the attack surface and may trigger additional install warnings. Users are more likely to install extensions with fewer permissions.

### Permission Categories Explained

Standard permissions grant access to specific Chrome APIs. Each permission unlocks a set of API methods and has specific privacy implications:

"activeTab" grants temporary access to the active tab when the user invokes the extension through a user gesture (clicking the toolbar button, using a keyboard shortcut, or selecting a context menu item). This is the most privacy-friendly permission for tab interaction because it requires explicit user intent and access expires when the tab navigates.

"tabs" grants access to the url, title, and favIconUrl properties of Tab objects. Without this permission, these fields are undefined. Basic tab operations like query, create, and update do not require the "tabs" permission.

"storage" grants access to chrome.storage API. This is one of the most commonly needed permissions with minimal privacy impact.

"scripting" grants access to chrome.scripting for programmatic script injection. Also requires host_permissions or activeTab for the target pages.

"notifications" allows displaying system notifications. Users can disable notifications for specific extensions.

"contextMenus" allows creating right-click context menu items.

"alarms" allows scheduling recurring events.

### Host Permissions Strategy

Host permissions control which websites the extension can interact with. They trigger prominent install warnings and affect the extension's access to web page content.

```json
{
  "host_permissions": ["https://specific-api.example.com/*"]
}
```

Prefer specific host patterns over broad patterns. Use activeTab instead of broad host_permissions when the extension only needs access on user gesture. Avoid `<all_urls>` unless absolutely necessary -- it triggers the strongest install warning.

### Optional Permissions

Request permissions at runtime when the user first needs the feature. This reduces the install-time warning and allows the extension to function with a minimal initial permission set.

```javascript
// Request optional permission when user enables a feature
async function enableBookmarksFeature() {
  const granted = await chrome.permissions.request({
    permissions: ['bookmarks']
  });

  if (granted) {
    // Permission granted, initialize feature
    await initializeBookmarksSync();
  } else {
    // Permission denied, show message
    showMessage('Bookmarks access is required for this feature');
  }
}

// Check before using optional features
async function checkPermission(permission) {
  return chrome.permissions.contains({ permissions: [permission] });
}

// Remove permissions when feature is disabled
async function disableBookmarksFeature() {
  await chrome.permissions.remove({ permissions: ['bookmarks'] });
}
```

```json
{
  "permissions": ["storage", "activeTab"],
  "optional_permissions": ["bookmarks", "downloads", "history"],
  "optional_host_permissions": ["https://*.example.com/*"]
}
```

## Input Validation

### Message Validation

Always validate messages received from content scripts or external sources. Content scripts operate in the context of web pages that may be malicious. Never assume message data is well-formed or trustworthy.

```javascript
// service-worker.js - Validate incoming messages

// Define expected message schemas
const MESSAGE_SCHEMAS = {
  'save-data': {
    required: ['key', 'value'],
    types: { key: 'string', value: 'object' },
    validate: (msg) => msg.key.length <= 100 && msg.key.match(/^[a-zA-Z0-9_-]+$/)
  },
  'fetch-url': {
    required: ['url'],
    types: { url: 'string' },
    validate: (msg) => {
      try {
        const url = new URL(msg.url);
        return url.protocol === 'https:';
      } catch { return false; }
    }
  }
};

function validateMessage(message) {
  if (!message || typeof message !== 'object') return false;
  if (typeof message.action !== 'string') return false;

  const schema = MESSAGE_SCHEMAS[message.action];
  if (!schema) return false;

  // Check required fields
  for (const field of schema.required) {
    if (!(field in message)) return false;
  }

  // Check types
  for (const [field, expectedType] of Object.entries(schema.types)) {
    if (typeof message[field] !== expectedType) return false;
  }

  // Custom validation
  if (schema.validate && !schema.validate(message)) return false;

  return true;
}

chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (!validateMessage(message)) {
    sendResponse({ error: 'Invalid message format' });
    return false;
  }

  // Process validated message
  handleValidatedMessage(message, sender, sendResponse);
  return true;
});
```

### URL Validation

Always validate URLs before navigation, fetching, or injection to prevent open redirect and SSRF attacks.

```javascript
// Validate URLs against allowlists
const ALLOWED_DOMAINS = ['api.example.com', 'cdn.example.com'];

function isAllowedURL(urlString) {
  try {
    const url = new URL(urlString);
    if (url.protocol !== 'https:') return false;
    return ALLOWED_DOMAINS.some(domain =>
      url.hostname === domain || url.hostname.endsWith(`.${domain}`)
    );
  } catch {
    return false;
  }
}

// Safe navigation
async function openURL(urlString) {
  if (!isAllowedURL(urlString)) {
    console.error('Blocked navigation to:', urlString);
    return;
  }
  await chrome.tabs.create({ url: urlString });
}
```

### Data Sanitization

Sanitize all data before inserting into the DOM, whether in extension pages or content scripts.

```javascript
// Safe text insertion - never use innerHTML with untrusted data
function safeSetText(element, text) {
  element.textContent = text;
}

// Safe element creation
function createSafeElement(tag, attributes, textContent) {
  const el = document.createElement(tag);
  for (const [key, value] of Object.entries(attributes)) {
    // Only allow safe attributes
    const safeAttrs = ['class', 'id', 'data-', 'aria-', 'role', 'type', 'name', 'value', 'placeholder'];
    if (safeAttrs.some(safe => key === safe || key.startsWith('data-') || key.startsWith('aria-'))) {
      el.setAttribute(key, String(value));
    }
  }
  if (textContent) {
    el.textContent = textContent;
  }
  return el;
}

// Sanitize HTML if absolutely necessary (prefer textContent)
function sanitizeForDisplay(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML; // HTML entities are escaped
}
```

## XSS Prevention

### Extension Pages

In extension popup, options, and side panel pages:

- Never use innerHTML, outerHTML, or insertAdjacentHTML with untrusted data
- Never use document.write or document.writeln
- Always use textContent for text display
- Use createElement and setAttribute for DOM construction
- Avoid template literals inserted into DOM without escaping

### Content Scripts

In content scripts injected into web pages:

- Assume all page DOM content is untrusted
- Never extract and evaluate JavaScript from page elements
- Use Shadow DOM for injected UI to prevent page CSS interference
- Validate all data before sending to the service worker
- Clean up injected elements on extension update or uninstall

```javascript
// UNSAFE patterns - never do these
element.innerHTML = userData;                    // XSS
element.insertAdjacentHTML('beforeend', data);  // XSS
document.write(content);                        // XSS
eval(messageData);                              // Code injection
new Function(dynamicCode)();                    // Code injection
setTimeout(stringCode, 0);                      // Code injection

// SAFE patterns
element.textContent = userData;                 // Safe
element.setAttribute('title', userData);        // Safe for most attributes
const el = document.createElement('span');      // Safe DOM construction
el.textContent = userData;
element.appendChild(el);
```

## Secure Communication

### Between Extension Components

Communication between popup, options, side panel, and service worker is within the trusted extension context. However, still validate data structures to prevent bugs and ensure defensive coding.

### With Content Scripts

Content scripts run in web page environments that may be compromised. The host page can manipulate shared DOM elements, intercept events, and attempt to exploit the extension through the shared DOM surface.

Best practices:

- Validate all messages received from content scripts in the service worker
- Do not send sensitive data (API keys, tokens) to content scripts
- Use unique, hard-to-guess channel names for ports
- Rate-limit operations triggered by content script messages
- Do not expose extension functionality that content scripts do not need

### With External Websites

When using externally_connectable to receive messages from web pages:

- Strictly validate sender origin against your allowlist
- Never trust the content of external messages
- Implement rate limiting for external message handlers
- Log and monitor external message patterns for abuse detection

### With External APIs

- Always use HTTPS for external requests
- Store API keys in chrome.storage, not in source code
- Implement request signing or token-based authentication
- Validate and sanitize API responses before use
- Handle network errors and timeouts gracefully

## Web Accessible Resources Security

Files declared in web_accessible_resources are accessible from web pages, which means malicious pages can probe for your extension's presence by requesting these resources.

Minimize the number of web-accessible resources. Scope access to specific origins using matches. Use use_dynamic_url when possible to generate unique URLs that change on each session, preventing fingerprinting.

```json
{
  "web_accessible_resources": [
    {
      "resources": ["injected/styles.css"],
      "matches": ["https://*.example.com/*"],
      "use_dynamic_url": true
    }
  ]
}
```

## Security Checklist

Before publishing, verify:

- All JavaScript is bundled within the extension package
- No eval, Function, or other dynamic code evaluation
- No inline scripts or event handlers in HTML
- CSP is configured in manifest.json
- Permissions are minimized (activeTab preferred over broad host access)
- All messages from content scripts are validated
- All external URLs are validated against allowlists
- No sensitive data stored in content script accessible locations
- web_accessible_resources are scoped to necessary origins
- All external API calls use HTTPS
- Error handling does not leak sensitive information
- Optional permissions are used for non-essential features
