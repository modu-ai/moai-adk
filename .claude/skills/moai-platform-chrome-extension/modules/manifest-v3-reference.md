---
name: manifest-v3-reference
description: Complete manifest.json field reference for Chrome Extension Manifest V3
parent-skill: moai-platform-chrome-extension
version: 1.0.0
updated: 2026-02-01
---

# Manifest V3 Reference

## Overview

The manifest.json file is the configuration backbone of every Chrome extension. Manifest V3 introduced significant architectural changes from V2 to improve security, privacy, and performance. This module provides comprehensive field reference and migration guidance.

## Required Fields

Every manifest.json must include exactly three required fields.

manifest_version must be set to integer 3. This is the only accepted value for new extensions submitted to the Chrome Web Store.

name specifies the extension display name shown in the Chrome toolbar, chrome://extensions, and the Chrome Web Store. Maximum length is 75 characters.

version uses a dot-separated string format with one to four integers. Examples: "1.0", "1.0.0", "1.0.0.1". Chrome uses this for update detection.

```json
{
  "manifest_version": 3,
  "name": "My Extension",
  "version": "1.0.0"
}
```

## Descriptive Fields

description provides a plain text summary displayed in chrome://extensions and the Chrome Web Store. Maximum length is 132 characters. Keep descriptions concise and informative about extension purpose.

icons specifies PNG images at multiple sizes. Chrome uses these in the toolbar, extensions page, and Chrome Web Store. Declare sizes 16, 32, 48, and 128. Only PNG format is supported.

```json
{
  "description": "Enhances browsing experience with productivity tools",
  "icons": {
    "16": "icons/icon16.png",
    "32": "icons/icon32.png",
    "48": "icons/icon48.png",
    "128": "icons/icon128.png"
  }
}
```

## Background Service Worker

The background field configures the extension service worker. The service_worker property must be a single string path to the JavaScript file. It cannot be an array unlike MV2 background scripts.

Set type to "module" to enable ES module imports in the service worker. This allows using import and export statements.

```json
{
  "background": {
    "service_worker": "service-worker.js",
    "type": "module"
  }
}
```

Migration Note from MV2: The MV2 background.scripts array and background.page fields are replaced by the single service_worker string. Persistent background pages no longer exist; all background logic must be event-driven.

## Action Configuration

The action field configures the toolbar button. In MV3, this unifies the former browserAction and pageAction APIs into a single surface.

default_popup specifies the HTML file displayed when the toolbar button is clicked. default_icon specifies the toolbar button icon as either a string path or an object with size keys. default_title sets the tooltip text shown on hover.

```json
{
  "action": {
    "default_popup": "popup.html",
    "default_icon": {
      "16": "icons/action16.png",
      "32": "icons/action32.png"
    },
    "default_title": "Click to open extension"
  }
}
```

Migration Note from MV2: Replace "browser_action" or "page_action" with "action". Replace chrome.browserAction or chrome.pageAction API calls with chrome.action.

## Content Scripts

The content_scripts field declares scripts and stylesheets injected into web pages. Each entry in the array specifies:

matches (required): URL match patterns determining which pages receive the injection. Uses Chrome match pattern syntax with scheme, host, and path components.

js: Array of JavaScript files injected in order.

css: Array of CSS files injected in order.

run_at: Controls injection timing. "document_start" injects before DOM construction. "document_end" injects after DOM is complete but before subresources load. "document_idle" (default) injects after page load.

all_frames: Boolean controlling whether scripts inject into all frames or only the top frame. Default is false.

match_about_blank: Boolean controlling whether scripts inject into about:blank frames whose parent matches the patterns. Default is false.

```json
{
  "content_scripts": [
    {
      "matches": ["https://*.example.com/*"],
      "js": ["content/main.js"],
      "css": ["content/styles.css"],
      "run_at": "document_end",
      "all_frames": false
    }
  ]
}
```

## Permissions

permissions declares standard API access requirements. Each string names a Chrome API or capability.

Common permissions:

- "storage" for chrome.storage API
- "tabs" for chrome.tabs API (query with URL access)
- "activeTab" for temporary access to the active tab on user gesture
- "contextMenus" for chrome.contextMenus API
- "notifications" for chrome.notifications API
- "scripting" for chrome.scripting API
- "alarms" for chrome.alarms API
- "sidePanel" for chrome.sidePanel API
- "declarativeNetRequest" for network request filtering
- "declarativeNetRequestWithHostAccess" for network filtering with host access
- "identity" for chrome.identity OAuth2 API
- "offscreen" for chrome.offscreen API

```json
{
  "permissions": ["storage", "activeTab", "scripting", "alarms"]
}
```

## Host Permissions

host_permissions declares URL patterns for web page access. These patterns determine which sites the extension can interact with via content scripts, fetch requests, and tab manipulation.

Use specific patterns when possible. Prefer activeTab over broad host patterns to minimize install-time warnings.

```json
{
  "host_permissions": [
    "https://*.example.com/*",
    "https://api.service.com/*"
  ]
}
```

Migration Note from MV2: Host permissions were previously included in the permissions array. MV3 separates them into their own field for clearer permission presentation to users.

## Optional Permissions

optional_permissions and optional_host_permissions declare permissions the extension can request at runtime. These do not trigger install-time warnings and can be requested via chrome.permissions.request when the user performs a relevant action.

```json
{
  "optional_permissions": ["bookmarks", "downloads"],
  "optional_host_permissions": ["https://*.otherdomain.com/*"]
}
```

## Web Accessible Resources

web_accessible_resources declares files within the extension package that web pages or other extensions can access. Each entry specifies resources (file paths with optional wildcards) and access conditions.

In MV3, access must be scoped to specific origins via matches, or to specific extensions via extension_ids, or to all web pages via use_dynamic_url.

```json
{
  "web_accessible_resources": [
    {
      "resources": ["images/*.png", "styles/injected.css"],
      "matches": ["https://*.example.com/*"]
    }
  ]
}
```

Migration Note from MV2: The MV2 flat array format is replaced with objects that scope resource access by origin, improving security.

## Content Security Policy

content_security_policy configures CSP for extension pages. The extension_pages property sets the policy applied to popup, options, and other extension HTML pages.

MV3 enforces strict defaults: no inline scripts, no eval, no remote code. The policy can only be made more restrictive, not less.

```json
{
  "content_security_policy": {
    "extension_pages": "script-src 'self' 'wasm-unsafe-eval'; object-src 'self'"
  }
}
```

## Side Panel

side_panel configures the extension side panel displayed alongside web content. default_path specifies the HTML file rendered in the panel.

```json
{
  "side_panel": {
    "default_path": "sidepanel.html"
  }
}
```

## Keyboard Shortcuts

commands declares keyboard shortcuts for extension actions. Each command has a suggested_key object with platform-specific key combinations and an optional description.

The special "_execute_action" command triggers the toolbar button action.

```json
{
  "commands": {
    "_execute_action": {
      "suggested_key": {
        "default": "Ctrl+Shift+Y",
        "mac": "Command+Shift+Y"
      }
    },
    "toggle-feature": {
      "suggested_key": {
        "default": "Ctrl+Shift+U"
      },
      "description": "Toggle the feature on/off"
    }
  }
}
```

## Declarative Net Request

declarative_net_request configures static network filtering rules. Declare rule_resources as an array of objects with id, enabled status, and path to the JSON rules file.

```json
{
  "declarative_net_request": {
    "rule_resources": [
      {
        "id": "blocking_rules",
        "enabled": true,
        "path": "rules/blocking.json"
      }
    ]
  }
}
```

Migration Note from MV2: Replace chrome.webRequest blocking handlers with declarative rules. This is a fundamental architectural change in MV3 that improves performance and privacy.

## Options Page

options_ui configures the extension settings page. The page property specifies the HTML file. Set open_in_tab to false to embed within chrome://extensions (recommended).

```json
{
  "options_ui": {
    "page": "options.html",
    "open_in_tab": false
  }
}
```

## DevTools Extension

devtools_page specifies an HTML file loaded when Chrome DevTools opens. This page can use the chrome.devtools API to create panels, inspect elements, and access network information.

```json
{
  "devtools_page": "devtools.html"
}
```

## External Messaging

externally_connectable declares which external websites and extensions can communicate with the extension via chrome.runtime.sendMessage and chrome.runtime.connect.

ids lists extension IDs allowed to send messages. matches lists URL patterns for web pages allowed to send messages (only second-level domains or higher).

```json
{
  "externally_connectable": {
    "ids": ["abcdefghijklmnopabcdefghijklmnop"],
    "matches": ["https://*.example.com/*"]
  }
}
```

## OAuth2 Authentication

oauth2 configures OAuth2 authentication for the chrome.identity API. Specify client_id from Google Cloud Console and scopes for requested access.

```json
{
  "oauth2": {
    "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
    "scopes": ["https://www.googleapis.com/auth/userinfo.email"]
  }
}
```

## Address Bar Keyword

omnibox registers a keyword that activates the extension when typed in the Chrome address bar.

```json
{
  "omnibox": {
    "keyword": "myext"
  }
}
```

## Complete MV2 to MV3 Migration Checklist

Background Migration:

- Replace background.scripts or background.page with background.service_worker as a single string
- Convert persistent state from global variables to chrome.storage
- Move setTimeout/setInterval to chrome.alarms
- Register all event listeners at the top level of the service worker

API Migration:

- Replace chrome.browserAction with chrome.action
- Replace chrome.pageAction with chrome.action
- Replace chrome.webRequest blocking with chrome.declarativeNetRequest
- Replace chrome.tabs.executeScript with chrome.scripting.executeScript
- Replace chrome.tabs.insertCSS with chrome.scripting.insertCSS

Permission Migration:

- Move host patterns from permissions to host_permissions
- Update web_accessible_resources from flat array to scoped objects

Security Migration:

- Remove all remote code loading (CDN scripts, eval-based templates)
- Bundle all JavaScript within the extension package
- Update CSP to MV3 format under content_security_policy.extension_pages

## Complete Manifest V3 Template

```json
{
  "manifest_version": 3,
  "name": "Extension Name",
  "version": "1.0.0",
  "description": "Brief extension description under 132 characters",
  "icons": {
    "16": "icons/icon16.png",
    "32": "icons/icon32.png",
    "48": "icons/icon48.png",
    "128": "icons/icon128.png"
  },
  "background": {
    "service_worker": "service-worker.js",
    "type": "module"
  },
  "action": {
    "default_popup": "popup.html",
    "default_icon": {
      "16": "icons/action16.png",
      "32": "icons/action32.png"
    },
    "default_title": "Extension tooltip"
  },
  "content_scripts": [
    {
      "matches": ["https://*.example.com/*"],
      "js": ["content/main.js"],
      "css": ["content/styles.css"],
      "run_at": "document_end"
    }
  ],
  "permissions": ["storage", "activeTab", "scripting"],
  "host_permissions": [],
  "optional_permissions": [],
  "web_accessible_resources": [
    {
      "resources": ["images/*"],
      "matches": ["https://*.example.com/*"]
    }
  ],
  "content_security_policy": {
    "extension_pages": "script-src 'self' 'wasm-unsafe-eval'; object-src 'self'"
  },
  "options_ui": {
    "page": "options.html",
    "open_in_tab": false
  }
}
```
