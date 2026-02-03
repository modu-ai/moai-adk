---
name: moai-platform-chrome-extension
description: >
  Chrome Extension Manifest V3 development specialist covering service workers,
  content scripts, message passing, chrome.* APIs, side panel, declarativeNetRequest,
  permissions model, and Chrome Web Store publishing. Use when building browser
  extensions, implementing content scripts, configuring service workers, or
  publishing to Chrome Web Store. [KO: 크롬 확장 프로그램, 매니페스트 V3, 서비스 워커,
  콘텐츠 스크립트] [JA: Chrome拡張機能、マニフェストV3、サービスワーカー]
  [ZH: Chrome扩展程序、Manifest V3、Service Worker]
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash(npm:*) Bash(npx:*) Bash(node:*) WebFetch WebSearch mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "platform"
  status: "active"
  updated: "2026-02-01"
  modularized: "true"
  tags: "chrome-extension, manifest-v3, service-worker, content-script, messaging, chrome-api, browser-extension, web-store, side-panel, declarative-net-request"
  context7-libraries: "/nicedoc/chrome-extension-doc"
  related-skills: "moai-lang-typescript, moai-lang-javascript, moai-domain-frontend"
  aliases: "chrome-ext, browser-extension, crx"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 120
  level2_tokens: 8000

# MoAI Extension: Triggers
triggers:
  keywords: ["chrome extension", "manifest v3", "service worker", "content script", "chrome api", "browser extension", "popup", "side panel", "background script", "web store", "declarativeNetRequest", "chrome.runtime", "chrome.tabs", "chrome.storage", "chrome.scripting", "chrome.action", "manifest.json", "crx"]
  agents: ["expert-frontend", "expert-backend"]
  phases: ["plan", "run"]
---

# Chrome Extension Manifest V3 Development

## Quick Reference

Chrome Extension Manifest V3 Development Specialist enables building modern browser extensions with the latest Chrome platform APIs.

Auto-Triggers: Chrome extension projects detected via manifest.json with manifest_version 3, service worker files, content script declarations, chrome API usage patterns

### Core Capabilities

Manifest V3 Platform:

- Service workers replace persistent background pages for event-driven architecture
- Remote code execution removed for enhanced security
- declarativeNetRequest replaces blocking webRequest for network filtering
- Promise-based API methods across all chrome.* APIs
- action API unifies browserAction and pageAction into single surface
- Supported in Chrome 88 and later

Process Architecture:

- Service worker runs as single event-driven background script terminating when idle
- Content scripts execute in web page context within isolated worlds
- Popup and side panel provide dedicated UI surfaces
- Options page provides extension settings interface
- DevTools panel extends Chrome Developer Tools

Communication Patterns:

- One-time messages between service worker and content scripts via sendMessage
- Long-lived connections via connect with port-based communication
- Cross-extension messaging through externally_connectable declaration
- Web page to extension messaging for verified origins

Security Model:

- Content Security Policy restricts script sources to self only
- No inline scripts or remote code execution permitted
- Permissions declare required API access at install time
- Optional permissions allow runtime-requested access with user consent
- Host permissions control web page access patterns

### Context7 Documentation Access

For latest Chrome Extension API documentation, use the Context7 MCP tools:

Step 1 - Resolve library ID: Use mcp__context7__resolve-library-id with query "chrome extension" to get the Context7-compatible library ID.

Step 2 - Fetch documentation: Use mcp__context7__get-library-docs with the resolved library ID, specifying topic and token allocation.

Example topics include "manifest v3 configuration", "service worker lifecycle", "content scripts injection", "message passing patterns", "chrome.storage API", "side panel API", and "declarativeNetRequest rules".

---

## Module Index

This skill uses progressive disclosure with specialized modules for detailed implementation patterns.

### Core Modules

manifest-v3-reference covers the complete manifest.json field reference for Manifest V3 extensions. Topics include required and optional fields, field types and constraints, permission declarations, MV2 to MV3 migration notes, and extension configuration best practices.

service-worker-patterns covers service worker lifecycle, event registration, state management, and debugging. Topics include event-driven architecture, top-level listener registration, state persistence with chrome.storage, keep-alive strategies, offscreen documents for DOM access, and debugging with chrome://extensions.

content-scripts-guide covers content script injection methods, isolated worlds, and communication. Topics include static declaration in manifest, dynamic registration with chrome.scripting, programmatic injection, isolated world architecture, DOM access patterns, and security considerations.

messaging-patterns covers message passing between extension components. Topics include one-time messages with sendMessage, long-lived connections with connect and ports, async response patterns, cross-extension messaging, web page messaging, and error handling strategies.

apis-quick-reference covers the major chrome.* APIs with method signatures and permission requirements. Topics include chrome.runtime, chrome.tabs, chrome.storage, chrome.action, chrome.scripting, chrome.alarms, chrome.notifications, chrome.contextMenus, chrome.sidePanel, chrome.declarativeNetRequest, chrome.offscreen, chrome.identity, and chrome.commands.

ui-components covers popup, side panel, options page, DevTools panel, and content script UI. Topics include popup HTML and lifecycle, side panel configuration and API, options page patterns, DevTools extension integration, and injected UI from content scripts.

security-csp covers Content Security Policy, permissions model, and secure coding practices. Topics include CSP configuration for extension pages, minimum privilege permissions, input validation, XSS prevention, secure messaging patterns, and HTTPS enforcement.

publishing-guide covers Chrome Web Store submission and distribution. Topics include developer account setup, extension packaging, privacy policy requirements, review process, update mechanisms, and self-hosted distribution.

---

## Implementation Guide

### Manifest V3 Structure

Every Chrome extension requires a manifest.json file at the project root. Three fields are mandatory: manifest_version set to integer 3, name as the extension display name with maximum 75 characters, and version as a semver-compatible string.

The description field provides a summary shown in Chrome Web Store with maximum 132 characters. The icons object specifies PNG icons at 16, 32, 48, and 128 pixel sizes for various Chrome UI contexts.

For background processing, declare a service_worker field inside the background object as a single string path pointing to the service worker file. Set type to module when using ES module imports. The service worker path must be a single string, not an array.

Content scripts are declared as an array of objects, each specifying matches patterns for URL matching, js array for JavaScript files, optional css array for stylesheets, and run_at to control injection timing with values document_start, document_end, or document_idle.

For detailed field reference and migration guidance, see modules/manifest-v3-reference.md.

### Service Worker Architecture

Service workers in Manifest V3 replace persistent background pages with an event-driven model. The service worker runs only when responding to events and terminates when idle, reducing memory and CPU consumption.

All event listeners must be registered at the top level of the service worker script. Listeners registered inside callbacks, promises, or async functions will not persist across service worker restarts.

Since service workers have no DOM access, no window object, and no localStorage, use chrome.storage API for persistent state. Use the Alarms API for scheduled tasks instead of setTimeout or setInterval, as these timers do not survive service worker termination. Use fetch for network requests instead of XMLHttpRequest.

For long-running operations that require DOM access, use the Offscreen Documents API via chrome.offscreen.createDocument to create a hidden document with DOM capabilities.

For complete service worker patterns and debugging guidance, see modules/service-worker-patterns.md.

### Content Scripts

Content scripts execute JavaScript and CSS in the context of web pages. They run in isolated worlds, meaning they share DOM access with the host page but have separate JavaScript execution environments, preventing variable and function conflicts.

Three injection methods exist. Static injection declares scripts in the manifest content_scripts array with URL match patterns. Dynamic injection uses chrome.scripting.registerContentScripts for runtime registration. Programmatic injection uses chrome.scripting.executeScript to inject on demand, requiring either host_permissions or activeTab permission.

Content scripts have limited direct chrome API access: only dom, i18n, storage, and specific runtime methods including connect, sendMessage, onMessage, onConnect, getManifest, getURL, and id. All other API calls must go through message passing to the service worker.

For injection patterns, isolated world details, and security considerations, see modules/content-scripts-guide.md.

### Message Passing Patterns

Extensions communicate between components using Chrome message passing. One-time messages use chrome.runtime.sendMessage to reach the service worker and chrome.tabs.sendMessage to reach content scripts. Each message receives a single response through a callback or Promise.

Long-lived connections use chrome.runtime.connect or chrome.tabs.connect to establish ports. Ports remain open until either side calls disconnect, a listener is removed, or the containing tab unloads. Ports support ongoing bidirectional communication.

For async responses in one-time messaging, the onMessage listener must either return true to indicate an async sendResponse call, or return a Promise directly starting from Chrome 144.

All messages use JSON serialization with a maximum size of 64 MiB. Never trust message content from content scripts as the host page context could be compromised.

For complete messaging patterns including cross-extension and web page communication, see modules/messaging-patterns.md.

### Chrome APIs Reference

The chrome.runtime API provides extension lifecycle management, messaging, and manifest access. It handles installation, update, and suspend events, and provides methods for getting extension URLs and platform information.

The chrome.tabs API manages browser tabs with methods for querying, creating, updating, and removing tabs. The chrome.storage API provides three storage areas: local with 10 MB capacity, sync with 100 KB that synchronizes across signed-in devices, and session for in-memory storage that clears on browser restart.

The chrome.action API controls the toolbar button including badge text, icon, popup, and click handlers. The chrome.scripting API provides programmatic script and CSS injection into web pages.

The chrome.sidePanel API manages the extension side panel, a persistent UI surface alongside web content. The chrome.declarativeNetRequest API provides network request filtering using static and dynamic rules without blocking webRequest.

For complete API method signatures and permission requirements, see modules/apis-quick-reference.md.

### UI Components

Extensions support multiple UI surfaces. The popup is configured via action.default_popup in the manifest and displays as a standard HTML page when the toolbar button is clicked. Popups close when they lose focus and should load quickly.

The side panel is configured via side_panel.default_path in the manifest and provides a persistent panel alongside web content. The chrome.sidePanel API controls panel behavior, enabling per-tab or global panels.

The options page is configured via options_ui.page in the manifest and opens within chrome://extensions for extension settings. The DevTools panel extends Chrome Developer Tools using the devtools_page manifest field.

Content scripts can inject UI elements directly into web pages using DOM manipulation, applying custom CSS for styling.

For detailed UI implementation patterns, see modules/ui-components.md.

### Permissions Model

Permissions fall into four categories. Standard permissions declare API access requirements such as storage, tabs, activeTab, contextMenus, notifications, scripting, alarms, sidePanel, declarativeNetRequest, identity, and offscreen. These are granted at install time.

Host permissions specify URL patterns for web page access using patterns like https://*.example.com/*. Optional permissions and optional host permissions allow runtime requests through chrome.permissions.request, reducing the install-time permission prompt.

Prefer activeTab over broad host permissions to minimize permission warnings. Request only the minimum permissions necessary for extension functionality.

For detailed permission strategies and security guidance, see modules/security-csp.md.

### Security Best Practices

The Content Security Policy for Manifest V3 restricts script-src to self and wasm-unsafe-eval only. Inline scripts, eval, and remote code loading are prohibited. All JavaScript must be bundled within the extension package.

Content scripts run in isolated worlds but should be treated as potentially compromised since the host page can manipulate the shared DOM. Always validate and sanitize data received from content scripts in the service worker. Never use eval, document.write, or innerHTML with untrusted data. Use HTTPS for all external network requests.

For comprehensive security patterns and CSP configuration, see modules/security-csp.md.

---

## Advanced Patterns

For detailed implementation guidance on advanced topics, see the modules directory:

Manifest V3 Migration:

- Converting MV2 background pages to service workers
- Replacing blocking webRequest with declarativeNetRequest
- Updating remote code to bundled modules
- Adapting persistent state to chrome.storage patterns

Complex Service Worker Patterns:

- Multi-alarm scheduling for periodic tasks
- Service worker keep-alive for long operations
- Offscreen document management for audio, canvas, and DOM parsing
- Shared module imports across service worker and content scripts

Advanced Content Script Patterns:

- Dynamic script registration based on user preferences
- World isolation strategies for main world versus isolated world
- Shadow DOM injection for encapsulated UI components
- MutationObserver patterns for dynamic page content

Cross-Context Communication:

- Message routing between multiple content scripts
- Broadcast patterns to all tabs
- External website to extension communication
- Native messaging with local applications via chrome.runtime.connectNative

Storage Synchronization:

- chrome.storage.sync for cross-device settings
- chrome.storage.session for temporary data
- Storage change listeners for reactive updates
- Quota management and overflow strategies

---

## Troubleshooting

Common Issues and Solutions:

Service Worker Not Registering:

Verify the background.service_worker field in manifest.json is a single string path, not an array. Ensure the service worker file exists at the declared path. Check chrome://extensions for error messages on the extension card. Inspect the service worker console by clicking the service worker link on the extension details page.

Content Script Not Injecting:

Confirm the matches patterns in manifest.json correctly target the desired URLs. Verify run_at timing is appropriate for the page content being accessed. Check that the extension has the necessary host permissions. Inspect the target page console for content script errors.

Message Passing Failures:

Ensure channel names and message structures match between sender and receiver. Verify the receiving listener is registered before messages are sent. Check that sendResponse is called before the listener returns, or return true for async responses. Verify the target tab exists when using chrome.tabs.sendMessage.

Permission Denied Errors:

Confirm all required permissions are declared in manifest.json. For programmatic injection, verify either host_permissions or activeTab permission is granted. Check chrome://extensions for any permission warnings or disabled states. Use chrome.permissions.contains to verify runtime permissions.

Extension Not Appearing in Chrome Web Store:

Ensure manifest.json passes validation with no errors. Verify all declared resources including icons, HTML files, and scripts exist in the package. Check that the description does not exceed 132 characters. Review the developer dashboard for submission errors.

Debug Commands:

Open chrome://extensions to view all installed extensions and their status. Enable developer mode to access extension details and error logs. Click the service worker link to open its dedicated DevTools console. Use chrome://inspect to debug content scripts in page context.

---

## Works Well With

- moai-lang-typescript for TypeScript patterns in extension development
- moai-lang-javascript for JavaScript patterns and ES module usage
- moai-domain-frontend for React or framework-based popup and side panel UI
- moai-domain-backend for server-side API integration
- moai-workflow-testing for extension testing strategies

---

## Resources

### Module References

For detailed implementation patterns, see the modules directory:

- modules/manifest-v3-reference.md covers complete manifest.json field reference
- modules/service-worker-patterns.md covers service worker lifecycle and patterns
- modules/content-scripts-guide.md covers content script injection and communication
- modules/messaging-patterns.md covers all message passing patterns
- modules/apis-quick-reference.md covers chrome.* API method signatures
- modules/ui-components.md covers popup, side panel, options, and DevTools UI
- modules/security-csp.md covers CSP, permissions, and secure coding
- modules/publishing-guide.md covers Chrome Web Store publishing workflow

### External Documentation

For latest documentation, use Context7 to query:

- /nicedoc/chrome-extension-doc for Chrome Extension APIs

For official Chrome documentation, use WebFetch with:

- https://developer.chrome.com/docs/extensions/develop for development guides
- https://developer.chrome.com/docs/extensions/reference/api for API reference

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v1.0
Last Updated: 2026-02-01
Version: 1.0.0 (Initial Release)
Coverage: Manifest V3, Service Workers, Content Scripts, Messaging, Chrome APIs, UI, Security, Publishing
