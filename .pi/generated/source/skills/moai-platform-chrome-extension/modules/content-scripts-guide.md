---
name: content-scripts-guide
description: Content script injection methods, isolated worlds, DOM access, and security
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Content Scripts Guide

## Overview

Content scripts are JavaScript and CSS files that run in the context of web pages. They can read and modify the DOM of visited pages, enabling extensions to interact with web content. Content scripts execute in isolated worlds, which means they share DOM access with the host page but have completely separate JavaScript execution environments.

## Isolated World Architecture

Chrome creates a separate JavaScript execution environment called an isolated world for each content script. This isolation means:

The content script and the host page cannot access each other's JavaScript variables, functions, or objects. The content script's window object is distinct from the page's window object. Both the content script and the page can read and modify the same DOM tree. CSS modifications from the content script apply to the page normally.

This isolation prevents naming conflicts between extension code and page code, protects extension logic from malicious page scripts, and prevents the extension from accidentally breaking page functionality.

However, the shared DOM means the host page can observe DOM changes made by the content script, and vice versa. Do not store sensitive data in the DOM.

## Static Declaration

Static content scripts are declared in manifest.json and injected automatically on matching pages. This is the simplest method and requires no special permissions beyond the URL match patterns.

```json
{
  "content_scripts": [
    {
      "matches": ["https://*.example.com/*", "https://docs.example.org/*"],
      "js": ["content/setup.js", "content/main.js"],
      "css": ["content/styles.css"],
      "run_at": "document_end",
      "all_frames": false,
      "match_about_blank": false
    },
    {
      "matches": ["<all_urls>"],
      "js": ["content/universal.js"],
      "run_at": "document_idle",
      "exclude_matches": ["https://admin.example.com/*"]
    }
  ]
}
```

Match Pattern Syntax:

Patterns follow the format scheme://host/path where:

- scheme can be http, https, or the wildcard * matching both
- host can be a specific domain, a wildcard subdomain like *.example.com, or * for all hosts
- path uses * as wildcard matching any path segment

Special patterns:

- `<all_urls>` matches any URL with http or https scheme
- exclude_matches removes specific URLs from the matched set
- include_globs and exclude_globs provide additional filtering with glob syntax

Injection Timing with run_at:

document_start: Script runs before the DOM is constructed. Use this when you need to intercept resources, inject early CSS, or block page scripts. The document object exists but has no content yet.

document_end: Script runs after the DOM is fully constructed but before images and subframes finish loading. This is the most common choice for DOM manipulation tasks.

document_idle: Script runs after the page is fully loaded or after a timeout, whichever comes first. This is the default and is best for non-critical operations that should not delay page rendering.

## Dynamic Registration

Dynamic content scripts are registered at runtime using chrome.scripting.registerContentScripts. This allows the extension to modify injection targets based on user preferences or runtime conditions. Requires the "scripting" permission.

```javascript
// Register content scripts dynamically
async function registerScripts() {
  await chrome.scripting.registerContentScripts([
    {
      id: 'feature-enhancement',
      matches: ['https://*.example.com/*'],
      js: ['content/feature.js'],
      css: ['content/feature.css'],
      runAt: 'document_end',
      allFrames: false
    }
  ]);
}

// Update existing registered scripts
async function updateScripts(newMatches) {
  await chrome.scripting.updateContentScripts([
    {
      id: 'feature-enhancement',
      matches: newMatches
    }
  ]);
}

// Unregister scripts when feature is disabled
async function unregisterScripts() {
  await chrome.scripting.unregisterContentScripts({
    ids: ['feature-enhancement']
  });
}

// List all registered scripts
async function listRegisteredScripts() {
  const scripts = await chrome.scripting.getRegisteredContentScripts();
  console.log('Registered scripts:', scripts);
}
```

Dynamically registered scripts persist across service worker restarts and browser sessions until explicitly unregistered. They are useful for user-configurable site lists and toggling features on or off.

## Programmatic Injection

Programmatic injection uses chrome.scripting.executeScript to inject code on demand. This requires either host_permissions for the target URL or activeTab permission (granted when the user clicks the extension action).

```javascript
// Inject a script file into the active tab
async function injectScript() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    files: ['content/injected.js']
  });
}

// Inject a function directly
async function injectFunction(tabId) {
  const results = await chrome.scripting.executeScript({
    target: { tabId },
    func: (color) => {
      document.body.style.backgroundColor = color;
      return document.title;
    },
    args: ['#f0f0f0']
  });

  // results is an array of InjectionResult objects
  console.log('Page title:', results[0].result);
}

// Inject into specific frames
async function injectIntoFrames(tabId, frameIds) {
  await chrome.scripting.executeScript({
    target: { tabId, frameIds },
    files: ['content/frame-script.js']
  });
}

// Inject CSS
async function injectCSS(tabId) {
  await chrome.scripting.insertCSS({
    target: { tabId },
    css: '.extension-highlight { background: yellow !important; }'
  });
}

// Remove injected CSS
async function removeCSS(tabId) {
  await chrome.scripting.removeCSS({
    target: { tabId },
    css: '.extension-highlight { background: yellow !important; }'
  });
}
```

## Available Chrome APIs

Content scripts have direct access to a limited set of Chrome APIs:

- chrome.dom for DOM manipulation utilities
- chrome.i18n for internationalization strings
- chrome.storage for reading and writing extension storage
- chrome.runtime.connect for establishing port connections
- chrome.runtime.sendMessage for sending one-time messages
- chrome.runtime.onMessage for receiving messages
- chrome.runtime.onConnect for receiving port connections
- chrome.runtime.getManifest for reading manifest data
- chrome.runtime.getURL for resolving extension resource URLs
- chrome.runtime.id for the extension ID

All other Chrome APIs must be accessed by sending messages to the service worker.

```javascript
// content/main.js

// Direct API access - chrome.storage
async function loadSettings() {
  const { settings } = await chrome.storage.local.get('settings');
  return settings || {};
}

// Direct API access - chrome.runtime.getURL
function getExtensionResource(path) {
  return chrome.runtime.getURL(path);
}

// Indirect API access - request from service worker
async function queryTabs() {
  const response = await chrome.runtime.sendMessage({
    action: 'query-tabs',
    params: { currentWindow: true }
  });
  return response.tabs;
}

// Indirect API access - create notification via service worker
async function notify(title, message) {
  await chrome.runtime.sendMessage({
    action: 'create-notification',
    params: { title, message }
  });
}
```

## DOM Interaction Patterns

Content scripts have full access to the page DOM. Use standard DOM APIs for reading, modifying, and observing page content.

```javascript
// Read page content
function extractPageData() {
  const title = document.title;
  const meta = document.querySelector('meta[name="description"]')?.content || '';
  const headings = [...document.querySelectorAll('h1, h2, h3')].map(h => ({
    level: h.tagName,
    text: h.textContent.trim()
  }));
  const links = [...document.querySelectorAll('a[href]')].map(a => ({
    text: a.textContent.trim(),
    href: a.href
  }));

  return { title, meta, headings, links };
}

// Modify page content safely
function highlightElements(selector, color = '#ffeb3b') {
  const elements = document.querySelectorAll(selector);
  elements.forEach(el => {
    el.style.outline = `2px solid ${color}`;
    el.dataset.extensionHighlighted = 'true';
  });
}

// Observe DOM changes for dynamic content
function observePageChanges(targetSelector, callback) {
  const target = document.querySelector(targetSelector);
  if (!target) return null;

  const observer = new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type === 'childList') {
        for (const node of mutation.addedNodes) {
          if (node.nodeType === Node.ELEMENT_NODE) {
            callback(node);
          }
        }
      }
    }
  });

  observer.observe(target, {
    childList: true,
    subtree: true
  });

  return observer;
}

// Clean up when content script is destroyed
function cleanup() {
  // Remove injected elements
  document.querySelectorAll('[data-extension-highlighted]').forEach(el => {
    el.style.outline = '';
    delete el.dataset.extensionHighlighted;
  });
}
```

## Shadow DOM for Encapsulated UI

When injecting UI elements into web pages, use Shadow DOM to encapsulate styles and prevent conflicts with the host page CSS.

```javascript
// Create encapsulated UI using Shadow DOM
function createExtensionUI() {
  const host = document.createElement('div');
  host.id = 'my-extension-root';
  const shadow = host.attachShadow({ mode: 'closed' });

  shadow.innerHTML = `
    <style>
      :host {
        all: initial;
        position: fixed;
        bottom: 20px;
        right: 20px;
        z-index: 2147483647;
        font-family: system-ui, sans-serif;
      }
      .panel {
        background: white;
        border: 1px solid #ddd;
        border-radius: 8px;
        padding: 16px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        width: 300px;
        max-height: 400px;
        overflow-y: auto;
      }
      .header {
        font-size: 14px;
        font-weight: 600;
        margin-bottom: 12px;
        color: #333;
      }
      .close-btn {
        float: right;
        background: none;
        border: none;
        cursor: pointer;
        font-size: 18px;
        color: #666;
      }
    </style>
    <div class="panel">
      <button class="close-btn" id="close">&times;</button>
      <div class="header">Extension Panel</div>
      <div id="content"></div>
    </div>
  `;

  shadow.getElementById('close').addEventListener('click', () => {
    host.remove();
  });

  document.body.appendChild(host);
  return shadow.getElementById('content');
}
```

## Security Considerations

Content scripts should be treated as operating in a potentially hostile environment. The host page can manipulate the shared DOM, observe changes, and attempt to exploit the extension.

Never trust data from the page DOM without validation and sanitization. Avoid using innerHTML, outerHTML, or document.write with untrusted data. Use textContent for text insertion and createElement with setAttribute for element creation.

Do not store secrets, API keys, or sensitive data in the DOM or in variables accessible via the shared DOM. Use chrome.storage and message passing to keep sensitive data in the service worker context.

Avoid modifying page JavaScript behavior by overwriting prototypes or global variables, as this can break page functionality and create security vectors.

When loading extension resources into the page, ensure they are declared in web_accessible_resources and scoped to the minimum necessary origins.

```javascript
// SAFE: Use textContent and createElement
function displayData(container, data) {
  const item = document.createElement('div');
  item.textContent = data.text; // Safe - no HTML parsing
  item.setAttribute('class', 'extension-item');
  container.appendChild(item);
}

// UNSAFE: Never use innerHTML with untrusted data
function unsafeDisplay(container, data) {
  // DO NOT DO THIS
  container.innerHTML = data.html; // XSS vulnerability
}

// SAFE: Validate messages from service worker
chrome.runtime.onMessage.addListener((message, sender) => {
  // Verify sender is our extension
  if (sender.id !== chrome.runtime.id) return;

  // Validate message structure
  if (!message.action || typeof message.action !== 'string') return;

  switch (message.action) {
    case 'highlight':
      if (typeof message.selector === 'string') {
        highlightElements(message.selector);
      }
      break;
  }
});
```

## Communication with Service Worker

Content scripts communicate with the service worker primarily through chrome.runtime.sendMessage for one-time messages and chrome.runtime.connect for ongoing connections.

```javascript
// One-time message to service worker
async function requestFromBackground(action, params = {}) {
  try {
    const response = await chrome.runtime.sendMessage({ action, ...params });
    if (response.error) {
      console.error('Background error:', response.error);
      return null;
    }
    return response.data;
  } catch (error) {
    // Extension context may be invalidated (e.g., extension updated)
    console.error('Message failed:', error.message);
    return null;
  }
}

// Long-lived connection
function connectToBackground() {
  const port = chrome.runtime.connect({ name: 'content-script' });

  port.onMessage.addListener((message) => {
    console.log('Received from background:', message);
  });

  port.onDisconnect.addListener(() => {
    console.log('Disconnected from background');
    // Optionally reconnect
  });

  return port;
}
```

## Handling Extension Updates

When the extension updates, existing content scripts become orphaned. Their chrome.runtime connection is severed, and API calls will fail. Handle this gracefully by detecting the disconnection.

```javascript
// Detect extension update / context invalidation
function isExtensionContextValid() {
  try {
    chrome.runtime.getManifest();
    return true;
  } catch {
    return false;
  }
}

// Wrap API calls with context check
async function safeSendMessage(message) {
  if (!isExtensionContextValid()) {
    console.log('Extension context invalidated, cleaning up');
    cleanup();
    return null;
  }

  try {
    return await chrome.runtime.sendMessage(message);
  } catch (error) {
    if (error.message.includes('Extension context invalidated')) {
      cleanup();
      return null;
    }
    throw error;
  }
}
```
