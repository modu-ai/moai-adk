---
name: moai-framework-electron
description: >
  Electron 33+ desktop app development specialist covering Main/Renderer
  process architecture, IPC communication, auto-update, packaging with
  Electron Forge and electron-builder, and security best practices. Use when
  building cross-platform desktop applications, implementing native OS
  integrations, or packaging Electron apps for distribution. [KO: Electron
  데스크톱 앱, 크로스플랫폼 개발, IPC 통신] [JA: Electronデスクトップアプリ、
  クロスプラットフォーム開発] [ZH: Electron桌面应用、跨平台开发]
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "framework"
  status: "active"
  updated: "2026-01-10"
  modularized: "false"
  tags: "electron, desktop, cross-platform, nodejs, chromium, ipc, auto-update, electron-builder, electron-forge"
  context7-libraries: "/electron/electron, /electron/forge, /electron-userland/electron-builder"
  related-skills: "moai-lang-typescript, moai-domain-frontend, moai-lang-javascript"
---

# Electron 33+ Desktop Development

## Quick Reference

Electron 33+ Desktop App Development Specialist enables building cross-platform desktop applications with web technologies.

Auto-Triggers: Electron projects detected via electron.vite.config.ts or electron-builder.yml files, desktop app development requests, IPC communication pattern implementation

### Core Capabilities

Electron 33 Platform:

- Chromium 130 rendering engine for modern web features
- Node.js 20.18 runtime for native system access
- Native ESM support in main process
- WebGPU API support for GPU-accelerated graphics

Process Architecture:

- Main process runs as single instance per application with full Node.js access
- Renderer processes display web content in sandboxed environments
- Preload scripts bridge main and renderer with controlled API exposure
- Utility processes handle background tasks without blocking UI

IPC Communication:

- Type-safe invoke/handle patterns for request-response communication
- contextBridge API for secure renderer access to main process functionality
- Event-based messaging for push notifications from main to renderer

Auto-Update Support:

- electron-updater integration with GitHub and S3 publishing
- Differential updates for smaller download sizes
- Update notification and installation management

Packaging Options:

- Electron Forge for integrated build tooling and plugin ecosystem
- electron-builder for flexible multi-platform distribution

Security Features:

- contextIsolation for preventing prototype pollution
- Sandbox enforcement for renderer process isolation
- Content Security Policy configuration
- Input validation patterns for IPC handlers

### Project Initialization

Creating a new Electron application requires running the create-electron-app command with the vite-typescript template. Install electron-builder as a development dependency for packaging. Add electron-updater as a runtime dependency for auto-update functionality.

For detailed commands and configuration, see reference.md Quick Commands section.

---

## Implementation Guide

### Project Structure

Recommended Directory Layout:

The source directory should contain four main subdirectories:

Main Directory: Contains the main process entry point, IPC handler definitions organized by domain, business logic services, and window management modules

Preload Directory: Contains the preload script entry point and exposed API definitions that bridge main and renderer

Renderer Directory: Contains the web application built with React, Vue, or Svelte, including the HTML entry point and Vite configuration

Shared Directory: Contains TypeScript types and constants shared between main and renderer processes

The project root should include the electron.vite.config.ts for build configuration, electron-builder.yml for packaging options, and resources directory for app icons and assets.

### Main Process Setup

Application Lifecycle Management:

The main process initialization follows a specific sequence. First, enable sandbox globally using app.enableSandbox() to ensure all renderer processes run in isolated environments. Then request single instance lock to prevent multiple app instances from running simultaneously.

Window creation should occur after the app ready event fires. Configure BrowserWindow with security-focused webPreferences including contextIsolation enabled, nodeIntegration disabled, sandbox enabled, and webSecurity enabled. Set the preload script path to expose safe APIs to the renderer.

Handle platform-specific behaviors: on macOS, re-create windows when the dock icon is clicked if no windows exist. On other platforms, quit the application when all windows close.

For implementation examples, see examples.md Main Process Entry Point section.

### Type-Safe IPC Communication

IPC Type Definition Pattern:

Define an interface that maps channel names to their payload types. Group channels by domain such as file operations, window operations, and storage operations. This enables type checking for both main process handlers and renderer invocations.

Main Process Handler Registration:

Register IPC handlers in a dedicated module that imports from the shared types. Each handler should validate input using a schema validation library such as Zod before processing. Use ipcMain.handle for request-response patterns and return structured results.

Preload Script Implementation:

Create an API object that wraps ipcRenderer.invoke calls for each channel. Use contextBridge.exposeInMainWorld to make this API available in the renderer as window.electronAPI. Include cleanup functions for event listeners to prevent memory leaks.

For complete IPC implementation patterns, see examples.md Type-Safe IPC Implementation section.

### Security Best Practices

Mandatory Security Settings:

Every BrowserWindow must have webPreferences configured with four critical settings. contextIsolation must always be enabled to prevent renderer code from accessing Electron internals. nodeIntegration must always be disabled in renderer processes. sandbox must always be enabled for process-level isolation. webSecurity must never be disabled to maintain same-origin policy enforcement.

Content Security Policy:

Configure session-level CSP headers using webRequest.onHeadersReceived. Restrict default-src to self, script-src to self without unsafe-inline, and connect-src to allowed API domains. This prevents XSS attacks and unauthorized resource loading.

Input Validation:

Every IPC handler must validate inputs before processing. Prevent path traversal attacks by rejecting paths containing parent directory references. Validate file names against reserved characters. Use allowlists for permitted directories when implementing file access.

For security implementation details, see reference.md Security Best Practices section.

### Auto-Update Implementation

Update Service Architecture:

Create an UpdateService class that manages the electron-updater lifecycle. Initialize with the main window reference to enable UI notifications. Configure autoDownload as false to give users control over bandwidth usage.

Event Handling:

Handle update-available events by notifying the renderer and prompting the user for download confirmation. Track download-progress events to display progress indicators. Handle update-downloaded events by prompting for restart.

User Notification Pattern:

Use system dialogs to prompt users when updates are available and when downloads complete. Send events to the renderer for in-app notification display. Support both immediate and deferred installation.

For complete update service implementation, see examples.md Auto-Update Integration section.

### App Packaging

Electron Builder Configuration:

Configure the appId with reverse-domain notation for platform registration. Specify productName for display in system UI. Set up platform-specific targets for macOS, Windows, and Linux.

macOS Configuration:

Set category for App Store classification. Enable hardenedRuntime and configure entitlements for notarization. Configure universal builds targeting both x64 and arm64 architectures.

Windows Configuration:

Specify icon path for executable and installer. Configure NSIS installer options including installation directory selection. Set up code signing with appropriate hash algorithms.

Linux Configuration:

Configure category for desktop environment integration. Set up multiple targets including AppImage for universal distribution and deb/rpm for package manager installation.

For complete configuration examples, see reference.md Configuration section.

---

## Advanced Patterns

For comprehensive documentation on advanced topics, see reference.md and examples.md:

Window State Persistence:

- Saving and restoring window position and size across sessions
- Handling multiple displays and display changes
- Managing maximized and fullscreen states

Multi-Window Management:

- Creating secondary windows with proper parent-child relationships
- Sharing state between multiple windows
- Coordinating window lifecycle events

System Tray and Native Menus:

- Creating and updating system tray icons with context menus
- Building application menus with keyboard shortcuts
- Platform-specific menu patterns for macOS and Windows

Utility Processes:

- Spawning utility processes for CPU-intensive background tasks
- Communicating with utility processes via MessageChannel
- Managing utility process lifecycle and error handling

Native Module Integration:

- Rebuilding native modules for Electron Node.js version
- Using better-sqlite3 for local database storage
- Integrating keytar for secure credential storage

Protocol Handlers and Deep Linking:

- Registering custom URL protocols for app launching
- Handling deep links on different platforms
- OAuth callback handling via custom protocols

Performance Optimization:

- Lazy loading windows and heavy modules
- Optimizing startup time with deferred initialization
- Memory management for long-running applications

---

## Works Well With

- moai-lang-typescript - TypeScript patterns for type-safe Electron development
- moai-domain-frontend - React, Vue, or Svelte renderer development
- moai-lang-javascript - Node.js patterns for main process
- moai-domain-backend - Backend API integration
- moai-workflow-testing - Testing strategies for desktop apps

---

## Troubleshooting

Common Issues and Solutions:

White Screen on Launch:

Verify the preload script path is correctly configured relative to the built output directory. Check that loadFile or loadURL paths point to existing files. Enable devTools to inspect console errors. Review CSP settings that may block script execution.

IPC Not Working:

Confirm channel names match exactly between main handlers and renderer invocations. Ensure handlers are registered before windows load content. Verify contextBridge usage follows the correct pattern with exposeInMainWorld.

Native Modules Fail:

Run electron-rebuild after npm install to recompile native modules. Match the Node.js version embedded in Electron. Add a postinstall script to automate rebuilding.

Auto-Update Not Working:

Verify the application is code-signed as updates require signing. Check publish configuration in electron-builder.yml. Enable electron-updater logging to diagnose connection issues. Review firewall settings that may block update checks.

Debug Commands:

Rebuild native modules with npx electron-rebuild. Check Electron version with npx electron --version. Enable verbose update logging with DEBUG=electron-updater environment variable.

---

## Resources

For complete code examples and configuration templates, see:

- reference.md - Detailed API documentation, version matrix, Context7 library mappings
- examples.md - Production-ready code examples for all patterns

For latest documentation, use Context7 to query:

- /electron/electron for core Electron APIs
- /electron/forge for Electron Forge tooling
- /electron-userland/electron-builder for packaging configuration

---

Version: 2.0.0
Last Updated: 2026-01-10
Changes: Restructured to comply with CLAUDE.md Documentation Standards - removed all code examples, converted to narrative text format
