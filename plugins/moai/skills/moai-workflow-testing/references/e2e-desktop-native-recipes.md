# Desktop-Native E2E Recipes (OS accessibility lane)

Per-OS automation recipes for native desktop toolkits (AppKit, WinUI/Win32, Qt, GTK), which are driven through the host OS accessibility layer rather than a browser engine. All three OS recipes below are documented; only the recipe matching the HOST OS is probed and executed — a recipe whose target OS differs from the host OS is declarative documentation for that host, and the host-OS/target-OS mismatch is stated in the report instead of being probed.

Scripts and flows live under `e2e/desktop-native/`; AX-tree snapshots and run logs ride the `e2e/.runs/<timestamp>-<slug>.log` timestamped-log convention.

## macOS recipe — axcli (default) / appium-mac2 + WebdriverIO (fallback)

- Default: `axcli` — AXUIElement tree snapshots + background-safe UI actions with Playwright-like selectors. Install: `cargo install axcli` (young project — PIN the version and record it in the flow header). Probe: `axcli --version`.
- Fallback: appium-mac2-driver + WebdriverIO — reuses the existing Tauri WDIO lane; requires Xcode. Install: `npm i -g appium && appium driver install mac2`. Probe: `appium driver list --installed`.
- Prerequisite — Accessibility permission (TCC): the executing terminal/host process must be granted macOS Accessibility permission before any AX-tree read. When the grant is missing, surface the grant path (System Settings → Privacy & Security → Accessibility) and return a structured blocker report — never fail silently, never prompt the user.

## Windows recipe — FlaUI.WebDriver (default) / pywinauto (fallback)

- Default: FlaUI.WebDriver + WebdriverIO — W3C WebDriver2 over UIA3; FlaUI.WebDriver is EXPERIMENTAL (PIN the release, v0.4.0), so smoke-probe the running server with `GET /status` before any session.
- Fallback: pywinauto — `pip install pywinauto`; `print_control_identifiers()` is the UIA tree dump. Probe: `python -c "import pywinauto"`.

## Linux recipe — dogtail (default) / ydotool + xdotool blind injection (fallback)

- Default: dogtail 2.x over AT-SPI2. Prerequisites: distribution at-spi2 packages installed; Qt apps additionally require `QT_LINUX_ACCESSIBILITY_ALWAYS_ON=1` (without it the AT-SPI tree is empty). Wayland caveat: dogtail's Wayland support is GNOME-only (via ponytail); route non-GNOME Wayland desktops to the fallback. Probe: `python -c "import dogtail"`.
- Fallback: blind input injectors — ydotool (Wayland) / xdotool (X11) — PAIRED with screenshot verification (blind injection without verification is not a recipe). Probe: `ydotool --version`.

## Evidence-source and token-cost ordering (all OS)

- Cross-OS floor: the AX-tree text snapshot loop — a FILTERED accessibility-tree text read costs hundreds of tokens per read and is the first-choice evidence source on every OS.
- Screenshot loop: the computer-use screenshot loop costs ~1.1-1.6K tokens/frame and is non-deterministic — reserved for FINAL visual evidence artifacts only; NOT acceptable as CI-repeatable acceptance evidence.
- Token-cost ordering (hard): filtered AX-tree text snapshot ≪ full tree JSON < single screenshot < screenshot loop.
