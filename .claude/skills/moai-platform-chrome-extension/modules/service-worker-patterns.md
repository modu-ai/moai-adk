---
name: service-worker-patterns
description: Service worker lifecycle, event registration, state management, and debugging patterns
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Service Worker Patterns

## Overview

Manifest V3 replaces persistent background pages with service workers. Service workers are event-driven scripts that run only when needed and terminate when idle. This fundamental architecture change requires specific patterns for event handling, state management, and long-running operations.

## Lifecycle

Service workers follow a well-defined lifecycle. Chrome loads the service worker when an event it listens for fires, such as a message from a content script, a toolbar button click, or an alarm trigger. The service worker initializes, processes the event, and terminates when all event handlers complete and no pending operations remain.

Chrome may terminate an idle service worker after approximately 30 seconds of inactivity, or after 5 minutes of continuous execution. This means the service worker script executes from scratch each time it wakes up, losing all in-memory state.

The install event fires when the extension is first installed or updated. The activate event fires after installation is complete. Use these events for one-time initialization tasks like setting default storage values or creating context menus.

```javascript
// service-worker.js

// Install event - runs once on install or update
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'install') {
    // First-time installation setup
    chrome.storage.local.set({
      settings: { theme: 'light', notifications: true },
      installDate: Date.now()
    });

    // Create context menu items
    chrome.contextMenus.create({
      id: 'main-action',
      title: 'Process with Extension',
      contexts: ['selection']
    });
  }

  if (details.reason === 'update') {
    // Extension updated - migrate data if needed
    console.log(`Updated from ${details.previousVersion}`);
  }
});
```

## Top-Level Event Registration

All event listeners must be registered at the top level of the service worker script. Chrome records which events the service worker listens for during initial execution. Listeners registered inside callbacks, promises, setTimeout, or conditional blocks will not persist across restarts.

```javascript
// CORRECT: Top-level registration
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  handleMessage(message, sender, sendResponse);
  return true; // Keep channel open for async response
});

chrome.alarms.onAlarm.addListener((alarm) => {
  handleAlarm(alarm);
});

chrome.action.onClicked.addListener((tab) => {
  handleActionClick(tab);
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  handleContextMenuClick(info, tab);
});

// WRONG: Conditional or delayed registration
// These listeners will NOT persist across service worker restarts
async function setup() {
  // DO NOT register listeners here
  chrome.runtime.onMessage.addListener(() => { /* ... */ });
}

setTimeout(() => {
  // DO NOT register listeners here
  chrome.alarms.onAlarm.addListener(() => { /* ... */ });
}, 1000);
```

## State Management

Since service workers lose all in-memory state when they terminate, persistent data must be stored using chrome.storage. Choose the appropriate storage area based on your needs.

chrome.storage.local provides up to 10 MB of local storage that persists across browser sessions. Use this for extension data, user preferences, and cached content.

chrome.storage.sync provides up to 100 KB that synchronizes across all Chrome instances where the user is signed in. Use this for user settings that should follow the user across devices.

chrome.storage.session provides in-memory storage that persists only for the browser session. It is faster than local storage but clears when the browser closes. Use this for temporary state like active feature flags.

```javascript
// State management pattern using chrome.storage

// Save state
async function saveState(key, value) {
  await chrome.storage.local.set({ [key]: value });
}

// Load state with default value
async function loadState(key, defaultValue) {
  const result = await chrome.storage.local.get({ [key]: defaultValue });
  return result[key];
}

// Reactive state updates - listen for changes
chrome.storage.onChanged.addListener((changes, areaName) => {
  if (areaName === 'local') {
    for (const [key, { oldValue, newValue }] of Object.entries(changes)) {
      console.log(`Storage key "${key}" changed from`, oldValue, 'to', newValue);
    }
  }
});

// Session storage for temporary data
async function setSessionData(data) {
  await chrome.storage.session.set(data);
}

// Example: Tracking active tabs with persistent state
chrome.tabs.onActivated.addListener(async (activeInfo) => {
  const state = await loadState('activeTabHistory', []);
  state.push({ tabId: activeInfo.tabId, timestamp: Date.now() });
  // Keep only last 50 entries
  if (state.length > 50) state.splice(0, state.length - 50);
  await saveState('activeTabHistory', state);
});
```

## Alarms API for Scheduled Tasks

Replace setTimeout and setInterval with the chrome.alarms API. Alarms persist across service worker restarts and have a minimum interval of 30 seconds in Manifest V3.

```javascript
// Create a periodic alarm
chrome.alarms.create('check-updates', {
  delayInMinutes: 1,       // First fire after 1 minute
  periodInMinutes: 30      // Repeat every 30 minutes
});

// Create a one-time alarm
chrome.alarms.create('delayed-task', {
  delayInMinutes: 5        // Fire once after 5 minutes
});

// Handle alarm events (must be top-level)
chrome.alarms.onAlarm.addListener(async (alarm) => {
  switch (alarm.name) {
    case 'check-updates':
      await performUpdateCheck();
      break;
    case 'delayed-task':
      await executeDelayedTask();
      break;
  }
});

// Clear a specific alarm
async function cancelAlarm(name) {
  await chrome.alarms.clear(name);
}

// List all active alarms
async function listAlarms() {
  const alarms = await chrome.alarms.getAll();
  console.log('Active alarms:', alarms);
}
```

## Keep-Alive Patterns

For operations that take longer than the idle timeout, use keep-alive strategies to prevent premature termination. Periodically calling a Chrome API resets the idle timer.

```javascript
// Keep-alive pattern for long-running operations
async function performLongOperation() {
  const keepAlive = setInterval(() => {
    chrome.runtime.getPlatformInfo(() => {});
  }, 25000); // Reset idle timer every 25 seconds

  try {
    // Your long-running operation
    const data = await fetchLargeDataset();
    await processData(data);
    await saveResults();
  } finally {
    clearInterval(keepAlive);
  }
}

// Alternative: Use chrome.offscreen for truly long operations
async function startOffscreenProcessing() {
  // Check if offscreen document already exists
  const existing = await chrome.offscreen.hasDocument();
  if (!existing) {
    await chrome.offscreen.createDocument({
      url: 'offscreen.html',
      reasons: ['DOM_PARSER'],
      justification: 'Parse HTML content from fetched pages'
    });
  }

  // Send work to offscreen document
  chrome.runtime.sendMessage({
    target: 'offscreen',
    action: 'parse-html',
    data: htmlContent
  });
}
```

## Offscreen Documents

When service workers need DOM access (for parsing HTML, playing audio, using canvas, or clipboard operations), use the Offscreen Documents API. Each extension can have one offscreen document at a time.

Valid reasons for creating offscreen documents include: TESTING, AUDIO_PLAYBACK, IFRAME_SCRIPTING, DOM_SCRAPING, BLOBS, DOM_PARSER, USER_MEDIA, DISPLAY_MEDIA, WEB_RTC, CLIPBOARD, LOCAL_STORAGE, WORKERS, BATTERY_STATUS, MATCH_MEDIA, and GEOLOCATION.

```javascript
// service-worker.js - Create and communicate with offscreen document

async function ensureOffscreenDocument() {
  if (await chrome.offscreen.hasDocument()) return;

  await chrome.offscreen.createDocument({
    url: 'offscreen/offscreen.html',
    reasons: ['DOM_PARSER'],
    justification: 'Parse and extract data from HTML content'
  });
}

async function parseHTML(htmlString) {
  await ensureOffscreenDocument();

  return new Promise((resolve) => {
    chrome.runtime.onMessage.addListener(function listener(message) {
      if (message.target === 'service-worker' && message.action === 'parse-result') {
        chrome.runtime.onMessage.removeListener(listener);
        resolve(message.data);
      }
    });

    chrome.runtime.sendMessage({
      target: 'offscreen',
      action: 'parse-html',
      data: htmlString
    });
  });
}
```

```html
<!-- offscreen/offscreen.html -->
<!DOCTYPE html>
<html>
<head><script src="offscreen.js"></script></head>
<body></body>
</html>
```

```javascript
// offscreen/offscreen.js
chrome.runtime.onMessage.addListener((message) => {
  if (message.target !== 'offscreen') return;

  if (message.action === 'parse-html') {
    const parser = new DOMParser();
    const doc = parser.parseFromString(message.data, 'text/html');

    // Extract data using DOM APIs
    const title = doc.querySelector('title')?.textContent || '';
    const headings = [...doc.querySelectorAll('h1, h2, h3')].map(h => h.textContent);
    const links = [...doc.querySelectorAll('a[href]')].map(a => ({
      text: a.textContent,
      href: a.getAttribute('href')
    }));

    chrome.runtime.sendMessage({
      target: 'service-worker',
      action: 'parse-result',
      data: { title, headings, links }
    });
  }
});
```

## Network Requests

Service workers support the fetch API for network requests. XMLHttpRequest is not available. Handle network errors gracefully since the service worker may not have network access.

```javascript
// Fetch with error handling and timeout
async function fetchWithTimeout(url, options = {}, timeoutMs = 10000) {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), timeoutMs);

  try {
    const response = await fetch(url, {
      ...options,
      signal: controller.signal
    });

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  } catch (error) {
    if (error.name === 'AbortError') {
      throw new Error(`Request to ${url} timed out after ${timeoutMs}ms`);
    }
    throw error;
  } finally {
    clearTimeout(timeout);
  }
}

// Cached fetch pattern
async function cachedFetch(url, cacheKey, maxAgeMs = 3600000) {
  const cached = await chrome.storage.local.get(cacheKey);
  if (cached[cacheKey] && Date.now() - cached[cacheKey].timestamp < maxAgeMs) {
    return cached[cacheKey].data;
  }

  const data = await fetchWithTimeout(url);
  await chrome.storage.local.set({
    [cacheKey]: { data, timestamp: Date.now() }
  });
  return data;
}
```

## ES Module Support

With type set to "module" in the manifest background field, the service worker supports ES module imports. Organize code into focused modules for maintainability.

```javascript
// service-worker.js (entry point)
import { setupMessageHandlers } from './handlers/messages.js';
import { setupAlarmHandlers } from './handlers/alarms.js';
import { setupContextMenus } from './handlers/context-menus.js';
import { setupTabHandlers } from './handlers/tabs.js';

// Register all handlers at top level
setupMessageHandlers();
setupAlarmHandlers();
setupContextMenus();
setupTabHandlers();

// Installation handler
chrome.runtime.onInstalled.addListener(async (details) => {
  if (details.reason === 'install') {
    await initializeExtension();
  }
});

async function initializeExtension() {
  await chrome.storage.local.set({ initialized: true });
}
```

```javascript
// handlers/messages.js
export function setupMessageHandlers() {
  chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
    // Handler logic
    return true;
  });
}
```

## Debugging Service Workers

Open chrome://extensions and enable Developer mode. Each extension card shows its status and provides a link to inspect the service worker.

Click "service worker" under the extension name to open DevTools for the service worker context. Use the Console, Sources, and Network panels as with regular DevTools.

Use chrome://serviceworker-internals for low-level service worker state inspection including registration status, running state, and event dispatch history.

Common debugging patterns:

- Add console.log at the top of the service worker to confirm wake-up events
- Use chrome.runtime.onSuspend to log when the service worker is about to terminate
- Check chrome.runtime.lastError after API calls for error information
- Monitor chrome.storage changes to verify state persistence

```javascript
// Debug logging for service worker lifecycle
console.log('Service worker starting', new Date().toISOString());

chrome.runtime.onSuspend.addListener(() => {
  console.log('Service worker suspending', new Date().toISOString());
});

chrome.runtime.onStartup.addListener(() => {
  console.log('Browser started', new Date().toISOString());
});
```
