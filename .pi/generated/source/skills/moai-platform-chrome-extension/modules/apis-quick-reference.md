---
name: apis-quick-reference
description: Quick reference for major chrome.* APIs with method signatures and permissions
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Chrome APIs Quick Reference

## Overview

This module provides a quick reference for the most commonly used chrome.* APIs in Manifest V3 extensions. Each section lists the required permission, key methods, events, and common usage patterns.

## chrome.runtime

Permission: None required (available to all extension contexts)

Purpose: Extension lifecycle management, messaging, manifest access, and platform information.

### Key Methods

chrome.runtime.sendMessage(message) sends a one-time message to the service worker or other extension pages. Returns a Promise resolving to the response.

chrome.runtime.connect(connectInfo) establishes a long-lived port connection to the service worker. The connectInfo object accepts a name property.

chrome.runtime.getURL(path) converts a relative extension resource path to a fully qualified URL.

chrome.runtime.getManifest() returns the parsed manifest.json as an object.

chrome.runtime.getPlatformInfo() returns platform details including os, arch, and nacl_arch.

chrome.runtime.setUninstallURL(url) sets a URL to open when the user uninstalls the extension.

chrome.runtime.openOptionsPage() opens the extension options page.

chrome.runtime.reload() reloads the extension.

chrome.runtime.id is the extension ID string.

### Key Events

chrome.runtime.onInstalled fires when the extension is installed, updated, or Chrome updates. The callback receives a details object with reason (install, update, chrome_update) and previousVersion.

chrome.runtime.onStartup fires when the browser profile starts with the extension enabled.

chrome.runtime.onSuspend fires just before the service worker is terminated. Use for cleanup.

chrome.runtime.onMessage fires when a message is sent via sendMessage from any context.

chrome.runtime.onConnect fires when a port connection is established.

chrome.runtime.onMessageExternal fires when a message arrives from another extension or web page.

chrome.runtime.onConnectExternal fires when a port connection arrives from another extension.

```javascript
// Common runtime patterns
chrome.runtime.onInstalled.addListener(({ reason }) => {
  if (reason === 'install') {
    chrome.runtime.openOptionsPage();
  }
});

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  // Handle messages
  return true; // For async responses
});
```

## chrome.tabs

Permission: "tabs" for URL/title access; no permission needed for basic tab operations

Purpose: Browser tab management including creation, querying, updating, and removal.

### Key Methods

chrome.tabs.query(queryInfo) finds tabs matching the query criteria. Common properties: active, currentWindow, url, status, windowId. Returns a Promise resolving to an array of Tab objects.

chrome.tabs.create(createProperties) creates a new tab. Properties include url, active, index, windowId, pinned. Returns the created Tab.

chrome.tabs.update(tabId, updateProperties) modifies a tab. Properties include url, active, muted, pinned.

chrome.tabs.remove(tabIds) closes one or more tabs by ID. Accepts a single ID or array.

chrome.tabs.get(tabId) retrieves a Tab object by ID.

chrome.tabs.sendMessage(tabId, message) sends a message to a content script in the specified tab.

chrome.tabs.connect(tabId, connectInfo) opens a port to a content script in the specified tab.

chrome.tabs.reload(tabId) reloads a tab.

chrome.tabs.duplicate(tabId) duplicates a tab.

chrome.tabs.group(options) groups tabs. Options include tabIds and groupId.

chrome.tabs.ungroup(tabIds) removes tabs from their groups.

### Key Events

chrome.tabs.onCreated fires when a new tab is created.

chrome.tabs.onUpdated fires when a tab property changes. Callback receives tabId, changeInfo, and tab.

chrome.tabs.onRemoved fires when a tab is closed.

chrome.tabs.onActivated fires when the active tab changes. Callback receives activeInfo with tabId and windowId.

chrome.tabs.onMoved fires when a tab is moved within a window.

```javascript
// Common tabs patterns
const [activeTab] = await chrome.tabs.query({ active: true, currentWindow: true });

const newTab = await chrome.tabs.create({ url: 'https://example.com', active: false });

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === 'complete') {
    // Page finished loading
  }
});
```

## chrome.storage

Permission: "storage"

Purpose: Persistent key-value storage with three areas optimized for different use cases.

### Storage Areas

chrome.storage.local provides up to 10 MB of local storage. Data persists across browser sessions.

chrome.storage.sync provides up to 100 KB that synchronizes across Chrome instances via the user's Google account. Individual items limited to 8 KB. Maximum 512 items.

chrome.storage.session provides in-memory storage for the browser session. Clears when the browser closes. Accessible only from the extension service worker and extension pages (not content scripts by default, unless access is enabled).

### Key Methods

storage.get(keys) retrieves values by key. Keys can be a string, array of strings, or object with default values. Returns a Promise.

storage.set(items) stores key-value pairs. Items is an object.

storage.remove(keys) removes one or more keys.

storage.clear() removes all items from the storage area.

storage.getBytesInUse(keys) returns storage consumption in bytes.

storage.setAccessLevel(accessOptions) for session storage, controls whether content scripts can access it.

### Key Events

chrome.storage.onChanged fires when any storage value changes. Callback receives changes (object mapping keys to {oldValue, newValue}) and areaName ("local", "sync", or "session").

```javascript
// Common storage patterns
await chrome.storage.local.set({ key: 'value', settings: { theme: 'dark' } });

const { settings } = await chrome.storage.local.get({ settings: { theme: 'light' } });

chrome.storage.onChanged.addListener((changes, area) => {
  if (area === 'local' && changes.settings) {
    applySettings(changes.settings.newValue);
  }
});

// Enable content script access to session storage
await chrome.storage.session.setAccessLevel({
  accessLevel: 'TRUSTED_AND_UNTRUSTED_CONTEXTS'
});
```

## chrome.action

Permission: None required (configured via manifest "action" field)

Purpose: Toolbar button management including icon, badge, popup, and click handling.

### Key Methods

chrome.action.setIcon(details) sets the toolbar icon. Details include imageData or path with size keys, and optional tabId for per-tab icons.

chrome.action.setBadgeText(details) sets badge text. Details include text and optional tabId.

chrome.action.setBadgeBackgroundColor(details) sets badge background color.

chrome.action.setBadgeTextColor(details) sets badge text color.

chrome.action.setTitle(details) sets the tooltip title.

chrome.action.setPopup(details) sets the popup HTML page. Pass empty string to disable popup.

chrome.action.openPopup() programmatically opens the popup.

chrome.action.enable(tabId) / chrome.action.disable(tabId) controls button state per tab.

### Key Events

chrome.action.onClicked fires when the toolbar button is clicked AND no popup is configured. If a popup is set, clicking opens the popup instead of firing this event.

```javascript
// Common action patterns
chrome.action.setBadgeText({ text: '5' });
chrome.action.setBadgeBackgroundColor({ color: '#4688F1' });

chrome.action.onClicked.addListener(async (tab) => {
  // Only fires when no popup is set
  await chrome.scripting.executeScript({
    target: { tabId: tab.id },
    files: ['content/toggle.js']
  });
});
```

## chrome.scripting

Permission: "scripting"

Purpose: Programmatic script and CSS injection into web pages. Also manages dynamically registered content scripts.

### Key Methods

chrome.scripting.executeScript(details) injects JavaScript into a tab. Target specifies tabId and optional frameIds or allFrames. Injection via files array or func with optional args.

chrome.scripting.insertCSS(details) injects CSS into a tab. Specify css string or files array.

chrome.scripting.removeCSS(details) removes previously injected CSS.

chrome.scripting.registerContentScripts(scripts) dynamically registers content scripts.

chrome.scripting.updateContentScripts(scripts) updates registered scripts.

chrome.scripting.unregisterContentScripts(filter) removes registered scripts.

chrome.scripting.getRegisteredContentScripts(filter) lists registered scripts.

```javascript
// Inject function with arguments
const results = await chrome.scripting.executeScript({
  target: { tabId: tab.id },
  func: (greeting) => {
    return `${greeting}, page title is: ${document.title}`;
  },
  args: ['Hello']
});
console.log(results[0].result); // "Hello, page title is: ..."
```

## chrome.alarms

Permission: "alarms"

Purpose: Schedule events to fire at specific times or intervals. Minimum interval is 30 seconds in Manifest V3.

### Key Methods

chrome.alarms.create(name, alarmInfo) creates an alarm. AlarmInfo accepts delayInMinutes, periodInMinutes, and when (timestamp).

chrome.alarms.get(name) retrieves a specific alarm.

chrome.alarms.getAll() retrieves all active alarms.

chrome.alarms.clear(name) cancels a specific alarm.

chrome.alarms.clearAll() cancels all alarms.

### Key Events

chrome.alarms.onAlarm fires when an alarm triggers. Callback receives the Alarm object with name, scheduledTime, and periodInMinutes.

```javascript
chrome.alarms.create('periodic-check', { periodInMinutes: 60 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === 'periodic-check') performCheck();
});
```

## chrome.notifications

Permission: "notifications"

Purpose: System-level notifications displayed outside the browser window.

### Key Methods

chrome.notifications.create(notificationId, options) creates a notification. Options include type (basic, image, list, progress), title, message, iconUrl, and optional buttons.

chrome.notifications.update(notificationId, options) updates an existing notification.

chrome.notifications.clear(notificationId) removes a notification.

### Key Events

chrome.notifications.onClicked fires when the user clicks the notification body.

chrome.notifications.onButtonClicked fires when the user clicks a notification button.

chrome.notifications.onClosed fires when a notification is closed.

```javascript
chrome.notifications.create('update-available', {
  type: 'basic',
  iconUrl: 'icons/icon128.png',
  title: 'Update Available',
  message: 'A new version is ready to install.',
  buttons: [{ title: 'Install Now' }, { title: 'Later' }]
});
```

## chrome.contextMenus

Permission: "contextMenus"

Purpose: Custom items in the browser right-click context menu.

### Key Methods

chrome.contextMenus.create(createProperties) adds a menu item. Properties include id, title, contexts (page, selection, link, image, etc.), parentId, and type (normal, checkbox, radio, separator).

chrome.contextMenus.update(id, updateProperties) modifies an existing item.

chrome.contextMenus.remove(menuItemId) removes a specific item.

chrome.contextMenus.removeAll() removes all items.

### Key Events

chrome.contextMenus.onClicked fires when a menu item is clicked. Callback receives info (menuItemId, selectionText, linkUrl, srcUrl, pageUrl) and tab.

```javascript
chrome.contextMenus.create({
  id: 'search-selection',
  title: 'Search for "%s"',
  contexts: ['selection']
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === 'search-selection') {
    chrome.tabs.create({
      url: `https://search.example.com?q=${encodeURIComponent(info.selectionText)}`
    });
  }
});
```

## chrome.sidePanel

Permission: "sidePanel"

Purpose: Persistent side panel UI alongside web content.

### Key Methods

chrome.sidePanel.setOptions(options) configures the side panel. Options include path (HTML file), enabled, and optional tabId for per-tab configuration.

chrome.sidePanel.getOptions(options) retrieves current configuration.

chrome.sidePanel.open(options) opens the side panel programmatically. Options include windowId or tabId.

chrome.sidePanel.setPanelBehavior(behavior) controls open behavior. Set openPanelOnActionClick to true to open the side panel when the toolbar button is clicked.

```javascript
// Open side panel on action click
chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });

// Per-tab side panel content
chrome.tabs.onActivated.addListener(async ({ tabId }) => {
  await chrome.sidePanel.setOptions({
    tabId,
    path: 'sidepanel/tab-specific.html',
    enabled: true
  });
});
```

## chrome.declarativeNetRequest

Permission: "declarativeNetRequest" or "declarativeNetRequestWithHostAccess"

Purpose: Network request filtering using declarative rules without intercepting requests.

### Key Methods

chrome.declarativeNetRequest.updateDynamicRules(options) adds or removes dynamic rules. Options include addRules and removeRuleIds.

chrome.declarativeNetRequest.getDynamicRules() retrieves all dynamic rules.

chrome.declarativeNetRequest.updateSessionRules(options) manages session-scoped rules.

chrome.declarativeNetRequest.getMatchedRules(filter) retrieves rules that matched recent requests.

### Rule Structure

Each rule has an id, priority, action (block, redirect, modifyHeaders, allow, upgradeScheme), and condition (urlFilter or regexFilter, resourceTypes, domains, excludedDomains).

```javascript
// Block specific URLs dynamically
await chrome.declarativeNetRequest.updateDynamicRules({
  addRules: [{
    id: 1,
    priority: 1,
    action: { type: 'block' },
    condition: {
      urlFilter: 'tracking.example.com',
      resourceTypes: ['script', 'xmlhttprequest']
    }
  }],
  removeRuleIds: []
});
```

## chrome.offscreen

Permission: "offscreen"

Purpose: Create hidden documents with DOM access for use from the service worker.

### Key Methods

chrome.offscreen.createDocument(parameters) creates an offscreen document. Parameters include url, reasons (array of OffscreenReason), and justification string.

chrome.offscreen.closeDocument() closes the active offscreen document.

chrome.offscreen.hasDocument() checks if an offscreen document exists. Returns boolean.

## chrome.identity

Permission: "identity"

Purpose: OAuth2 authentication and token management.

### Key Methods

chrome.identity.getAuthToken(details) obtains an OAuth2 token. Set interactive to true for user consent UI.

chrome.identity.removeCachedAuthToken(details) removes a cached token.

chrome.identity.launchWebAuthFlow(details) launches a web auth flow for non-Google OAuth providers. Specify url and interactive.

chrome.identity.getProfileUserInfo(details) retrieves the signed-in user's email and ID.

## chrome.commands

Permission: None required (configured via manifest "commands" field)

Purpose: Keyboard shortcut handling.

### Key Events

chrome.commands.onCommand fires when a registered keyboard shortcut is pressed. Callback receives the command name and the active tab.

```javascript
chrome.commands.onCommand.addListener((command, tab) => {
  if (command === 'toggle-feature') {
    toggleFeature(tab.id);
  }
});
```

## chrome.permissions

Permission: None required

Purpose: Runtime permission management for optional permissions.

### Key Methods

chrome.permissions.request(permissions) prompts the user to grant permissions. The permissions object includes optional permissions and origins arrays.

chrome.permissions.remove(permissions) revokes granted permissions.

chrome.permissions.contains(permissions) checks if specific permissions are granted.

chrome.permissions.getAll() retrieves all currently granted permissions.

### Key Events

chrome.permissions.onAdded fires when new permissions are granted.

chrome.permissions.onRemoved fires when permissions are revoked.
