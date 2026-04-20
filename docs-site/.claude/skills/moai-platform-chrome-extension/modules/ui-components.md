---
name: ui-components
description: Popup, side panel, options page, DevTools panel, and content script UI patterns
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# UI Components

## Overview

Chrome extensions provide multiple UI surfaces for user interaction. Each surface has different lifecycle characteristics, capabilities, and appropriate use cases. This module covers all UI component types with implementation patterns and best practices.

## Popup

The popup is the most common extension UI. It appears when the user clicks the toolbar button and closes when it loses focus. Popups are standard HTML pages with access to all extension APIs.

### Configuration

Configure the popup in manifest.json under the action field. Set default_popup to the HTML file path. The popup page loads fresh each time it opens and unloads when it closes.

```json
{
  "action": {
    "default_popup": "popup/popup.html",
    "default_icon": {
      "16": "icons/icon16.png",
      "32": "icons/icon32.png"
    },
    "default_title": "Open Extension"
  }
}
```

### Popup HTML Structure

```html
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <link rel="stylesheet" href="popup.css">
</head>
<body>
  <div class="container">
    <header class="header">
      <h1>Extension Name</h1>
    </header>
    <main id="content">
      <div class="loading">Loading...</div>
    </main>
    <footer class="footer">
      <button id="options-btn">Settings</button>
    </footer>
  </div>
  <script src="popup.js"></script>
</body>
</html>
```

### Popup JavaScript Pattern

```javascript
// popup/popup.js

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', async () => {
  await loadData();
  setupEventListeners();
});

async function loadData() {
  const content = document.getElementById('content');

  try {
    // Load data from storage or service worker
    const { settings } = await chrome.storage.local.get('settings');
    const data = await chrome.runtime.sendMessage({ action: 'get-status' });

    renderContent(content, settings, data);
  } catch (error) {
    content.innerHTML = '';
    const errorEl = document.createElement('div');
    errorEl.className = 'error';
    errorEl.textContent = `Error: ${error.message}`;
    content.appendChild(errorEl);
  }
}

function renderContent(container, settings, data) {
  container.innerHTML = '';

  const list = document.createElement('ul');
  list.className = 'item-list';

  for (const item of data.items) {
    const li = document.createElement('li');
    li.textContent = item.name;
    li.addEventListener('click', () => handleItemClick(item));
    list.appendChild(li);
  }

  container.appendChild(list);
}

function setupEventListeners() {
  document.getElementById('options-btn').addEventListener('click', () => {
    chrome.runtime.openOptionsPage();
  });
}

function handleItemClick(item) {
  chrome.runtime.sendMessage({
    action: 'process-item',
    itemId: item.id
  });
  window.close(); // Close popup after action
}
```

### Popup CSS

```css
/* popup/popup.css */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  width: 360px;
  min-height: 200px;
  max-height: 600px;
  font-family: system-ui, -apple-system, sans-serif;
  font-size: 14px;
  color: #333;
  overflow-y: auto;
}

.container {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.header {
  padding: 12px 16px;
  border-bottom: 1px solid #e0e0e0;
  background: #f8f9fa;
}

.header h1 {
  font-size: 16px;
  font-weight: 600;
}

main {
  flex: 1;
  padding: 12px 16px;
}

.footer {
  padding: 8px 16px;
  border-top: 1px solid #e0e0e0;
  text-align: right;
}

.item-list {
  list-style: none;
}

.item-list li {
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
}

.item-list li:hover {
  background: #e8f0fe;
}

.error {
  color: #d93025;
  padding: 8px;
  background: #fce8e6;
  border-radius: 4px;
}

.loading {
  text-align: center;
  padding: 20px;
  color: #666;
}

button {
  padding: 6px 16px;
  border: 1px solid #dadce0;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 13px;
}

button:hover {
  background: #f1f3f4;
}
```

### Popup Best Practices

Keep popup initialization fast since it loads on every click. Cache data in chrome.storage and load from cache first, then refresh from service worker. Set explicit width and max-height on the body to control popup dimensions. Avoid complex operations that block rendering.

The popup closes when it loses focus, so save any pending state before operations that might cause focus loss. Use window.close() to programmatically close the popup after completing actions.

## Side Panel

The side panel provides a persistent UI surface alongside web content. Unlike popups, side panels remain open as the user navigates. Available since Chrome 114.

### Configuration

```json
{
  "side_panel": {
    "default_path": "sidepanel/sidepanel.html"
  },
  "permissions": ["sidePanel"]
}
```

### Side Panel Implementation

```html
<!-- sidepanel/sidepanel.html -->
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <link rel="stylesheet" href="sidepanel.css">
</head>
<body>
  <div class="panel">
    <header>
      <h1>Side Panel</h1>
      <div id="tab-info"></div>
    </header>
    <main id="content"></main>
  </div>
  <script src="sidepanel.js"></script>
</body>
</html>
```

```javascript
// sidepanel/sidepanel.js

// Side panel persists across navigation
document.addEventListener('DOMContentLoaded', async () => {
  await initialize();
  listenForTabChanges();
});

async function initialize() {
  const { settings } = await chrome.storage.local.get('settings');
  applySettings(settings);
}

// React to active tab changes
function listenForTabChanges() {
  chrome.tabs.onActivated.addListener(async ({ tabId }) => {
    const tab = await chrome.tabs.get(tabId);
    updateTabInfo(tab);
  });

  chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
    if (changeInfo.status === 'complete' && tab.active) {
      updateTabInfo(tab);
    }
  });
}

function updateTabInfo(tab) {
  const tabInfo = document.getElementById('tab-info');
  tabInfo.textContent = tab.title || tab.url;
}
```

### Per-Tab Side Panel Content

```javascript
// service-worker.js - Configure different content per tab
chrome.tabs.onActivated.addListener(async ({ tabId }) => {
  const tab = await chrome.tabs.get(tabId);
  const url = new URL(tab.url);

  if (url.hostname.includes('github.com')) {
    await chrome.sidePanel.setOptions({
      tabId,
      path: 'sidepanel/github.html',
      enabled: true
    });
  } else if (url.hostname.includes('docs.google.com')) {
    await chrome.sidePanel.setOptions({
      tabId,
      path: 'sidepanel/docs.html',
      enabled: true
    });
  } else {
    await chrome.sidePanel.setOptions({
      tabId,
      path: 'sidepanel/default.html',
      enabled: true
    });
  }
});

// Open side panel on toolbar button click
chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });
```

### Side Panel Best Practices

Side panels persist across page navigation, so manage state carefully and update content when the active tab changes. Avoid heavy polling; use chrome.tabs events to react to navigation changes. Side panels share the same extension context as popups and options pages, with full chrome API access.

## Options Page

The options page provides extension configuration UI. It opens within the chrome://extensions page or in a new tab depending on configuration.

### Configuration

```json
{
  "options_ui": {
    "page": "options/options.html",
    "open_in_tab": false
  }
}
```

Setting open_in_tab to false embeds the options page within chrome://extensions, providing a cleaner experience. Setting it to true opens in a full browser tab.

### Options Page Implementation

```html
<!-- options/options.html -->
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <link rel="stylesheet" href="options.css">
</head>
<body>
  <div class="options-container">
    <h1>Extension Settings</h1>

    <section class="setting-group">
      <h2>General</h2>
      <label class="setting">
        <span>Enable notifications</span>
        <input type="checkbox" id="notifications-enabled">
      </label>
      <label class="setting">
        <span>Theme</span>
        <select id="theme-select">
          <option value="light">Light</option>
          <option value="dark">Dark</option>
          <option value="system">System</option>
        </select>
      </label>
    </section>

    <section class="setting-group">
      <h2>Sites</h2>
      <div id="site-list"></div>
      <button id="add-site">Add Site</button>
    </section>

    <div class="status" id="status"></div>
  </div>
  <script src="options.js"></script>
</body>
</html>
```

```javascript
// options/options.js

document.addEventListener('DOMContentLoaded', loadSettings);

async function loadSettings() {
  const { settings } = await chrome.storage.local.get({
    settings: {
      notificationsEnabled: true,
      theme: 'system',
      sites: []
    }
  });

  document.getElementById('notifications-enabled').checked = settings.notificationsEnabled;
  document.getElementById('theme-select').value = settings.theme;
  renderSiteList(settings.sites);

  // Auto-save on change
  document.getElementById('notifications-enabled').addEventListener('change', saveSettings);
  document.getElementById('theme-select').addEventListener('change', saveSettings);
  document.getElementById('add-site').addEventListener('click', addSite);
}

async function saveSettings() {
  const settings = {
    notificationsEnabled: document.getElementById('notifications-enabled').checked,
    theme: document.getElementById('theme-select').value,
    sites: getSiteList()
  };

  await chrome.storage.local.set({ settings });
  showStatus('Settings saved');
}

function showStatus(message) {
  const status = document.getElementById('status');
  status.textContent = message;
  status.className = 'status visible';
  setTimeout(() => {
    status.className = 'status';
  }, 2000);
}

function renderSiteList(sites) {
  const container = document.getElementById('site-list');
  container.innerHTML = '';
  sites.forEach((site, index) => {
    const item = document.createElement('div');
    item.className = 'site-item';

    const input = document.createElement('input');
    input.type = 'text';
    input.value = site;
    input.addEventListener('change', saveSettings);

    const removeBtn = document.createElement('button');
    removeBtn.textContent = 'Remove';
    removeBtn.addEventListener('click', () => {
      item.remove();
      saveSettings();
    });

    item.appendChild(input);
    item.appendChild(removeBtn);
    container.appendChild(item);
  });
}

function addSite() {
  const container = document.getElementById('site-list');
  const item = document.createElement('div');
  item.className = 'site-item';
  const input = document.createElement('input');
  input.type = 'text';
  input.placeholder = 'https://example.com';
  input.addEventListener('change', saveSettings);
  const removeBtn = document.createElement('button');
  removeBtn.textContent = 'Remove';
  removeBtn.addEventListener('click', () => { item.remove(); saveSettings(); });
  item.appendChild(input);
  item.appendChild(removeBtn);
  container.appendChild(item);
}

function getSiteList() {
  return [...document.querySelectorAll('.site-item input')].map(i => i.value).filter(Boolean);
}
```

## DevTools Panel

Extensions can add custom panels to Chrome DevTools. The DevTools page is loaded each time DevTools opens and can create panels, access inspected page information, and monitor network activity.

### Configuration

```json
{
  "devtools_page": "devtools/devtools.html"
}
```

### DevTools Implementation

```html
<!-- devtools/devtools.html -->
<!DOCTYPE html>
<html>
<head><script src="devtools.js"></script></head>
<body></body>
</html>
```

```javascript
// devtools/devtools.js - Creates the DevTools panel
chrome.devtools.panels.create(
  'My Extension',           // Panel title
  'icons/icon16.png',       // Panel icon
  'devtools/panel.html',    // Panel HTML page
  (panel) => {
    panel.onShown.addListener((panelWindow) => {
      // Panel is visible - initialize or update
    });
    panel.onHidden.addListener(() => {
      // Panel is hidden - pause updates
    });
  }
);
```

```javascript
// devtools/panel.js - Panel page logic

// Access inspected page information
const tabId = chrome.devtools.inspectedWindow.tabId;

// Execute code in the inspected page
function inspectElement(selector) {
  chrome.devtools.inspectedWindow.eval(
    `document.querySelector('${selector}')`,
    (result, exceptionInfo) => {
      if (exceptionInfo) {
        console.error('Eval error:', exceptionInfo);
        return;
      }
      displayResult(result);
    }
  );
}

// Monitor network requests
chrome.devtools.network.onRequestFinished.addListener((request) => {
  if (request.response.content.mimeType.includes('json')) {
    request.getContent((content) => {
      logAPIResponse(request.request.url, content);
    });
  }
});
```

### DevTools Best Practices

DevTools pages load when DevTools opens and unload when DevTools closes. They have access to chrome.devtools APIs but limited access to other extension APIs. Use message passing to communicate with the service worker for operations requiring full API access.

## Content Script UI Injection

Content scripts can inject UI elements directly into web pages. For complex UI, use Shadow DOM for style encapsulation as described in the content-scripts-guide module.

### Inline Notification Bar

```javascript
// content/notification.js - Inject a notification bar into the page
function showNotificationBar(message, type = 'info') {
  // Remove existing bar if present
  const existing = document.getElementById('ext-notification-bar');
  if (existing) existing.remove();

  const bar = document.createElement('div');
  bar.id = 'ext-notification-bar';
  bar.setAttribute('style', `
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 2147483647;
    padding: 12px 20px;
    font-family: system-ui, sans-serif;
    font-size: 14px;
    text-align: center;
    background: ${type === 'info' ? '#1a73e8' : type === 'warning' ? '#f9ab00' : '#d93025'};
    color: white;
    box-shadow: 0 2px 8px rgba(0,0,0,0.2);
    transition: transform 0.3s ease;
  `);

  const text = document.createElement('span');
  text.textContent = message;

  const closeBtn = document.createElement('button');
  closeBtn.textContent = 'Dismiss';
  closeBtn.setAttribute('style', `
    margin-left: 16px;
    padding: 4px 12px;
    border: 1px solid rgba(255,255,255,0.5);
    border-radius: 4px;
    background: transparent;
    color: white;
    cursor: pointer;
    font-size: 13px;
  `);
  closeBtn.addEventListener('click', () => bar.remove());

  bar.appendChild(text);
  bar.appendChild(closeBtn);
  document.body.appendChild(bar);

  // Auto-dismiss after 5 seconds
  setTimeout(() => {
    if (bar.parentNode) {
      bar.style.transform = 'translateY(-100%)';
      setTimeout(() => bar.remove(), 300);
    }
  }, 5000);
}
```

## UI Communication Between Components

All extension UI components (popup, side panel, options, DevTools panel) can communicate with each other and the service worker using standard chrome.runtime messaging. They share the same extension context and can also access chrome.storage for shared state.

```javascript
// Any extension page can communicate with others
// Broadcast a message that all extension pages receive
chrome.runtime.sendMessage({ type: 'settings-updated', settings: newSettings });

// Listen for broadcasts
chrome.runtime.onMessage.addListener((message) => {
  if (message.type === 'settings-updated') {
    applySettings(message.settings);
  }
});
```
