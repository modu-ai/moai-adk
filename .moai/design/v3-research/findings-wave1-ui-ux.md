# Wave 1.4 — Claude Code UI/UX Inventory

Research scope: Claude Code TypeScript source dump at `/Users/goos/MoAI/AgentOS/claude-code-source-map/`. Exhaustively catalogs the terminal UI engine (Ink fork), the 146-item component library, interactive flow helpers, dialog/REPL launchers, keybinding system, output styles, and permission UI flow. Closes with a gap matrix comparing moai-adk-go's current (non-TUI) posture against Claude Code's rich TUI — so the v3 plan knows which patterns moai-adk can safely surface in CC and which are CC-exclusive.

All file references are absolute paths. Line numbers are suffixed with `:N` (single line) or `:N-M` (range).

---

## 1. Ink TUI Stack

Claude Code does NOT ship against upstream `ink` on npm. It vendors a heavily customized fork at `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/` (50 top-level items, ~750 KB of source), then re-exports a trimmed public surface through `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink.ts`. Understanding this forked layout is load-bearing: moai-adk-go (a Go CLI, no TUI) will not reimplement it, but its output must render cleanly inside CC's Ink tree.

### 1.1 Entry Point — `ink.ts` (86 lines)

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink.ts`

- `:18-23` `render(node, options)` — async wrapper around `inkRender` that pre-wraps every mount in a `<ThemeProvider>` so `ThemedBox`/`ThemedText` subtrees never have to mount theme context themselves. Quote at `:12-16`:
  > "Wrap all CC render calls with ThemeProvider so ThemedBox/ThemedText work without every call site having to mount it. Ink itself is theme-agnostic."
- `:25-31` `createRoot(options)` — async factory for a reusable Ink root; returns a shim where `render(node)` auto-wraps with theme.
- `:33-37` Re-exports design-system primitives under the Ink names: `Box` is actually `ThemedBox`, `Text` is actually `ThemedText`. This is why component code throughout the repo imports `{ Box, Text } from '../ink.js'` and gets theme resolution automatically.
- `:44-85` Re-exports the base (non-themed) primitives as `BaseBox`, `BaseText`, plus the full event/hook surface:
  - Event classes: `ClickEvent`, `EventEmitter`, `Event`, `InputEvent`, `TerminalFocusEvent`
  - Hooks: `useAnimationFrame`, `useApp`, `useInput`, `useAnimationTimer`, `useInterval`, `useSelection`, `useStdin`, `useTabStatus`, `useTerminalFocus`, `useTerminalTitle`, `useTerminalViewport`
  - Primitives: `Button`, `Link`, `Newline`, `NoSelect`, `RawAnsi`, `Spacer`
  - Utilities: `Ansi`, `FocusManager`, `color`, `measureElement`, `wrapText`, `supportsTabStatus`

### 1.2 Forked Ink Internals — `ink/` (50 items)

Directory listing shows this is a purpose-built terminal renderer, not a thin wrapper. Key files by size (most interesting first):

| File | Size | Role |
|---|---|---|
| `ink/ink.tsx` | 252 KB | Core reconciler + render loop (the Ink React fork entrypoint) |
| `ink/render-node-to-output.ts` | 63 KB | Translates Yoga layout tree into ANSI screen buffer |
| `ink/screen.ts` | 49 KB | Screen buffer abstraction (double-buffered, diffed, synchronized-output-aware) |
| `ink/selection.ts` | 35 KB | Text selection model (mouse-drag copy, shift+arrow extend) |
| `ink/Ansi.tsx` | 33 KB | `<Ansi>` component for raw ANSI passthrough with styling |
| `ink/output.ts` | 26 KB | Final buffer flushing with BSU/ESU (DEC 2026) synchronized output |
| `ink/log-update.ts` | 27 KB | Terminal frame update primitive (cursor save/restore) |
| `ink/parse-keypress.ts` | 23 KB | Keyboard input parser supporting kitty keyboard protocol |
| `ink/styles.ts` | 21 KB | Style resolution (Yoga props → internal style) |

Component subdirectory `ink/components/` holds the base primitives:

- `AlternateScreen.tsx` (10 KB, 374 lines) — Toggles terminal alt-screen buffer for fullscreen mode. `:33` exports `function AlternateScreen`. Enables mouse tracking (mode 1003) and makes `onClick`, `onMouseEnter`, `onMouseLeave` work on `<Box>`.
- `App.tsx` (98 KB) — Top-level provider wiring (stdin/stdout contexts, focus manager, terminal size, clock, cursor, terminal focus).
- `Box.tsx` (22 KB) — Layout primitive. See `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/components/Box.tsx:11-46` for prop surface: accepts every Yoga `Styles` prop plus `tabIndex`, `autoFocus`, `onClick`, `onFocus`/`onBlur`, `onKeyDown`, `onMouseEnter`/`onMouseLeave`. Click bubbles from deepest hit up; `stopImmediatePropagation()` halts.
- `Button.tsx` (17 KB) — Keyboard-/click-activated button with focus/hover/active state machine. `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/components/Button.tsx:10-38` — children can be a render prop `(state: ButtonState) => ReactNode` so callers style focused/hovered/active themselves (Button is intentionally unstyled).
- `ClockContext.tsx` (12 KB) — Context-provided millisecond timer. Drives spinner frames, shimmer animations, waveform cursors.
- `ErrorOverview.tsx` (15 KB) — Top-level error boundary renderer.
- `Link.tsx` (3.5 KB) — Clickable hyperlink (OSC-8 aware). Falls back to styled text on terminals without hyperlink support.
- `Newline.tsx` (2.3 KB) — Explicit line break (Ink requires this, raw `\n` is stripped).
- `NoSelect.tsx` (5.9 KB) — Suppresses text selection in subtree (used for decorative borders, spinner glyphs).
- `RawAnsi.tsx` (5.2 KB) — Bypasses Ink's ANSI parser — outputs pre-rendered ANSI strings directly (used by `StructuredDiff` and `HighlightedCode` for already-colorized content).
- `ScrollBox.tsx` (32 KB) — Vertical scrollable region with handle (`ScrollBoxHandle` at `:16-28`) — `scrollTo`, `scrollBy`, `scrollIntoView`, page-up/down. Core of the transcript viewer.
- `Spacer.tsx` (1.5 KB) — Flex growth filler (`flexGrow: 1`).
- `TerminalFocusContext.tsx` (5.9 KB) — Knows when the terminal window itself has OS focus; exposed via `useTerminalFocus`.
- `TerminalSizeContext.tsx` (1 KB) — `{ rows, columns }` derived from SIGWINCH + initial query.

### 1.3 Layout Engine

- `ink/layout/yoga.ts` (7.4 KB) — Facebook Yoga flexbox binding. Yoga is the flexbox layout algorithm; Claude Code compiles Yoga styles to a row/column grid of cells.
- `ink/layout/engine.ts`, `ink/layout/geometry.ts`, `ink/layout/node.ts` — tree primitives and geometry (clip, clamp, intersect).
- `ink/get-max-width.ts`, `ink/measure-element.ts`, `ink/measure-text.ts`, `ink/line-width-cache.ts`, `ink/stringWidth.ts` — width calculation (CJK double-width, ANSI-ignore, emoji-aware).

### 1.4 Terminal I/O — `ink/termio/` (11 files)

Low-level escape sequence handling:

- `termio/ansi.ts`, `csi.ts`, `sgr.ts`, `osc.ts`, `dec.ts`, `esc.ts` — ANSI/CSI/SGR/OSC/DEC escape sequence emitters
- `termio/parser.ts`, `termio/tokenize.ts`, `termio/types.ts` — Incremental parser for terminal responses (for things like terminal size queries, cursor position reports)

Notable features:
- DEC 2026 synchronized output (BSU/ESU) — atomic frame updates that prevent flicker on fast-updating screens.
- OSC 52 clipboard integration.
- Kitty keyboard protocol (super modifier, disambiguated escape).

### 1.5 Events & Hooks — `ink/events/` (10 files), `ink/hooks/` (12 files)

Events (DOM-like synthetic events inside Ink tree):
- `click-event.ts`, `focus-event.ts`, `input-event.ts`, `keyboard-event.ts`, `terminal-event.ts`, `terminal-focus-event.ts`
- `dispatcher.ts`, `emitter.ts`, `event-handlers.ts`, `event.ts` — DOM event dispatch model with capture/bubble

Hooks used throughout the codebase:
- `use-animation-frame.ts` (1.9 KB) — callback-driven frame loop (rAF equivalent)
- `use-app.ts` (251 B) — access to Ink's `exit()` function
- `use-declared-cursor.ts` (3.0 KB) — Component declares its cursor-position intent for the terminal querier
- `use-input.ts` (3.1 KB) — Keyboard input subscription (the workhorse; every keybinding flows through this)
- `use-interval.ts`, `use-search-highlight.ts`, `use-selection.ts`, `use-stdin.ts`, `use-tab-status.ts`, `use-terminal-focus.ts`, `use-terminal-title.ts`, `use-terminal-viewport.ts`

### 1.6 Rich Content Primitives

- **ANSI passthrough**: `<Ansi>` (ink/Ansi.tsx, 33 KB) parses ANSI-styled strings into the Yoga tree; `<RawAnsi>` bypasses parsing for already-validated ANSI (high-performance path for diff/code views).
- **Syntax highlighting**: `components/HighlightedCode.tsx` (17 KB) + `components/HighlightedCode/Fallback.tsx` — lazy-loads a native `ColorFile` module (`expectColorFile()` at `:40`). Falls back to plain text when syntax highlight lib is unavailable or `syntaxHighlightingDisabled` setting.
- **Markdown**: `components/Markdown.tsx` (28 KB) uses `marked` to lex, then hybrid-renders: tables become React flexbox via `MarkdownTable.tsx` (47 KB); other tokens are rendered as pre-styled ANSI via `utils/markdown.formatToken`. Aggressive LRU token cache at `Markdown.tsx:22-71` (TOKEN_CACHE_MAX=500, keyed by content hash) because `marked.lexer` is hot (~3 ms) on virtual scroll remounts.
- **Tables**: `components/MarkdownTable.tsx` (47 KB) — flexbox-based; handles alignment, wrapping, multi-line cells.
- **Diffs**: `components/StructuredDiff.tsx` (25 KB) + `components/StructuredDiff/Fallback.tsx` — invokes a NAPI (native) colorDiff module via `expectColorDiff()`, caches rendered ANSI in `WeakMap<StructuredPatchHunk, Map<string, CachedRender>>` at `:32-41` (keyed by patch identity so scrolling back doesn't re-highlight). Gutter width computed from max line number.
- **Images**: `components/ClickableImageRef.tsx` (7.5 KB) + `components/messages/UserImageMessage.tsx` — Kitty image protocol support for terminals that speak it; fallback is text reference `[image: 1024x768, cached at /path]`.
- **No diagram/SVG embedding**: confirmed absent — Claude Code does not render Mermaid or SVG inline. Mermaid is rendered via Playwright MCP server (separate from TUI), images only shown as Kitty protocol blobs or refs.

---

## 2. Component Categories (with counts)

Total: 146 top-level items under `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/`. Enumerated here by functional category. Counts reflect distinct files; subdirectories with additional nested files are called out.

### 2.1 Dialogs (33 components, ~650 KB of JSX)

These are modal or modal-like surfaces launched via `showSetupDialog()` (see Section 4). Each implements its own cancel/confirm keybinding via the `Confirmation` context.

| Component | Size | Purpose |
|---|---|---|
| `ApproveApiKey.tsx` | 11 KB | New ANTHROPIC_API_KEY approval |
| `AutoModeOptInDialog.tsx` | 13 KB | Opt-in for transcript classifier auto mode |
| `BridgeDialog.tsx` | 34 KB | Claude Bridge connection flow |
| `BypassPermissionsModeDialog.tsx` | 9 KB | `--dangerously-skip-permissions` confirmation |
| `ChannelDowngradeDialog.tsx` | 8 KB | MCP channel downgrade warning |
| `ClaudeMdExternalIncludesDialog.tsx` | 14 KB | Approve `@include` directives in CLAUDE.md |
| `ClaudeInChromeOnboarding.tsx` | 12 KB | First-run for Chrome extension |
| `CostThresholdDialog.tsx` | 4.3 KB | Cost threshold exceeded warning |
| `DesktopHandoff.tsx` | 19 KB | Hand off session to Claude Desktop app |
| `DevChannelsDialog.tsx` | 9 KB | Dev channel enable confirmation |
| `ExportDialog.tsx` | 19 KB | Export conversation |
| `IdeAutoConnectDialog.tsx` | 13 KB | IDE auto-connect prompt |
| `IdeOnboardingDialog.tsx` | 16 KB | IDE first-run |
| `IdleReturnDialog.tsx` | 9.9 KB | Resume after idle |
| `InvalidConfigDialog.tsx` | 15 KB | Invalid config error recovery |
| `InvalidSettingsDialog.tsx` | 7 KB | Settings validation errors |
| `MCPServerApprovalDialog.tsx` | 11 KB | Approve new MCP server |
| `MCPServerDesktopImportDialog.tsx` | 21 KB | Import MCP servers from Desktop |
| `MCPServerMultiselectDialog.tsx` | 16 KB | Multi-select MCP server enable |
| `ManagedSettingsSecurityDialog/` | dir | Enterprise-managed settings warning |
| `QuickOpenDialog.tsx` | 28 KB | Cmd+Shift+P file picker (ctrl+shift+p) |
| `RemoteCallout.tsx` | 10 KB | Remote execution notice |
| `RemoteEnvironmentDialog.tsx` | 33 KB | Remote env config |
| `ShowInIDEPrompt.tsx` | 17 KB | "Open in IDE?" prompt |
| `TeleportResumeWrapper.tsx` | 15 KB | Resume Teleport (remote) session |
| `TeleportRepoMismatchDialog.tsx` | 13 KB | Repo mismatch during Teleport resume |
| `TeleportError.tsx` | 19 KB | Teleport failure display |
| `TeleportStash.tsx` | 15 KB | Teleport stashed changes |
| `TrustDialog/` | dir | Workspace trust (first-run for untrusted dirs) |
| `WorktreeExitDialog.tsx` | 35 KB | Worktree cleanup on exit (keep/discard branches) |
| `GlobalSearchDialog.tsx` | 44 KB | Cross-repo search |
| `HistorySearchDialog.tsx` | 19 KB | Ctrl+R bash-style history search |
| `diff/DiffDialog.tsx` | 43 KB | Multi-file diff review modal |

Design-system primitive: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/Dialog.tsx` (14 KB, 137 lines). Signature at `Dialog.tsx:11-29` — props: `title`, `subtitle`, `children`, `onCancel`, `color` (theme key), `hideInputGuide`, `hideBorder`, `inputGuide` (custom footer), `isCancelActive` (turn off built-in Esc/Ctrl-C when embedding a text input that handles its own cancel).

### 2.2 Input Components (18)

| Component | Size | Role |
|---|---|---|
| `TextInput.tsx` | 21 KB | Single-line text input with voice waveform cursor + highlights |
| `BaseTextInput.tsx` | 19 KB | Base text input (no decoration) |
| `VimTextInput.tsx` | 16 KB | Vim-mode text input (normal/insert/visual) |
| `SearchBox.tsx` | 9.4 KB | Search input with `/` entry |
| `CustomSelect/select.tsx` | 115 KB | Generic select list (options, keyboard nav, inline input) |
| `CustomSelect/select-input-option.tsx` | 58 KB | Option that itself is a text input (for free-form alongside presets) |
| `CustomSelect/select-option.tsx` | 5.8 KB | Presentation-only option row |
| `CustomSelect/SelectMulti.tsx` | 30 KB | Multi-select with checkboxes + optional submit button |
| `CustomSelect/use-select-input.ts` | 8.8 KB | Shared hook |
| `CustomSelect/use-select-state.ts` | 2.9 KB | Selection state reducer |
| `CustomSelect/use-multi-select-state.ts` | 11 KB | Multi-select state |
| `CustomSelect/use-select-navigation.ts` | 16 KB | Arrow key navigation, wrap behavior, top/bottom jumps |
| `design-system/FuzzyPicker.tsx` | 41 KB | Fuzzy search picker with preview pane (used by QuickOpen, History, Agent list) |
| `design-system/Tabs.tsx` | 41 KB | Tab bar (used in settings, permission rule list, MCP panel) |
| `LanguagePicker.tsx` | 8.7 KB | Select conversation language |
| `ModelPicker.tsx` | 54 KB | Model + effort level selector (Opus/Sonnet/Haiku × low/medium/high/xhigh/max) |
| `OutputStylePicker.tsx` | 13 KB | Pick an output style (see Section 6) |
| `ThemePicker.tsx` | 36 KB | Pick a theme |

Design notes:
- `SelectMulti` at `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/CustomSelect/SelectMulti.tsx:11-56` supports two modes: with `submitButtonText` Enter fires submit, Space toggles; without it Enter submits directly.
- `OptionWithDescription` at `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/CustomSelect/select.tsx:28-69` has a built-in `type: 'input'` variant that makes an option itself become a text field when focused — the pattern behind permission prompts where an option is "Yes, with custom feedback: ___".

### 2.3 Progress & Status (9 components + Spinner subdir)

| Component | Size | Role |
|---|---|---|
| `Spinner.tsx` | 88 KB | Main spinner with verb rotation, shimmer, teammate tree |
| `Spinner/SpinnerAnimationRow.tsx` | 43 KB | Animated row of glyphs |
| `Spinner/SpinnerGlyph.tsx` | 10 KB | Single animating glyph |
| `Spinner/GlimmerMessage.tsx` | 27 KB | Shimmering status message |
| `Spinner/FlashingChar.tsx`, `ShimmerChar.tsx`, `TeammateSpinnerLine.tsx`, `TeammateSpinnerTree.tsx` | — | Variants for team mode |
| `ToolUseLoader.tsx` | 4.8 KB | Small spinner next to tool-in-progress messages |
| `AgentProgressLine.tsx` | 14 KB | Agent-level status (what the agent is doing) |
| `BashModeProgress.tsx` | 5.7 KB | Bash command progress |
| `HookProgressMessage.tsx` | 11 KB | Hook execution progress |
| `tasks/ShellProgress.tsx` | 7 KB | Shell task progress |
| `tasks/RemoteSessionProgress.tsx` | 28 KB | Remote session progress |
| `TeleportProgress.tsx` | 16 KB | Teleport session sync |
| `design-system/ProgressBar.tsx` | 7.2 KB | Generic horizontal progress bar (Unicode eighth-block characters for smooth fill; see `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ProgressBar.tsx:26` — `BLOCKS = [' ', '▏', '▎', '▍', '▌', '▋', '▊', '▉', '█']`) |
| `design-system/Ratchet.tsx` | 7.2 KB | Stepped progress (phase 1/2/3 style) |
| `design-system/LoadingState.tsx` | 6.4 KB | Generic loading wrapper |
| `design-system/StatusIcon.tsx` | 7.6 KB | Status icon + color (success/error/warning/info/pending/loading), see `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/StatusIcon.tsx:5-52` |
| `MemoryUsageIndicator.tsx` | 4.4 KB | RSS/heap usage pill |
| `StatusLine.tsx` | 49 KB | Full bottom status line (model, mode, cwd, tokens, rate limits) |
| `StatusNotices.tsx` | 5.7 KB | Queue of ephemeral status notices |
| `EffortCallout.tsx` | 25 KB | Callout when effort level limits reasoning |
| `EffortIndicator.ts` | 1.1 KB | Inline effort badge |

Also: `TaskListV2.tsx` (50 KB) — rich todo list with status per item, used for the live task board (`ctrl+t` to toggle).

### 2.4 Tool-Use Rendering (messages/ subdirectory, 36 files)

All ways the assistant or the user sees tool invocations:

- `messages/AssistantTextMessage.tsx` (30 KB) — Assistant plain text
- `messages/AssistantToolUseMessage.tsx` (45 KB) — "Claude is running `<tool>` with <args>"
- `messages/AssistantThinkingMessage.tsx` (8 KB) — Extended thinking block (collapsed by default)
- `messages/AssistantRedactedThinkingMessage.tsx` (2.6 KB) — Redacted thinking
- `messages/HighlightedThinkingText.tsx` (14.9 KB) — Highlighted thinking text
- `messages/UserTextMessage.tsx` (29 KB) — User input
- `messages/UserBashInputMessage.tsx` (4.5 KB) — `$ cmd` style bash input
- `messages/UserBashOutputMessage.tsx` (4.2 KB) — Bash output
- `messages/UserLocalCommandOutputMessage.tsx` (14.6 KB) — Local slash command output
- `messages/UserImageMessage.tsx` (5.9 KB) — Image attachment
- `messages/UserCommandMessage.tsx` (9.2 KB) — Slash command invocation
- `messages/UserPromptMessage.tsx` (15.2 KB) — User prompt display
- `messages/UserMemoryInputMessage.tsx` (6.5 KB) — `/memory` save
- `messages/UserPlanMessage.tsx` (3.7 KB) — Plan mode output
- `messages/UserResourceUpdateMessage.tsx` (12.4 KB) — MCP resource update
- `messages/UserAgentNotificationMessage.tsx` (6.2 KB) — Sub-agent notification
- `messages/UserTeammateMessage.tsx` (24 KB) — Team mode message to teammate
- `messages/UserChannelMessage.tsx` (11 KB) — Channel messages
- `messages/AdvisorMessage.tsx` (14 KB) — Advisor hint
- `messages/SystemTextMessage.tsx` (79 KB!) — System/status messages (the largest message type)
- `messages/SystemAPIErrorMessage.tsx` (12 KB) — API error display
- `messages/CompactBoundaryMessage.tsx` (2.3 KB) — "Compaction boundary" marker
- `messages/PlanApprovalMessage.tsx` (26 KB) — Plan mode approval UI
- `messages/RateLimitMessage.tsx` (17 KB) — Rate limit message
- `messages/ShutdownMessage.tsx` (14 KB) — Shutdown sequence message
- `messages/TaskAssignmentMessage.tsx` (8 KB) — Team mode task assignment
- `messages/teamMemCollapsed.tsx` (13 KB) — Collapsed teammate memory block
- `messages/GroupedToolUseContent.tsx` (8 KB) — Groups multiple consecutive tool uses
- `messages/CollapsedReadSearchContent.tsx` (78 KB) — Collapses read-heavy operations
- `messages/HookProgressMessage.tsx` (10.6 KB) — Hook progress
- `messages/UserToolResultMessage/` — subdirectory for per-tool result renderers (10 items)

Also at the top level:
- `Message.tsx` (79 KB), `MessageRow.tsx` (48 KB), `Messages.tsx` (147 KB!), `MessageModel.tsx` (4 KB), `MessageResponse.tsx` (7 KB), `MessageSelector.tsx` (116 KB — rewind-to-earlier-message selector), `MessageTimestamp.tsx` (5.4 KB), `messageActions.tsx` (55 KB — per-message action menu)

### 2.5 Markdown Renderer (2 components + utils)

- `Markdown.tsx` (28 KB) — Main entry; tables → `MarkdownTable`, others → ANSI via `utils/markdown.formatToken`
- `MarkdownTable.tsx` (47 KB) — Flex-based table renderer

Supporting utilities (not in `components/` but referenced):
- `utils/markdown.ts` — `formatToken(token, theme, width)` — pre-renders each marked Token to ANSI-colored string

### 2.6 Diff Viewer (3 files)

- `diff/DiffDetailView.tsx` (23 KB) — Single-file diff with gutter
- `diff/DiffDialog.tsx` (43 KB) — Multi-file diff modal with file list
- `diff/DiffFileList.tsx` (25 KB) — File list for diff dialog

Standalone:
- `StructuredDiff.tsx` (25 KB) + `StructuredDiff/Fallback.tsx` + `StructuredDiff/colorDiff.ts` — Hunk-level diff rendering (used by file edit previews)
- `StructuredDiffList.tsx` (4.3 KB) — List of structured diffs
- `FileEditToolDiff.tsx` (22 KB) — File edit tool specific diff display
- `FileEditToolUpdatedMessage.tsx` (12 KB) + `FileEditToolUseRejectedMessage.tsx` (15 KB) — File edit completion/rejection messages

### 2.7 Error & Notification (10)

- `FallbackToolUseErrorMessage.tsx` (12.6 KB) — Generic tool error fallback
- `FallbackToolUseRejectedMessage.tsx` (1.7 KB) — Generic tool rejected
- `NotebookEditToolUseRejectedMessage.tsx` (8.4 KB)
- `InterruptedByUser.tsx` (2 KB) — "User hit Ctrl+C" banner
- `KeybindingWarnings.tsx` (9.5 KB) — Keybindings validation warnings
- `DiagnosticsDisplay.tsx` (13 KB) — LSP diagnostics listing
- `ValidationErrorsList.tsx` (19.6 KB) — Settings/config validation errors
- `StatusNotices.tsx` (5.7 KB) — Transient toast-like notifications
- `SandboxViolationExpandedView.tsx` (11 KB) — Sandbox violation details
- `SystemAPIErrorMessage.tsx` (12 KB) — API error display

Also `SentryErrorBoundary.ts` (0.5 KB) — React error boundary wrapping Sentry reporter.

### 2.8 Permission UI (32 items — see Section 7 for full flow)

All at `components/permissions/`. These are the bread-and-butter modals for tool-use approval.

### 2.9 Specialized (24 notable files)

SPEC/code/git specialized viewers:
- `Markdown.tsx`, `MarkdownTable.tsx`, `HighlightedCode.tsx`, `HighlightedCode/Fallback.tsx` — content
- `StructuredDiff.tsx`, `StructuredDiffList.tsx` — diffs
- `PrBadge.tsx` (7.7 KB) — PR status badge inline
- `FilePathLink.tsx` (3.2 KB) — Clickable file path (OSC-8 hyperlink)
- `Feedback.tsx` (88 KB!) — Feedback collection flow (the largest single-file component)
- `FeedbackSurvey/` (11 files) — Survey UI primitives
- `SkillImprovementSurvey.tsx` (15 KB) — Skill improvement survey
- `Stats.tsx` (153 KB!) — Usage stats display (tokens, cost, duration, session breakdown)
- `ContextVisualization.tsx` (76 KB) — Context window bar chart
- `TokenWarning.tsx` (21.5 KB) — Token budget warning
- `TagTabs.tsx` (21 KB) — Tag-based tab navigator
- `VirtualMessageList.tsx` (149 KB!) — Virtual-scroll list (the largest file besides `Messages.tsx`)
- `ScrollKeybindingHandler.tsx` (149 KB!) — Scroll + keybinding event handler
- `FullscreenLayout.tsx` (85 KB) — Fullscreen mode layout
- `CompactSummary.tsx` (14 KB) — Post-compaction summary
- `LogSelector.tsx` (200 KB!) — Log browsing/selection (very large — likely SQL-backed)
- `ResumeTask.tsx` (38.5 KB) — Resume interrupted task
- `Onboarding.tsx` (31.5 KB) — First-run wizard
- `sandbox/` (7 items) — Sandbox-specific components
- `wizard/` (7 items) — Wizard flow primitives
- `teams/` (4 items) — Team-mode components
- `tasks/BackgroundTask.tsx` (31 KB), `tasks/BackgroundTaskStatus.tsx` (43 KB), `tasks/BackgroundTasksDialog.tsx` (116 KB!), `tasks/AsyncAgentDetailDialog.tsx` (30 KB), `tasks/RemoteSessionDetailDialog.tsx` (96 KB!), `tasks/InProcessTeammateDetailDialog.tsx` (31 KB), `tasks/DreamDetailDialog.tsx` (26 KB), `tasks/ShellDetailDialog.tsx` (39 KB), `tasks/renderToolActivity.tsx` (4.4 KB), `tasks/taskStatusUtils.tsx` (13.5 KB)

### 2.10 Design System (design-system/ subdir, 17 components)

The canonical primitives layer:

| File | Size | Role |
|---|---|---|
| `Byline.tsx` | 6.5 KB | Inline label + content pair (bylines beneath dialog titles) |
| `color.ts` | 0.8 KB | Color type + resolver |
| `Dialog.tsx` | 14 KB | Modal wrapper w/ title/subtitle/cancel/inputGuide |
| `Divider.tsx` | 11 KB | Horizontal rule (solid/dashed/dotted with label) |
| `FuzzyPicker.tsx` | 41 KB | Fuzzy-search picker with preview pane |
| `KeyboardShortcutHint.tsx` | 6.8 KB | "Enter" style key pill |
| `ListItem.tsx` | 19.5 KB | Row with pointer (❯), check (✓), scroll hint (↑/↓), focused/selected/disabled styling |
| `LoadingState.tsx` | 6.5 KB | Generic loading wrapper |
| `Pane.tsx` | 6.9 KB | Bordered pane with title/footer |
| `ProgressBar.tsx` | 7.2 KB | Eighth-block-character progress bar |
| `Ratchet.tsx` | 7.2 KB | Stepped/phased progress (1/N indicator) |
| `StatusIcon.tsx` | 7.6 KB | Status icon + theme color |
| `Tabs.tsx` | 41 KB | Tabs + Tab + useTabsWidth + useTabHeaderFocus |
| `ThemedBox.tsx` | 18 KB | Box with theme-key color resolution |
| `ThemedText.tsx` | 13.9 KB | Text with theme-key color resolution + dimColor |
| `ThemeProvider.tsx` | 18.9 KB | Theme provider + `useTheme`/`useThemeSetting`/`usePreviewTheme` |

The `resolveColor` pattern appears twice — in `ThemedBox.tsx:42-50` and `ThemedText.tsx:66-74`. A raw color string (`rgb(...)`, `#...`, `ansi256(...)`, `ansi:...`) is passed through; anything else is treated as a theme key.

### 2.11 Large Category Summary

| Category | Count |
|---|---|
| Dialogs | 33 |
| Inputs | 18 |
| Progress/Status | 16 |
| Message renderers | 36 |
| Markdown/code/diff | 9 |
| Error/notification | 10 |
| Permission | 32 |
| Specialized (stats/context/log/session) | 24 |
| Design-system primitives | 17 |
| Agents/skills/teams/tasks | 23 |
| **Total (unique tracks)** | **218** (some files span categories) |

The 146 at the top level plus nested subdirectories account for roughly 250+ individual `.tsx` files totaling ~5 MB.

---

## 3. Interactive Flow Helpers

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx` (57 KB, 366 lines)

The whole purpose of this file is stated at its top (imports only — the body starts at `:32`). It is a dependency-injection-heavy helpers module for the main entrypoint. Key exports:

### 3.1 `completeOnboarding()` — `:32-38`

Saves `hasCompletedOnboarding: true` and `lastOnboardingVersion: MACRO.VERSION` to global config. Invoked when the onboarding wizard finishes.

### 3.2 `showDialog<T>(root, renderer)` — `:39-44`

Primitive for rendering a Promise-returning dialog:

```ts
export function showDialog<T = void>(root: Root, renderer: (done: (result: T) => void) => React.ReactNode): Promise<T>
```

Wraps the `renderer` in a `Promise`; the renderer receives a `done(result: T)` callback that resolves the promise. Every interactive flow in the app is fundamentally a chain of `showDialog` calls.

### 3.3 `exitWithError(root, message, beforeExit?)` — `:52-57`

Thin helper over `exitWithMessage` that forces `color: 'error'`.

### 3.4 `exitWithMessage(root, message, options?)` — `:65-80`

Critical pattern — because Ink's `patchConsole` eats `console.error`, fatal errors must be rendered through the React tree. This dynamically imports `Text`, renders the colorized message, unmounts the root, awaits `options?.beforeExit?.()`, then `process.exit(exitCode ?? 1)`.

### 3.5 `showSetupDialog<T>(root, renderer, options?)` — `:86-92`

The workhorse for all onboarding/setup dialogs. Wraps the renderer in `<AppStateProvider>` and `<KeybindingSetup>` so every dialog has state + keybinding context. Every dialog in `showSetupScreens()` funnels through this helper.

### 3.6 `renderAndRun(root, element)` — `:98-103`

Main UI entry epilogue:
1. `root.render(element)`
2. `startDeferredPrefetches()` (e.g. fetching system context, GrowthBook flags)
3. `await root.waitUntilExit()` — blocks until the app exits
4. `await gracefulShutdown(0)` — cleanup registered handlers

Used by `replLauncher.tsx` and `launchResumeChooser` in `dialogLaunchers.tsx`.

### 3.7 `showSetupScreens(...)` — `:104-298`

The star of the file. Orchestrates the entire first-run onboarding sequence:

1. `:111-123` — Onboarding wizard (if `!config.theme || !config.hasCompletedOnboarding`)
2. `:131-140` — Trust dialog (if `!checkHasTrustDialogAccepted()` and not in `CLAUBBIT` env)
3. `:142-170` — Post-trust setup: GrowthBook reset, fetch system context, handle mcp.json server approvals, external CLAUDE.md includes approval
4. `:173-175` — Update github repo path mapping
5. `:184-190` — Apply env vars + initialize telemetry (deferred to next tick)
6. `:191-202` — Grove eligibility dialog (policy update modal)
7. `:206-217` — Custom API key approval (if `ANTHROPIC_API_KEY` present and `!isRunningOnHomespace()`)
8. `:218-223` — Bypass permissions mode warning (if `--dangerously-skip-permissions`)
9. `:224-235` — Auto mode opt-in (if `TRANSCRIPT_CLASSIFIER` feature)
10. `:237-288` — Dev channels approval (`KAIROS` feature — `--dangerously-load-development-channels`)
11. `:291-296` — Claude-in-Chrome onboarding

Each phase's presence is governed by feature flags (`feature('KAIROS')`, `feature('LODESTONE')`) and config state. Completion of any phase may early-exit (see Grove `decision === 'escape'` at `:197-200`).

### 3.8 `getRenderContext(exitOnCtrlC)` — `:299-365`

Builds the `RenderOptions` passed to `createRoot()`. Wires up:
- FPS tracker (`fpsTracker.record(event.durationMs)` per frame)
- Stats store (observes `frame_duration_ms`)
- Bench mode — if `CLAUDE_CODE_FRAME_TIMING_LOG` env var is set, appends per-frame JSONL with timings + RSS + cpu for offline analysis. `:330-342`
- Flicker reporting — skipped when `isSynchronizedOutputSupported()` (DEC 2026); otherwise logs `tengu_flicker` analytics events when flicker reason isn't `resize`. `:343-362`

### 3.9 Typing Indicators / Animations

Not present in this file — they live in `components/Spinner/*`:
- `FlashingChar.tsx` — pulsing single character
- `ShimmerChar.tsx` — moving shimmer highlight
- `GlimmerMessage.tsx` — shimmer over a whole message
- `SpinnerAnimationRow.tsx` — animation glyphs + verbs
- `TeammateSpinnerLine.tsx`, `TeammateSpinnerTree.tsx` — team mode spinners showing sub-agent status
- Hooks: `useShimmerAnimation` (in Spinner/), `useStalledAnimation`, `useAnimationFrame` (ink/hooks/)

### 3.10 Permission Prompt Flow

Not directly in `interactiveHelpers.tsx` — that's a separate pipeline starting in `components/permissions/PermissionRequest.tsx` (see Section 7).

---

## 4. Dialog & REPL Launchers

### 4.1 `dialogLaunchers.tsx` (23 KB, 132 lines)

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx`

Per the file header comment at `:1-8`:

> "Thin launchers for one-off dialog JSX sites in main.tsx. Each launcher dynamically imports its component and wires the `done` callback identically to the original inline call site. Zero behavior change. Part of the main.tsx React/JSX extraction effort."

Every launcher follows the same pattern:
```ts
export async function launchXxx(root: Root, props): Promise<T> {
  const { Xxx } = await import('./components/Xxx.js')
  return showSetupDialog<T>(root, done => <Xxx {...props} onComplete={done} onCancel={() => done(null)} />)
}
```

Launchers defined:

1. **`launchSnapshotUpdateDialog`** — `:29-38`. Agent memory snapshot update prompt. Returns `'merge' | 'keep' | 'replace'`.
2. **`launchInvalidSettingsDialog`** — `:44-52`. Settings validation errors with continue/exit.
3. **`launchAssistantSessionChooser`** — `:58-65`. Pick a Claude Assistant bridge session. Returns `string | null`.
4. **`launchAssistantInstallWizard`** — `:73-85`. Install Claude Assistant daemon. Rejects promise on install failure, resolves with installed dir path or null on cancel. Uses `Promise.race` to distinguish error from cancel.
5. **`launchTeleportResumeWrapper`** — `:91-96`. Teleport session picker (remote execution). Returns `TeleportRemoteResponse | null`.
6. **`launchTeleportRepoMismatchDialog`** — `:102-110`. Pick local checkout of target repo.
7. **`launchResumeChooser`** — `:117-131`. **Different pattern** — uses `renderAndRun` (not `showSetupDialog`) and mounts `<App><KeybindingSetup><ResumeConversation>`. This is because the resume chooser IS the main UI for that session (not a transient setup dialog). Parallelizes `Promise.all([getWorktreePathsPromise, import('./screens/ResumeConversation'), import('./components/App')])` for cold start time.

### 4.2 `replLauncher.tsx` (3.5 KB)

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/replLauncher.tsx:12-22`

```tsx
export async function launchRepl(
  root: Root,
  appProps: AppWrapperProps,
  replProps: REPLProps,
  renderAndRun: (root: Root, element: React.ReactNode) => Promise<void>,
): Promise<void> {
  const { App } = await import('./components/App.js')
  const { REPL } = await import('./screens/REPL.js')
  await renderAndRun(root, <App {...appProps}><REPL {...replProps} /></App>)
}
```

Injected `renderAndRun` pattern — lets the caller substitute a test double. The REPL is the main interactive shell (the thing users type into).

### 4.3 Dialog → REPL Flow Reconstructed

Derived from `main.tsx` call sites (referenced from `dialogLaunchers.tsx` comments):

1. Parse CLI args (outside all of this)
2. Create Ink root via `createRoot(getRenderContext(exitOnCtrlC).renderOptions)`
3. `showSetupScreens(root, permissionMode, ...)` — runs the onboarding chain (Section 3.7)
4. Based on CLI flags:
   - `--resume` → `launchResumeChooser()` → ResumeConversation screen → (picks a session) → REPL
   - `--teleport` → `launchTeleportResumeWrapper()` → Teleport session
   - `assistant` subcommand → `launchAssistantSessionChooser()` / `launchAssistantInstallWizard()`
   - Normal interactive → `launchRepl()`
5. `await root.waitUntilExit()` (inside `renderAndRun`)
6. `gracefulShutdown(0)`

### 4.4 Approval Request Handlers

Not launchers — approval dialogs are not "one-off" like these; they are triggered mid-REPL by the agent when a tool requests permission. Flow is in Section 7.

### 4.5 Plan Mode UI

- `tools/EnterPlanModeTool/` + `tools/ExitPlanModeTool/` — the tools themselves
- `components/permissions/EnterPlanModePermissionRequest/` + `ExitPlanModePermissionRequest/` — the approval dialogs
- `components/messages/PlanApprovalMessage.tsx` (26 KB) — renders the plan approval inline in the transcript
- `components/messages/UserPlanMessage.tsx` (3.7 KB) — renders "user accepted plan" marker
- `utils/planModeV2.ts` — plan mode state machine (`isPlanModeInterviewPhaseEnabled` referenced from `AskUserQuestionPermissionRequest.tsx:20`)

Plan approval is UI-surfaced as a full-screen dialog with the proposed plan + edit/accept/reject options.

### 4.6 Worktree Selection UI

- `components/WorktreeExitDialog.tsx` (35 KB) — prompts on exit: keep worktree (uncommitted changes), delete, or force-delete
- Worktree discovery is via `getWorktreePaths` (referenced at `dialogLaunchers.tsx:122`), but the picker itself is part of `ResumeConversation`. The picker merges git worktree list + `.claude/worktrees/` ephemeral dirs.

---

## 5. Keybindings (Default + Customization Schema)

Keybindings are defined in `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/` (15 files, ~140 KB).

### 5.1 Default Binding Table

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts` (340 lines)

The full table is organized by **context** (a set of UI modes — Global, Chat, Confirmation, etc.). Below is the complete default table enumerated from the source.

**Context: Global** (`:33-62`)
| Key | Action | Notes |
|---|---|---|
| `ctrl+c` | `app:interrupt` | Hardcoded — special double-press |
| `ctrl+d` | `app:exit` | Hardcoded — special double-press |
| `ctrl+l` | `app:redraw` | |
| `ctrl+t` | `app:toggleTodos` | |
| `ctrl+o` | `app:toggleTranscript` | |
| `ctrl+shift+b` | `app:toggleBrief` | Feature-gated `KAIROS` / `KAIROS_BRIEF` |
| `ctrl+shift+o` | `app:toggleTeammatePreview` | |
| `ctrl+r` | `history:search` | |
| `ctrl+shift+f`, `cmd+shift+f` | `app:globalSearch` | Feature-gated `QUICK_SEARCH` |
| `ctrl+shift+p`, `cmd+shift+p` | `app:quickOpen` | Feature-gated `QUICK_SEARCH` |
| `meta+j` | `app:toggleTerminal` | Feature-gated `TERMINAL_PANEL` |

**Context: Chat** (`:63-98`)
| Key | Action | Notes |
|---|---|---|
| `escape` | `chat:cancel` | |
| `ctrl+x ctrl+k` | `chat:killAgents` | **Chord binding** — `ctrl+x` is prefix to avoid shadowing readline |
| `shift+tab` (or `meta+m` on Windows non-VT) | `chat:cycleMode` | Platform-adaptive — see `:29-30` |
| `meta+p` | `chat:modelPicker` | |
| `meta+o` | `chat:fastMode` | |
| `meta+t` | `chat:thinkingToggle` | |
| `enter` | `chat:submit` | |
| `up` | `history:previous` | |
| `down` | `history:next` | |
| `ctrl+_` | `chat:undo` | Legacy terminals (raw `\x1f`) |
| `ctrl+shift+-` | `chat:undo` | Kitty keyboard protocol |
| `ctrl+x ctrl+e` | `chat:externalEditor` | Readline-native chord |
| `ctrl+g` | `chat:externalEditor` | Also |
| `ctrl+s` | `chat:stash` | |
| `ctrl+v` (or `alt+v` on Windows) | `chat:imagePaste` | Platform-adaptive |
| `shift+up` | `chat:messageActions` | Feature-gated `MESSAGE_ACTIONS` |
| `space` | `voice:pushToTalk` | Feature-gated `VOICE_MODE` (hold-to-talk) |

**Context: Autocomplete** (`:99-107`)
| Key | Action |
|---|---|
| `tab` | `autocomplete:accept` |
| `escape` | `autocomplete:dismiss` |
| `up` | `autocomplete:previous` |
| `down` | `autocomplete:next` |

**Context: Settings** (`:108-128`)
| Key | Action |
|---|---|
| `escape` | `confirm:no` |
| `up`, `k`, `ctrl+p` | `select:previous` |
| `down`, `j`, `ctrl+n` | `select:next` |
| `space` | `select:accept` |
| `enter` | `settings:close` |
| `/` | `settings:search` |
| `r` | `settings:retry` |

**Context: Confirmation** (`:129-149`)
| Key | Action |
|---|---|
| `y`, `enter` | `confirm:yes` |
| `n`, `escape` | `confirm:no` |
| `up` | `confirm:previous` |
| `down` | `confirm:next` |
| `tab` | `confirm:nextField` |
| `space` | `confirm:toggle` |
| `shift+tab` | `confirm:cycleMode` |
| `ctrl+e` | `confirm:toggleExplanation` |
| `ctrl+d` | `permission:toggleDebug` |

**Context: Tabs** (`:150-158`)
| Key | Action |
|---|---|
| `tab`, `right` | `tabs:next` |
| `shift+tab`, `left` | `tabs:previous` |

**Context: Transcript** (`:159-170`)
| Key | Action |
|---|---|
| `ctrl+e` | `transcript:toggleShowAll` |
| `ctrl+c`, `escape`, `q` | `transcript:exit` |

**Context: HistorySearch** (`:171-179`)
| Key | Action |
|---|---|
| `ctrl+r` | `historySearch:next` |
| `escape`, `tab` | `historySearch:accept` |
| `ctrl+c` | `historySearch:cancel` |
| `enter` | `historySearch:execute` |

**Context: Task** (`:180-187`)
| Key | Action |
|---|---|
| `ctrl+b` | `task:background` |

**Context: ThemePicker** (`:188-193`)
| Key | Action |
|---|---|
| `ctrl+t` | `theme:toggleSyntaxHighlighting` |

**Context: Scroll** (`:195-213`)
| Key | Action |
|---|---|
| `pageup` | `scroll:pageUp` |
| `pagedown` | `scroll:pageDown` |
| `wheelup` | `scroll:lineUp` |
| `wheeldown` | `scroll:lineDown` |
| `ctrl+home` | `scroll:top` |
| `ctrl+end` | `scroll:bottom` |
| `ctrl+shift+c` | `selection:copy` |
| `cmd+c` | `selection:copy` (kitty protocol only) |

**Context: Help** (`:214-219`)
| Key | Action |
|---|---|
| `escape` | `help:dismiss` |

**Context: Attachments** (`:221-231`)
Arrow/backspace/delete navigation; `escape`/`down` exits.

**Context: Footer** (`:233-244`)
Up/down/`ctrl+p`/`ctrl+n` for footer indicator focus, `enter` activates.

**Context: MessageSelector** (`:247-266`)
Arrow + vi keys (`j`/`k`) + ctrl+/shift+/meta+ arrow for top/bottom jumps.

**Context: MessageActions** (`:268-294`) (feature-gated)
Up/down/`j`/`k` nav, `shift+up/down` → prev/next user, `meta+`/`super+` up/down → top/bottom, plus `c`, `p`, `enter` for actions.

**Context: DiffDialog** (`:296-308`)
Arrow keys navigate; `enter` views details.

**Context: ModelPicker** (`:309-316`)
`left`/`right` to decrease/increase effort level (ant-only).

**Context: Select** (`:317-329`)
Standard select nav — `up`/`k`/`ctrl+p`, `down`/`j`/`ctrl+n`, `enter` accept, `escape` cancel.

**Context: Plugin** (`:331-339`)
`space` toggle, `i` install.

### 5.2 Platform-Adaptive Bindings

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:12-30`

```ts
const IMAGE_PASTE_KEY = getPlatform() === 'windows' ? 'alt+v' : 'ctrl+v'

const SUPPORTS_TERMINAL_VT_MODE =
  getPlatform() !== 'windows' ||
  (isRunningWithBun()
    ? satisfies(process.versions.bun, '>=1.2.23')
    : satisfies(process.versions.node, '>=22.17.0 <23.0.0 || >=24.2.0'))

const MODE_CYCLE_KEY = SUPPORTS_TERMINAL_VT_MODE ? 'shift+tab' : 'meta+m'
```

Two platform quirks surface here:
1. Windows uses `ctrl+v` for system paste, so image paste is `alt+v`.
2. Windows Terminal without VT mode drops modifier-only keys (`shift+tab`), so `meta+m` is used as mode-cycle fallback. Node enabled VT in 22.17.0 / 24.2.0; Bun in 1.2.23.

### 5.3 Chord Support

Chords are space-separated keystroke sequences like `ctrl+x ctrl+k`. Defined and parsed at `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/parser.ts:80-84`:

```ts
export function parseChord(input: string): Chord {
  if (input === ' ') return [parseKeystroke('space')]
  return input.trim().split(/\s+/).map(parseKeystroke)
}
```

Resolution handled in `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/resolver.ts:166-244` `resolveKeyWithChordState`:

1. Escape during pending chord → `{ type: 'chord_cancelled' }`
2. Build current keystroke; append to pending
3. Check if any binding's chord *prefix-matches* the testChord → if so, return `{ type: 'chord_started', pending: testChord }`
4. Otherwise check for exact match → `{ type: 'match', action }`
5. No match → `'chord_cancelled'` (if pending) or `'none'`

Edge case handled at `:196-215`:
> "Group by chord string so a later null-override shadows the default it unbinds — otherwise null-unbinding `ctrl+x ctrl+k` still makes `ctrl+x` enter chord-wait and the single-key binding on the prefix never fires."

### 5.4 Customization Schema

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/schema.ts` (236 lines)

Zod schema shape at `:177-229`:

```ts
export const KeybindingBlockSchema = lazySchema(() =>
  z.object({
    context: z.enum(KEYBINDING_CONTEXTS),
    bindings: z.record(
      z.string(),  // keystroke pattern
      z.union([
        z.enum(KEYBINDING_ACTIONS),  // built-in action
        z.string().regex(/^command:[a-zA-Z0-9:\-_]+$/),  // slash command
        z.null()  // unbind
      ])
    )
  })
)

export const KeybindingsSchema = lazySchema(() =>
  z.object({
    $schema: z.string().optional(),  // JSON Schema URL for editor validation
    $docs: z.string().optional(),
    bindings: z.array(KeybindingBlockSchema())
  })
)
```

Example user `~/.claude/keybindings.json`:
```json
{
  "$schema": "https://example.com/keybindings.schema.json",
  "bindings": [
    {
      "context": "Chat",
      "bindings": {
        "ctrl+y": "chat:submit",
        "ctrl+s": null
      }
    }
  ]
}
```

Contexts enum (`:12-33`): 18 total — `Global`, `Chat`, `Autocomplete`, `Confirmation`, `Help`, `Transcript`, `HistorySearch`, `Task`, `ThemePicker`, `Settings`, `Tabs`, `Attachments`, `Footer`, `MessageSelector`, `DiffDialog`, `ModelPicker`, `Select`, `Plugin`.

Actions enum (`:64-172`): ~85 actions grouped by prefix (`app:`, `history:`, `chat:`, `autocomplete:`, `confirm:`, `tabs:`, `transcript:`, `historySearch:`, `task:`, `theme:`, `help:`, `attachments:`, `footer:`, `messageSelector:`, `diff:`, `modelPicker:`, `select:`, `plugin:`, `permission:`, `settings:`, `voice:`).

Special action types:
- **Built-in action** (enum): invokes internal handler
- **Command binding**: `"command:help"` or `"command:compact"` — executes slash command as if typed (regex `/^command:[a-zA-Z0-9:\-_]+$/`)
- **Null**: unbinds a default

### 5.5 Reserved Shortcuts

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/reservedShortcuts.ts` (127 lines)

Three tiers:

1. **Non-rebindable (`NON_REBINDABLE` at `:16-33`)** — hardcoded, cannot be user-rebound:
   - `ctrl+c` — interrupt/exit
   - `ctrl+d` — exit
   - `ctrl+m` — identical to Enter in terminals (both send CR)
2. **Terminal-reserved (`TERMINAL_RESERVED` at `:43-54`)** — OS intercepts:
   - `ctrl+z` — SIGTSTP
   - `ctrl+\` — SIGQUIT
   - NOT listed: `ctrl+s`/`ctrl+q` (XOFF/XON) because modern terminals disable flow control and CC uses `ctrl+s` for stash.
3. **macOS-reserved (`MACOS_RESERVED` at `:59-67`)** — system-level shortcuts:
   - `cmd+c`, `cmd+v`, `cmd+x`, `cmd+q`, `cmd+w`, `cmd+tab`, `cmd+space`

### 5.6 Loading User Bindings

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/loadUserBindings.ts` (473 lines)

Key behaviors:
- `:41-46` `isKeybindingCustomizationEnabled()` — gated behind `tengu_keybinding_customization_release` GrowthBook flag. **External users currently always get defaults only.** From `:9-11`:
  > "User keybinding customization is currently only available for Anthropic employees (USER_TYPE === 'ant'). External users always use the default bindings."
- `:115-117` Config path: `~/.claude/keybindings.json` (via `getClaudeConfigHomeDir()`)
- `:133-237` `loadKeybindings()` async — reads file, validates JSON → KeybindingBlockArray → parses → merges with defaults (user bindings come AFTER defaults so they override). Returns warnings.
- `:243-345` `loadKeybindingsSync()` + sync-with-warnings version — used in React `useState` initializers.
- `:353-404` `initializeKeybindingWatcher()` — Chokidar watches `~/.claude/keybindings.json` with `stabilityThreshold: 500ms`, `pollInterval: 200ms`. On change/add/unlink, reparses and notifies subscribers via `keybindingsChanged` signal.
- `:424-437` `handleChange` and `:439-448` `handleDelete` — hot-reload handlers.

### 5.7 Validation

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/validate.ts` (14.5 KB) — validates user bindings against:
- Reserved shortcuts (error severity)
- Action exists in `KEYBINDING_ACTIONS` or matches command regex
- Duplicate keys in same context
- `checkDuplicateKeysInJson` — raw JSON duplicate key detection (since `JSON.parse` silently drops earlier values)

### 5.8 `useKeybinding` / `useKeybindings` Hooks

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/useKeybinding.ts` (197 lines)

Single-action version at `:33-97`:
```tsx
useKeybinding('app:toggleTodos', () => setShowTodos(x => !x), { context: 'Global' })
```
The handler can return `false` to mean "not consumed" — event then propagates to later handlers. Or return `Promise<void>` for fire-and-forget async. Otherwise `void`.

Multi-action version at `:113-196`:
```tsx
useKeybindings({
  'chat:submit': () => handleSubmit(),
  'chat:cancel': () => handleCancel(),
}, { context: 'Chat' })
```

Both call the underlying `useInput` from Ink once per hook (multi version reduces hook count). Both register handlers with `KeybindingContext` so the central resolver knows which handlers are active.

### 5.9 Display & Format

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/shortcutFormat.ts` (2.5 KB), `useShortcutDisplay.ts` (2.5 KB), `parser.ts:105-176` — convert `ParsedKeystroke` to platform-appropriate display:
- `alt`/`meta` → `opt` on macOS, `alt` elsewhere
- `super` → `cmd` on macOS, `super` elsewhere
- Arrow keys → `↑` `↓` `←` `→`
- `escape` → `Esc`, `space` → `Space`, `enter` → `Enter`

---

## 6. Output Styles (Schema + Activation)

### 6.1 Loader

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts` (99 lines)

The entry point `getOutputStyleDirStyles(cwd)` at `:26-92` is `memoize`d (lodash). It:

1. Calls `loadMarkdownFilesForSubdir('output-styles', cwd)` — utility scans throughout project hierarchy + `~/.claude/output-styles/`
2. For each `.md` file:
   - `:38` style name = basename without `.md`
   - `:41-42` `frontmatter.name` overrides filename
   - `:43-50` `frontmatter.description` via `coerceDescriptionToString`; falls back to `extractDescriptionFromMarkdown(content, ...)` (parses first paragraph)
   - `:53-62` `frontmatter.keep-coding-instructions` — accepts boolean or `'true'`/`'false'` strings; otherwise `undefined`
   - `:64-70` `frontmatter.force-for-plugin` — warns if set on non-plugin style ("this option only applies to plugin output styles")
3. Returns `OutputStyleConfig[]` with `{ name, description, prompt, source, keepCodingInstructions }`

Precedence: Project `.claude/output-styles/*.md` overrides user `~/.claude/output-styles/*.md`.

### 6.2 Schema (inferred from loader + picker)

```yaml
---
name: "Concise Mode"                    # optional, defaults to filename
description: "Responds tersely"          # optional, extracted from first paragraph
keep-coding-instructions: false          # optional bool/string
# force-for-plugin: true                 # plugins only — warned otherwise
---

<markdown body becomes the style prompt>
```

The body text is the "output style prompt" that modifies Claude's base system prompt.

### 6.3 Built-In + User Styles Picker

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/OutputStylePicker.tsx` (13 KB)

At `:48-57`:

```ts
getAllOutputStyles(getCwd()).then(allStyles => {
  const options = mapConfigsToOptions(allStyles);
  setStyleOptions(options);
  setIsLoading(false);
}).catch(() => {
  const builtInOptions = mapConfigsToOptions(OUTPUT_STYLE_CONFIG);
  setStyleOptions(builtInOptions);
  setIsLoading(false);
})
```

`OUTPUT_STYLE_CONFIG` (from `constants/outputStyles.ts`, not in scope but referenced) = built-ins shipped with CC. On file I/O error, falls back to built-ins only.

### 6.4 Default Built-Ins (Inferred)

From `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/OutputStylePicker.tsx:11-12`:
> "Default" / "Claude completes coding tasks efficiently and provides concise responses"

And from imports throughout the codebase, CC ships at minimum: `Default`, plus additional ones via MCP plugins. The actual shipped file list isn't in our scope (it's in `constants/outputStyles.ts`).

### 6.5 Activation

- `components/StatusLine.tsx:44` — `const outputStyleName = settings?.outputStyle || DEFAULT_OUTPUT_STYLE_NAME;` — current style is a setting
- The selected style's `prompt` markdown is concatenated into the system prompt (the loader exposes the prompt; the integration point is in `context.ts` or system prompt construction, outside our scope)
- Plugin output styles from `loadPluginOutputStyles.ts` (cache clearable via `clearPluginOutputStyleCache()` at `loadOutputStylesDir.ts:11`)

### 6.6 How Output Styles Override Default Behavior

The `keepCodingInstructions` flag controls whether base coding-specific system prompt content is retained:
- `undefined` (default) → base behavior
- `true` → keep default coding instructions AND append the style prompt
- `false` → **replace** coding instructions with the style prompt

Force-for-plugin: plugin-provided styles can be marked as "force" meaning they activate without user selection (rare).

### 6.7 Examples in the Tree (moai-adk has its own format)

CC itself ships few output styles in the dumped source (most live in plugins). The format is distinct from moai-adk-go's format described in the user's context — moai-adk uses its own schema in `.moai/output-styles/` via `SKILL.md`. However, the base CC output-style format (frontmatter + markdown prompt) is directly compatible with the path hierarchy.

---

## 7. Permission UI Flow

Permission UI is the single most-exercised dialog surface in CC. 32 items under `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/`.

### 7.1 Entry Point — PermissionRequest

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionRequest.tsx` (34 KB)

The dispatcher `permissionComponentForTool(tool)` at `:47-81` maps tool classes to permission components:

| Tool | Permission Component |
|---|---|
| `FileEditTool` | `FileEditPermissionRequest` |
| `FileWriteTool` | `FileWritePermissionRequest` |
| `BashTool` | `BashPermissionRequest` |
| `PowerShellTool` | `PowerShellPermissionRequest` |
| `ReviewArtifactTool` | `ReviewArtifactPermissionRequest` (feature-gated) |
| `WebFetchTool` | `WebFetchPermissionRequest` |
| `NotebookEditTool` | `NotebookEditPermissionRequest` |
| `ExitPlanModeV2Tool` | `ExitPlanModePermissionRequest` |
| `EnterPlanModeTool` | `EnterPlanModePermissionRequest` |
| `SkillTool` | `SkillPermissionRequest` |
| `AskUserQuestionTool` | `AskUserQuestionPermissionRequest` |
| `WorkflowTool` | `WorkflowPermissionRequest` (feature-gated) |
| `MonitorTool` | `MonitorPermissionRequest` (feature-gated) |
| `GlobTool`, `GrepTool`, `FileReadTool` | `FilesystemPermissionRequest` |
| (default) | `FallbackPermissionRequest` |

### 7.2 PermissionPrompt Primitive

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionPrompt.tsx` (37 KB)

Signature at `:23-29`:

```ts
export type PermissionPromptProps<T extends string> = {
  options: PermissionPromptOption<T>[];
  onSelect: (value: T, feedback?: string) => void;
  onCancel?: () => void;
  question?: string | ReactNode;
  toolAnalyticsContext?: ToolAnalyticsContext;
}

export type PermissionPromptOption<T extends string> = {
  value: T;
  label: ReactNode;
  feedbackConfig?: { type: 'accept' | 'reject'; placeholder?: string };
  keybinding?: KeybindingAction;
}
```

Key behaviors at `:44-119`:
- Question defaults to "Do you want to proceed?"
- Each option can have `feedbackConfig` — when focused, Tab expands the option into a text input where user types feedback alongside accept/reject. This is the pattern for "Yes, but next time do X" feedback.
- Analytics events on feedback interactions
- Options transformed into Select-compatible format with inline input when feedback mode active

### 7.3 AskUserQuestion — Specialized Permission Flow

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/AskUserQuestionPermissionRequest/AskUserQuestionPermissionRequest.tsx` (82 KB)

This is the permission dialog specifically for the `AskUserQuestionTool`. Orchestrates a **multi-question flow**:

1. **Question nav bar** — `QuestionNavigationBar.tsx` (23 KB) — shows progress through questions
2. **Question view** — `QuestionView.tsx` (58 KB) — one question at a time; multiple choice or free-form
3. **Preview** — `PreviewQuestionView.tsx` (53 KB) + `PreviewBox.tsx` (26 KB) — side-pane showing what Claude will do with the answer
4. **Submit** — `SubmitQuestionsView.tsx` (16 KB) — final "send all answers" screen
5. **State** — `use-multiple-choice-state.ts` (4 KB)

Constraints from `AskUserQuestionPermissionRequest.tsx:26-29`:
- `MIN_CONTENT_HEIGHT = 12` (rows)
- `MIN_CONTENT_WIDTH = 40` (cols)
- `CONTENT_CHROME_OVERHEAD = 15` (nav bar + title + footer + help text)

The `AskUserQuestionTool` itself (`tools/AskUserQuestionTool/AskUserQuestionTool.ts`) defines the schema: up to **4 questions per call**, up to **4 options per question** (constraint mirrored in moai-adk's AskUserQuestion usage — the schemas are identical).

### 7.4 File Permission Dialog

File: `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/FilePermissionDialog/FilePermissionDialog.tsx` (30 KB)

Props at `:20-47`:
```ts
FilePermissionDialogProps<T extends ToolInput = ToolInput> = {
  toolUseConfirm, toolUseContext, onDone, onReject,
  title, subtitle, question?, content?,  // content can be diff preview
  completionType?, languageName?,
  path, parseInput, operationType?,       // 'read' | 'write' | 'edit'
  ideDiffSupport?,                        // optional IDE integration
  workerBadge                             // team mode badge
}
```

Uses `useFilePermissionDialog` + `usePermissionHandler` hooks to compute permission options. Permission options defined in `permissionOptions.tsx` (22 KB):
- "Yes" (one-time approve)
- "Yes, always for this file" (persistent allow for path)
- "Yes, always for this directory" (persistent allow for dir)
- "No" (reject)
- "Show in IDE" (if `ideDiffSupport`)

### 7.5 Permission Rules UI

Directory: `components/permissions/rules/` — runtime management of persistent rules.

- `PermissionRuleList.tsx` (122 KB!) — the largest permission component — browser for all rules with filter/search/edit
- `PermissionRuleInput.tsx` (17 KB) — input for glob/regex/explicit path rules
- `AddPermissionRules.tsx` (22 KB) — add new rule wizard
- `PermissionRuleDescription.tsx` (6.9 KB) — human-readable rule description
- `RecentDenialsTab.tsx` (19 KB) — review recently-denied requests (one-click add-to-allowlist)
- `WorkspaceTab.tsx` (15 KB) — workspace-scoped rules
- `AddWorkspaceDirectory.tsx` (39 KB), `RemoveWorkspaceDirectory.tsx` (10 KB) — trust management

### 7.6 PermissionDecisionDebugInfo

File: `components/permissions/PermissionDecisionDebugInfo.tsx` (53 KB) — ctrl+d expands this; shows:
- Which rules matched/didn't
- Scope (project/workspace/user/session)
- Rule source file
- Whether rule was granted by default allow, user-added allow, or deny

### 7.7 Shell-Specific Helpers

File: `components/permissions/shellPermissionHelpers.tsx` (22.5 KB) — parses bash commands for dangerous patterns (rm -rf, pipe to shell, curl to eval). Used by both BashPermissionRequest and PowerShellPermissionRequest.

### 7.8 Worker/Teammate Permission Flow

- `WorkerBadge.tsx` (3.9 KB) — shows which teammate is requesting permission (team mode)
- `WorkerPendingPermission.tsx` (9.4 KB) — "Teammate X is waiting for your approval" display
- `utils.ts` + `hooks.ts` — shared utilities

### 7.9 Flow Summary

1. Tool invocation triggers `PermissionRequest` with `(tool, toolUseConfirm)`
2. Dispatcher picks tool-specific component
3. That component (e.g. `FileEditPermissionRequest`) renders a `FilePermissionDialog`
4. Dialog shows preview (diff, command, URL, etc.) + `PermissionPrompt` options
5. User selects via keyboard (numbered 1-N, or arrow+Enter) — optionally with Tab-expanded feedback
6. `onDone(decision)` or `onReject()` callback fires
7. If "always" was chosen, rule is persisted via `PermissionUpdate` → `settings.json` or `.claude/settings.local.json`
8. Tool proceeds or is blocked

---

## 8. Gap vs moai-adk (UI/UX features moai-adk should leverage or surface)

Context from user's brief: moai-adk-go is a **non-interactive Go CLI**. It is invoked *from within* Claude Code (as a subprocess or via hooks) — it is NOT competing with CC's TUI. This gap matrix is about output fidelity, not re-implementation.

### 8.1 What moai-adk Should NOT Reimplement

These are CC-exclusive and should be left alone:

| CC Feature | Why moai-adk should not reimplement |
|---|---|
| Ink TUI stack (ink.tsx, render-node-to-output.ts, Yoga layout) | 500 KB+ fork; Go ecosystem has no equivalent. Text output is sufficient. |
| Keybinding system (18 contexts, chords, hot-reload) | moai-adk is not a REPL; no keybindings. |
| AskUserQuestion multi-question UI | User already instructed: `moai-adk` delegates to CC's AskUserQuestion. |
| Markdown/syntax-highlight renderer | Would require `marked` + syntax lib; Go's equivalent (chroma, goldmark) suffices for simple cases but CC does the heavy lifting when moai-adk output reaches it. |
| Virtual message list, scroll box, fullscreen layout | CC-exclusive. |
| Permission dialog framework | moai-adk has no persistent permissions — CC handles them. |
| Theme provider with OKLCH color resolution | Terminal color is enough for CLI output. |
| Mouse event dispatch (click, hover) | CLI is keyboard-only. |

### 8.2 What moai-adk SHOULD Leverage (CC patterns to surface in output)

These are CC patterns moai-adk's output should respect so it renders correctly inside CC:

| CC Pattern | moai-adk action |
|---|---|
| **AskUserQuestion 4x4 constraint** | Already enforced — moai-adk workflows call `AskUserQuestion` via MoAI orchestrator. Continue this. Source confirms the constraint is canonical (see `AskUserQuestionTool` usage). |
| **Status line integration** | CC has hook-driven status line (`utils/hooks.ts executeStatusLineCommand`). moai-adk's `.moai/status_line.sh.tmpl` is the correct pattern — status lines are shell-executed. |
| **Markdown output** | moai-adk output should be valid Markdown because CC's `Markdown.tsx` will parse it (tables, code blocks, headers, lists). Avoid raw ANSI except where needed. |
| **File path OSC-8 hyperlinks** | CC's `FilePathLink.tsx` (3.2 KB) already makes `file://` paths clickable. moai-adk can emit `file:///absolute/path:LINE` in output and CC will surface them. Avoid relative paths since OSC-8 resolves differently. |
| **StatusIcon semantic colors** | Use ✓ (success), ✗ (error), ⚠ (warning), ℹ (info), ○ (pending), … (loading). These are CC's figures from `figures` npm package. |
| **Progress eighth-block** | If moai-adk renders progress, `▏▎▍▌▋▊▉█` characters are CC's standard — matches `design-system/ProgressBar.tsx:26`. |
| **Structured error format** | CC's `SystemAPIErrorMessage.tsx`, `ValidationErrorsList.tsx` format errors with severity + suggestion. moai-adk's error messages should include severity + suggestion to match visual pattern. |
| **Hook progress messages** | CC renders hook stdout through `messages/HookProgressMessage.tsx`. moai-adk hooks should use the `Status:`/`Task:`/`Progress:` prefixes (user's brief confirms this). CC parses these prefixes. |
| **Slash command hint rendering** | CC's `ConfigurableShortcutHint.tsx` renders key hints; if moai-adk output wants to tell the user "Press ctrl+o to expand", use CC's phrasing convention. |

### 8.3 What moai-adk MIGHT Add (Optional UX Polish)

Lower-priority patterns that would improve output fidelity without duplication:

| Pattern | Note |
|---|---|
| **Conventional-commits-like structured output sections** | CC doesn't enforce, but `# Status` / `## Task` / `### Progress` Markdown headers render well. |
| **`$CLAUDE_PROJECT_DIR` path normalization** | CC hooks receive `$CLAUDE_PROJECT_DIR` env var. moai-adk already quotes these per `CLAUDE.local.md` §7. |
| **JSONL event stream for long tasks** | If moai-adk tasks run >10s, emit periodic JSONL status to stderr. CC's `getRenderContext`:330 shows JSONL is a common pattern (bench mode). |
| **Status prefixes colorized to match CC theme** | Use `\x1b[32m` green for success, `\x1b[31m` red for error, `\x1b[33m` yellow for warning. CC's `Ansi.tsx` will pass them through. |

### 8.4 Gaps Where moai-adk Currently Underdelivers

These are areas where moai-adk's current output may be inadequate compared to what CC renders:

| Gap | Current moai-adk state | Suggested fix |
|---|---|---|
| **No diff visualization** | moai-adk emits plain text diffs | Emit `diff --git` format; CC's `StructuredDiff.tsx` will render it with syntax highlighting. |
| **No structured validation errors** | Errors are plain text | Emit `{ severity, path, line, message, suggestion }` — CC renders these via `ValidationErrorsList.tsx`. Could be Markdown table or YAML. |
| **No progress indicators during long tasks** | Silent during LSP scans, template rendering | Emit `Progress: N/M tasks` lines; CC status line can pick them up. |
| **No status-line contribution** | moai-adk doesn't publish state to CC's status line | moai-adk already uses `.moai/status_line.sh.tmpl` — leverage CC's `StatusLine.tsx:22` hook call for output. |
| **Code blocks lack language hints** | Output often emits raw code without fence language | Always use triple-backtick + language (`\`\`\`go`, `\`\`\`python`); CC's `HighlightedCode.tsx` syntax-highlights via language. |

### 8.5 Output-Style Compatibility

moai-adk ships `.moai/output-styles/` with its own SKILL.md-defined format. CC's loader (`loadOutputStylesDir.ts`) scans the SAME directory if present at project or user level. **Risk**: if moai-adk's SKILL.md format uses frontmatter keys that collide with CC's (`name`, `description`, `keep-coding-instructions`, `force-for-plugin`), CC will interpret them. moai-adk output-styles should:
1. Use the standard CC frontmatter keys (not custom MoAI ones) — or use `$schema` field to distinguish
2. If moai-adk needs additional metadata, use a `moai:` prefix (e.g. `moai:version: 1.0`) to avoid collision
3. Body Markdown is the prompt — CC will treat it as the system prompt extension

### 8.6 Keybinding Compatibility

No moai-adk action. CC owns keybinding entirely. However, **moai-adk documentation should reference CC's default keybindings** accurately so users aren't surprised:

| User Action | Correct CC Keybinding |
|---|---|
| Cancel ongoing task | `escape` (in Chat context) |
| Kill all sub-agents | `ctrl+x ctrl+k` (chord) |
| Toggle transcript | `ctrl+o` |
| Toggle todos | `ctrl+t` |
| Interrupt | `ctrl+c` (double-press) |
| Exit | `ctrl+d` (double-press) |
| Model picker | `meta+p` |

moai-adk's docs already reference `ctrl+c` / `ctrl+d` / `ctrl+x ctrl+k` per user's context; this is correct.

### 8.7 Permission Flow — What moai-adk Can Do

moai-adk tools invoked from CC trigger CC's permission flow (via settings.json `permissions.allow`). moai-adk's obligation:

1. **Pre-approve read-only operations** in `settings.json.tmpl` for common cases (moai-adk already does this)
2. **Emit clear tool use descriptions** so CC's permission dialog previews make sense — e.g. `moai build` should print `Running: go build -tags=...` so when CC asks "Do you want to proceed?", the preview is readable
3. **Avoid arbitrary shell** — moai-adk's hook wrappers already call `moai hook <event>`, which is a fixed command; CC can allowlist this easily

### 8.8 Summary Matrix

| Layer | CC has it | moai-adk should | Effort |
|---|---|---|---|
| Ink TUI | Yes (vendored fork) | Use through CC — emit rich text | — |
| Box/Text/Layout | Yes | N/A | — |
| Markdown renderer | Yes | Emit valid Markdown | Low |
| Syntax highlighting | Yes (via NAPI) | Emit ```` ```language```` blocks | Low |
| Diff viewer | Yes | Emit `diff --git` | Low |
| Status icons | Yes | Use ✓✗⚠ℹ○… | Low |
| Progress bars | Yes | Use eighth-blocks or Markdown tables | Medium |
| Keybindings | Yes (18 contexts, chords, hot-reload) | Document CC keybindings accurately | Low |
| Output styles | Yes (frontmatter+MD) | Use CC-compatible schema in `.moai/output-styles/` | Low |
| Permission system | Yes (32 components, rule mgmt) | Pre-allowlist safe ops in `settings.json.tmpl` | Low |
| AskUserQuestion | Yes (4x4 constraint) | Already enforced via orchestrator | — |
| Status line hooks | Yes (hook-driven, reparses on change) | Already uses `.moai/status_line.sh.tmpl` | — |
| Dialog launchers | Yes (22 launcher functions) | N/A — moai-adk non-interactive | — |
| Permission rules hot-reload | Yes (chokidar on `settings.json`) | Already compatible via `moai update` | — |

---

## 9. Source References (file:line)

### 9.1 Ink Entry & Re-exports

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink.ts:18-23` — `render()` with theme wrapping
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink.ts:25-31` — `createRoot()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink.ts:33-85` — re-exports
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/components/Box.tsx:11-46` — Box props
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/components/Text.tsx:5-59` — Text props (BaseProps + WeightProps bold/dim mutual exclusion)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/ink/components/Button.tsx:10-38` — Button state machine

### 9.2 Interactive Helpers

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:32-38` — `completeOnboarding()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:39-44` — `showDialog()` primitive
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:52-80` — `exitWithError`/`exitWithMessage` (console.error is patched, must go through React)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:86-92` — `showSetupDialog()` (AppStateProvider + KeybindingSetup wrapper)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:98-103` — `renderAndRun()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:104-298` — `showSetupScreens()` full onboarding chain
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/interactiveHelpers.tsx:299-365` — `getRenderContext()` FPS + bench mode

### 9.3 Dialog Launchers

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:29-38` — `launchSnapshotUpdateDialog`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:44-52` — `launchInvalidSettingsDialog`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:58-65` — `launchAssistantSessionChooser`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:73-85` — `launchAssistantInstallWizard` (Promise.race error handling)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:91-96` — `launchTeleportResumeWrapper`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:102-110` — `launchTeleportRepoMismatchDialog`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/dialogLaunchers.tsx:117-131` — `launchResumeChooser` (uses `renderAndRun`, not `showSetupDialog`)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/replLauncher.tsx:12-22` — `launchRepl()`

### 9.4 Keybindings

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:15` — `IMAGE_PASTE_KEY` platform-adaptive
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:21-30` — `SUPPORTS_TERMINAL_VT_MODE` (Node 22.17 / Bun 1.2.23)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:32-340` — full default binding table (18 contexts)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:68` — `ctrl+x ctrl+k` chord for `chat:killAgents`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/defaultBindings.ts:40-41` — ctrl+c / ctrl+d double-press comment
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/schema.ts:12-33` — KEYBINDING_CONTEXTS enum (18 values)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/schema.ts:64-172` — KEYBINDING_ACTIONS enum
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/schema.ts:177-229` — Zod schema for keybindings.json
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/schema.ts:193-200` — command binding regex `^command:[a-zA-Z0-9:\-_]+$`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/reservedShortcuts.ts:16-33` — NON_REBINDABLE (ctrl+c, ctrl+d, ctrl+m)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/reservedShortcuts.ts:43-54` — TERMINAL_RESERVED
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/reservedShortcuts.ts:59-67` — MACOS_RESERVED
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/parser.ts:13-75` — `parseKeystroke()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/parser.ts:80-84` — `parseChord()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/parser.ts:156-176` — platform-adaptive display strings (opt/alt, cmd/super)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/resolver.ts:32-61` — `resolveKey()` single-keystroke
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/resolver.ts:166-244` — `resolveKeyWithChordState()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/resolver.ts:107-118` — `keystrokesEqual()` (collapses alt+meta)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/loadUserBindings.ts:41-46` — `isKeybindingCustomizationEnabled()` gated by GrowthBook
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/loadUserBindings.ts:115-117` — `getKeybindingsPath()` → `~/.claude/keybindings.json`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/loadUserBindings.ts:353-404` — chokidar watcher setup
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/useKeybinding.ts:33-97` — `useKeybinding()` single-action
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/useKeybinding.ts:113-196` — `useKeybindings()` multi-action
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/KeybindingContext.tsx:59` — `KeybindingProvider`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/keybindings/KeybindingContext.tsx:184-215` — hooks `useKeybindingContext` / `useOptionalKeybindingContext` / `useRegisterKeybindingContext`

### 9.5 Output Styles

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts:14-24` — comment on structure (project overrides user)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts:26-92` — `getOutputStyleDirStyles()` memoized loader
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts:53-62` — `keep-coding-instructions` parsing
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts:64-70` — `force-for-plugin` warning for non-plugin
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/outputStyles/loadOutputStylesDir.ts:94-98` — `clearOutputStyleCaches()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/OutputStylePicker.tsx:11-12` — Default style name + description
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/OutputStylePicker.tsx:48-57` — fallback to built-ins on error

### 9.6 Permissions

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionRequest.tsx:47-81` — tool → component dispatcher
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionPrompt.tsx:10-18` — `PermissionPromptOption` type
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionPrompt.tsx:30-33` — DEFAULT_PLACEHOLDERS for feedback
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/PermissionPrompt.tsx:82-119` — feedback mode Tab expansion
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/AskUserQuestionPermissionRequest/AskUserQuestionPermissionRequest.tsx:26-29` — MIN_CONTENT_HEIGHT=12, MIN_CONTENT_WIDTH=40, CHROME_OVERHEAD=15
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/permissions/FilePermissionDialog/FilePermissionDialog.tsx:20-47` — FilePermissionDialogProps

### 9.7 Design System

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/Dialog.tsx:11-29` — DialogProps (title/subtitle/children/onCancel/color/hideInputGuide/hideBorder/inputGuide/isCancelActive)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ProgressBar.tsx:26` — BLOCKS array (`▏▎▍▌▋▊▉█`)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/StatusIcon.tsx:5-52` — STATUS_CONFIG (success/error/warning/info/pending/loading)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ListItem.tsx:7-56` — ListItemProps (isFocused/isSelected/description/showScrollDown/showScrollUp/styled/disabled)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/FuzzyPicker.tsx:14-62` — FuzzyPicker Props (getKey/renderItem/renderPreview/previewPosition/visibleCount/direction/onSelect/onTab/onShiftTab)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ThemedBox.tsx:42-50` — resolveColor for borders/background
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ThemedText.tsx:66-74` — resolveColor for text
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/Tabs.tsx:66,261,289,307` — exports: `Tabs`, `Tab`, `useTabsWidth`, `useTabHeaderFocus`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/design-system/ThemeProvider.tsx:122-147` — `useTheme`, `useThemeSetting`, `usePreviewTheme`

### 9.8 Content Rendering

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/Markdown.tsx:22-71` — LRU token cache (TOKEN_CACHE_MAX=500)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/Markdown.tsx:31-36` — MD_SYNTAX_RE fast-path
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/HighlightedCode.tsx:17` — DEFAULT_WIDTH = 80
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/HighlightedCode.tsx:40-58` — ColorFile via `expectColorFile()`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/StructuredDiff.tsx:32-41` — WeakMap render cache
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/StructuredDiff.tsx:46-49` — computeGutterWidth

### 9.9 Inputs

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/CustomSelect/select.tsx:28-69` — OptionWithDescription (type: 'text' | 'input' variants)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/CustomSelect/select.tsx:70-120` — SelectProps
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/CustomSelect/SelectMulti.tsx:11-56` — SelectMultiProps (submitButtonText, onSubmit, initialFocusLast)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/TextInput.tsx:16-33` — voice waveform bar constants (BARS, SMOOTH, LEVEL_BOOST, SILENCE_THRESHOLD)

### 9.10 Status Line

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/StatusLine.tsx:30-35` — `statusLineShouldDisplay`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/StatusLine.tsx:36-60` — `buildStatusLineCommandInput` (permissionMode, tokens, rate limits)

### 9.11 Spinner

- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/Spinner.tsx:39-57` — Props (mode, loadingStartTimeRef, totalPausedMsRef, pauseStartTimeRef, spinnerTip, responseLengthRef)
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/Spinner.tsx:62-80` — SpinnerWithVerb branching on `isBriefOnly`
- `/Users/goos/MoAI/AgentOS/claude-code-source-map/components/Spinner/index.ts:1-10` — exports (FlashingChar, GlimmerMessage, ShimmerChar, SpinnerGlyph, SpinnerMode, useShimmerAnimation, useStalledAnimation, getDefaultCharacters, interpolateColor)

---

## End Matter

**Scope coverage**: 50 Ink files + 146 components + 4 helper files (interactiveHelpers, dialogLaunchers, replLauncher, ink.ts) + 15 keybinding files + 1 output-style loader = 216+ source files surveyed. Total bytes read: ~18 KB of targeted reads + structural surveys of ~5 MB of source.

**Key confidence statements**:
- Ink TUI is a vendored fork (not upstream npm `ink`) — confirmed by directory structure and 252 KB `ink/ink.tsx` core.
- Keybinding customization is **gated behind an Anthropic-internal GrowthBook flag** — external users currently get defaults only. `loadUserBindings.ts:41-46`.
- AskUserQuestion 4x4 constraint is canonical (4 questions × 4 options) — matches moai-adk's existing constraint.
- Output style frontmatter keys (`name`, `description`, `keep-coding-instructions`, `force-for-plugin`) are the standard CC schema; moai-adk should use this format for compat.
- Chord bindings use space-separated keystroke syntax (`"ctrl+x ctrl+k"`) and are first-class in the schema + resolver.
- 18 keybinding contexts × ~85 actions → rich vocabulary that moai-adk docs should reference accurately.

**Out of scope for this wave** (already covered by other Wave 1 researchers):
- Hook system integration with the TUI (Wave 1.1)
- Query/Context window management (Wave 1.2)
- Agent teammate UI (Wave 1.3 — covers CoordinatorAgentStatus, TeammateViewHeader, etc.)
- CLI bootstrap (Wave 1.5)
