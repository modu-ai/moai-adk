---
name: messaging-patterns
description: One-time messages, long-lived connections, cross-extension and web page messaging
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Messaging Patterns

## Overview

Chrome extensions communicate between components -- service worker, content scripts, popup, side panel, and options page -- using message passing. Messages are serialized as JSON (not structured clone), with a maximum size of 64 MiB. This module covers all messaging patterns from simple one-time requests to complex cross-extension communication.

## One-Time Messages

One-time messages are the simplest communication pattern. The sender sends a message and optionally receives a single response. Use chrome.runtime.sendMessage to send to the service worker from any other context, and chrome.tabs.sendMessage to send from the service worker to a specific content script.

### Content Script to Service Worker

```javascript
// content/main.js - Send message to service worker
async function fetchData(query) {
  const response = await chrome.runtime.sendMessage({
    action: 'fetch-data',
    query: query
  });
  return response;
}

// Usage
const result = await fetchData('search term');
console.log('Result:', result);
```

### Service Worker to Content Script

```javascript
// service-worker.js - Send message to content script in specific tab
async function sendToContentScript(tabId, message) {
  try {
    const response = await chrome.tabs.sendMessage(tabId, message);
    return response;
  } catch (error) {
    // Tab may not have the content script loaded
    console.error(`Failed to message tab ${tabId}:`, error.message);
    return null;
  }
}

// Send to active tab
async function sendToActiveTab(message) {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (tab?.id) {
    return await sendToContentScript(tab.id, message);
  }
  return null;
}
```

### Handling Messages in Service Worker

```javascript
// service-worker.js - Central message handler
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  // Identify the sender
  const isContentScript = sender.tab !== undefined;
  const isPopup = sender.url?.includes(chrome.runtime.getURL(''));

  switch (message.action) {
    case 'fetch-data':
      handleFetchData(message.query)
        .then(data => sendResponse({ success: true, data }))
        .catch(err => sendResponse({ success: false, error: err.message }));
      return true; // IMPORTANT: Keep message channel open for async response

    case 'get-settings':
      chrome.storage.local.get('settings').then(result => {
        sendResponse(result.settings || {});
      });
      return true;

    case 'save-data':
      handleSaveData(message.data)
        .then(() => sendResponse({ success: true }))
        .catch(err => sendResponse({ success: false, error: err.message }));
      return true;

    default:
      sendResponse({ error: 'Unknown action' });
      return false;
  }
});

async function handleFetchData(query) {
  const response = await fetch(`https://api.example.com/search?q=${encodeURIComponent(query)}`);
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return await response.json();
}
```

### Async Response Pattern

When handling messages asynchronously, you must either return true from the onMessage listener (keeping the sendResponse channel open) or return a Promise (Chrome 144+).

```javascript
// Pattern 1: return true + sendResponse (all Chrome versions)
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === 'async-task') {
    performAsyncTask(message.data).then(result => {
      sendResponse({ result });
    }).catch(error => {
      sendResponse({ error: error.message });
    });
    return true; // Required to keep sendResponse valid
  }
});

// Pattern 2: Return Promise directly (Chrome 144+)
chrome.runtime.onMessage.addListener((message, sender) => {
  if (message.action === 'async-task') {
    return performAsyncTask(message.data)
      .then(result => ({ result }))
      .catch(error => ({ error: error.message }));
  }
});
```

## Long-Lived Connections (Ports)

Long-lived connections use ports for ongoing bidirectional communication. Establish a connection with chrome.runtime.connect (to service worker) or chrome.tabs.connect (to content script). Ports remain open until explicitly disconnected, the tab closes, or the extension is unloaded.

### Content Script to Service Worker Port

```javascript
// content/main.js - Establish long-lived connection
const port = chrome.runtime.connect({ name: 'content-channel' });

// Send messages through the port
port.postMessage({ action: 'subscribe', topic: 'updates' });

// Receive messages
port.onMessage.addListener((message) => {
  switch (message.type) {
    case 'update':
      handleUpdate(message.data);
      break;
    case 'config-change':
      applyConfig(message.config);
      break;
  }
});

// Handle disconnection
port.onDisconnect.addListener(() => {
  console.log('Disconnected from service worker');
  if (chrome.runtime.lastError) {
    console.error('Disconnect error:', chrome.runtime.lastError.message);
  }
  // Optionally attempt reconnection
});

// Disconnect when done
function cleanup() {
  port.disconnect();
}
```

### Service Worker Port Management

```javascript
// service-worker.js - Manage multiple port connections

const connectedPorts = new Map(); // portId -> port

chrome.runtime.onConnect.addListener((port) => {
  const portId = `${port.name}-${port.sender.tab?.id || 'extension'}`;
  connectedPorts.set(portId, port);

  console.log(`Port connected: ${portId}`);

  port.onMessage.addListener((message) => {
    handlePortMessage(port, portId, message);
  });

  port.onDisconnect.addListener(() => {
    connectedPorts.delete(portId);
    console.log(`Port disconnected: ${portId}`);
  });
});

function handlePortMessage(port, portId, message) {
  switch (message.action) {
    case 'subscribe':
      // Store subscription and send periodic updates
      port.postMessage({ type: 'subscribed', topic: message.topic });
      break;

    case 'request':
      // Process request and respond via port
      processRequest(message.data).then(result => {
        port.postMessage({ type: 'response', requestId: message.id, data: result });
      });
      break;
  }
}

// Broadcast to all connected ports
function broadcast(message) {
  for (const [portId, port] of connectedPorts) {
    try {
      port.postMessage(message);
    } catch (error) {
      connectedPorts.delete(portId);
    }
  }
}

// Send to specific content script tab
function sendToTab(tabId, message) {
  for (const [portId, port] of connectedPorts) {
    if (port.sender.tab?.id === tabId) {
      port.postMessage(message);
      return true;
    }
  }
  return false;
}
```

### Service Worker to Content Script Port

```javascript
// service-worker.js - Initiate connection to content script
async function connectToContentScript(tabId) {
  const port = chrome.tabs.connect(tabId, { name: 'sw-to-content' });

  port.onMessage.addListener((message) => {
    console.log(`Message from tab ${tabId}:`, message);
  });

  port.onDisconnect.addListener(() => {
    console.log(`Port to tab ${tabId} disconnected`);
  });

  return port;
}
```

## Popup and Side Panel Communication

Popup, side panel, and options pages are extension pages that can communicate with the service worker using the same chrome.runtime messaging APIs.

```javascript
// popup/popup.js or sidepanel/sidepanel.js
async function loadData() {
  const response = await chrome.runtime.sendMessage({
    action: 'get-data',
    source: 'popup'
  });
  displayData(response);
}

// Listen for updates from service worker
chrome.runtime.onMessage.addListener((message) => {
  if (message.target === 'popup') {
    updateUI(message);
  }
});

// Service worker sending to popup/side panel
// Note: popup must be open to receive messages
async function notifyPopup(data) {
  try {
    await chrome.runtime.sendMessage({
      target: 'popup',
      type: 'update',
      data
    });
  } catch {
    // Popup is closed, ignore error
  }
}
```

## Cross-Extension Messaging

Extensions can communicate with other extensions using chrome.runtime.sendMessage with a target extension ID. The receiving extension must have the sender listed in its externally_connectable manifest field, or use onMessageExternal without restrictions.

### Sending to Another Extension

```javascript
// sender-extension/service-worker.js
const TARGET_EXTENSION_ID = 'abcdefghijklmnopabcdefghijklmnop';

async function sendToOtherExtension(data) {
  const response = await chrome.runtime.sendMessage(
    TARGET_EXTENSION_ID,
    { action: 'shared-action', data }
  );
  return response;
}
```

### Receiving from Another Extension

```javascript
// receiver-extension/service-worker.js
chrome.runtime.onMessageExternal.addListener((message, sender, sendResponse) => {
  // Validate sender extension ID
  const allowedExtensions = ['abcdefghijklmnopabcdefghijklmnop'];
  if (!allowedExtensions.includes(sender.id)) {
    sendResponse({ error: 'Unauthorized extension' });
    return;
  }

  switch (message.action) {
    case 'shared-action':
      processSharedAction(message.data)
        .then(result => sendResponse({ success: true, data: result }))
        .catch(err => sendResponse({ error: err.message }));
      return true;
  }
});
```

### Cross-Extension Port Connection

```javascript
// Establish long-lived connection to another extension
const port = chrome.runtime.connect(TARGET_EXTENSION_ID, { name: 'cross-ext' });
port.postMessage({ greeting: 'hello' });

// Receive cross-extension port connections
chrome.runtime.onConnectExternal.addListener((port) => {
  console.log('External connection from:', port.sender.id);
  port.onMessage.addListener((message) => {
    handleExternalMessage(port, message);
  });
});
```

## Web Page to Extension Messaging

Web pages can send messages to extensions using chrome.runtime.sendMessage if the extension declares the page's origin in externally_connectable.matches.

### Extension Manifest Configuration

```json
{
  "externally_connectable": {
    "matches": ["https://*.example.com/*"]
  }
}
```

### Web Page Sending Message

```javascript
// On the web page (example.com)
// The extension ID is required
const EXTENSION_ID = 'abcdefghijklmnopabcdefghijklmnop';

chrome.runtime.sendMessage(EXTENSION_ID, {
  action: 'authenticate',
  token: 'user-token'
}, (response) => {
  if (response.success) {
    console.log('Authenticated with extension');
  }
});
```

### Extension Receiving Web Page Messages

```javascript
// service-worker.js
chrome.runtime.onMessageExternal.addListener((message, sender, sendResponse) => {
  // sender.url contains the page URL
  // sender.tab contains tab information
  // sender.id is undefined for web page senders

  // Validate origin
  const allowedOrigins = ['https://example.com', 'https://www.example.com'];
  const senderOrigin = new URL(sender.url).origin;

  if (!allowedOrigins.includes(senderOrigin)) {
    sendResponse({ error: 'Unauthorized origin' });
    return;
  }

  switch (message.action) {
    case 'authenticate':
      handleWebAuth(message.token, sender)
        .then(result => sendResponse({ success: true, data: result }))
        .catch(err => sendResponse({ success: false, error: err.message }));
      return true;
  }
});
```

## Error Handling

Robust error handling is essential for message passing since failures can occur from closed popups, unloaded content scripts, missing permissions, or extension updates.

```javascript
// Robust message sender with retry
async function sendMessageWithRetry(message, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await chrome.runtime.sendMessage(message);
      if (chrome.runtime.lastError) {
        throw new Error(chrome.runtime.lastError.message);
      }
      return response;
    } catch (error) {
      if (error.message.includes('Extension context invalidated')) {
        // Extension was updated, cannot recover
        throw error;
      }
      if (error.message.includes('Could not establish connection')) {
        // Receiving end does not exist, retry after delay
        if (attempt < maxRetries - 1) {
          await new Promise(r => setTimeout(r, 1000 * (attempt + 1)));
          continue;
        }
      }
      throw error;
    }
  }
}

// Safe tab messaging with existence check
async function safeTabMessage(tabId, message) {
  try {
    const tab = await chrome.tabs.get(tabId);
    if (!tab) return null;

    return await chrome.tabs.sendMessage(tabId, message);
  } catch (error) {
    if (error.message.includes('Could not establish connection')) {
      // Content script not loaded in this tab
      return null;
    }
    throw error;
  }
}
```

## Message Routing Pattern

For extensions with multiple message types and handlers, implement a centralized message router in the service worker.

```javascript
// service-worker.js - Message router pattern

const messageHandlers = new Map();

// Register handlers
function registerHandler(action, handler) {
  messageHandlers.set(action, handler);
}

// Central router
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  const handler = messageHandlers.get(message.action);

  if (!handler) {
    sendResponse({ error: `Unknown action: ${message.action}` });
    return false;
  }

  handler(message, sender)
    .then(result => sendResponse({ success: true, data: result }))
    .catch(error => sendResponse({ success: false, error: error.message }));

  return true; // Always async
});

// Register domain-specific handlers
registerHandler('fetch-data', async (msg) => {
  return await fetchFromAPI(msg.query);
});

registerHandler('save-settings', async (msg) => {
  await chrome.storage.local.set({ settings: msg.settings });
  return { saved: true };
});

registerHandler('get-tab-info', async (msg, sender) => {
  if (sender.tab) {
    return { tabId: sender.tab.id, url: sender.tab.url };
  }
  throw new Error('Not sent from a tab');
});
```

## Security Best Practices

Never trust messages from content scripts. The host page environment could be compromised. Always validate message structure, sanitize data, and avoid executing arbitrary code from messages.

Never use eval, Function constructor, or innerHTML with message data. Validate expected types and values before processing. Use an allowlist of accepted actions.

For cross-extension and web page messaging, always verify the sender identity using sender.id (extension ID) or sender.url (web page origin). Maintain an allowlist of trusted sources.
