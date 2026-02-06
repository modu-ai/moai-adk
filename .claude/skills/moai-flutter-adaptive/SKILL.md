---
name: flutter-adaptive-ui
description: Build adaptive and responsive Flutter UIs that work beautifully across all platforms and screen sizes. Use when creating Flutter apps that need to adapt layouts based on screen size, support multiple platforms including mobile tablet desktop and web, handle different input devices like touch mouse and keyboard, implement responsive navigation patterns, optimize for large screens and foldables, or use Capability and Policy patterns for platform-specific behavior.
---

# Flutter Adaptive UI

## Overview

Create Flutter applications that adapt gracefully to any screen size, platform, or input device. This skill provides comprehensive guidance for building responsive layouts that scale from mobile phones to large desktop displays while maintaining excellent user experience across touch, mouse, and keyboard interactions.

## Quick Reference

**Core Layout Rule:** Constraints go down. Sizes go up. Parent sets position.

**3-Step Adaptive Approach:**
1. **Abstract** - Extract common data from widgets
2. **Measure** - Determine available space (MediaQuery/LayoutBuilder)
3. **Branch** - Select appropriate UI based on breakpoints

**Key Breakpoints:**
* Compact (Mobile): width < 600
* Medium (Tablet): 600 <= width < 840  
* Expanded (Desktop): width >= 840

## Adaptive Workflow

Follow the 3-step approach to make your app adaptive.

### Step 1: Abstract

Identify widgets that need adaptability and extract common data. Common patterns:
- Navigation UI (switch between bottom bar and side rail)
- Dialogs (fullscreen on mobile, modal on desktop)
- Content lists (reflow from single to multi-column)

For navigation, create a shared `Destination` class with icon and label used by both `NavigationBar` and `NavigationRail`.

### Step 2: Measure

Choose the right measurement tool:

**MediaQuery.sizeOf(context)** - Use when you need app window size for top-level layout decisions
- Returns entire app window dimensions
- Better performance than `MediaQuery.of()` for size queries
- Rebuilds widget when window size changes

**LayoutBuilder** - Use when you need constraints for specific widget subtree
- Provides parent widget's constraints as `BoxConstraints`
- Local sizing information, not global window size
- Returns min/max width and height ranges

Example:
```dart
// For app-level decisions
final width = MediaQuery.sizeOf(context).width;

// For widget-specific constraints
LayoutBuilder(
  builder: (context, constraints) {
    if (constraints.maxWidth < 600) {
      return MobileLayout();
    }
    return DesktopLayout();
  },
)
```

### Step 3: Branch

Apply breakpoints to select appropriate UI. Don't base decisions on device type - use window size instead.

Example breakpoints (from Material guidelines):
```dart
class AdaptiveLayout extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;
    
    if (width >= 840) {
      return DesktopLayout();
    } else if (width >= 600) {
      return TabletLayout();
    }
    return MobileLayout();
  }
}
```

## Layout Fundamentals

### Understanding Constraints

Flutter layout follows one rule: **Constraints go down. Sizes go up. Parent sets position.**

Widgets receive constraints from parents, determine their size, then report size up to parent. Parents then position children.

Key limitation: Widgets can only decide size within parent constraints. They cannot know or control their own position.

For detailed examples and edge cases, see [layout-constraints.md](references/layout-constraints.md).

### Common Layout Patterns

**Row/Column**
- `Row` arranges children horizontally
- `Column` arranges children vertically
- Control alignment with `mainAxisAlignment` and `crossAxisAlignment`
- Use `Expanded` to make children fill available space proportionally

**Container**
- Add padding, margins, borders, background
- Can constrain size with width/height
- Without child/size, expands to fill constraints

**Expanded/Flexible**
- `Expanded` forces child to use available space
- `Flexible` allows child to use available space but can be smaller
- Use `flex` parameter to control proportions

For complete widget documentation, see [layout-basics.md](references/layout-basics.md) and [layout-common-widgets.md](references/layout-common-widgets.md).

## Best Practices

### Design Principles

**Break down widgets**
- Create small, focused widgets instead of large complex ones
- Improves performance with `const` widgets
- Makes testing and refactoring easier
- Share common components across different layouts

**Design to platform strengths**
- Mobile: Focus on capturing content, quick interactions, location awareness
- Tablet/Desktop: Focus on organization, manipulation, detailed work
- Web: Leverage deep linking and easy sharing

**Solve touch first**
- Start with great touch UI
- Test frequently on real mobile devices
- Layer on mouse/keyboard as accelerators, not replacements

### Implementation Guidelines

**Never lock orientation**
- Support both portrait and landscape
- Multi-window and foldable devices require flexibility
- Locked screens can be accessibility issues

**Avoid device type checks**
- Don't use `Platform.isIOS`, `Platform.isAndroid` for layout decisions
- Use window size instead
- Device type ≠ window size (windows, split screens, PiP)

**Use breakpoints, not orientation**
- Don't use `OrientationBuilder` for layout changes
- Use `MediaQuery.sizeOf` or `LayoutBuilder` with breakpoints
- Orientation doesn't indicate available space

**Don't fill entire width**
- On large screens, avoid full-width content
- Use multi-column layouts with `GridView` or flex patterns
- Constrain content width for readability

**Support multiple inputs**
- Implement keyboard navigation for accessibility
- Support mouse hover effects
- Handle focus properly for custom widgets

For complete best practices, see [adaptive-best-practices.md](references/adaptive-best-practices.md).

## Capabilities and Policies

Separate what your code *can* do from what it *should* do.

**Capabilities** (what code can do)
- API availability checks
- OS-enforced restrictions
- Hardware requirements (camera, GPS, etc.)

**Policies** (what code should do)
- App store guidelines compliance
- Design preferences
- Platform-specific features
- Feature flags

### Implementation Pattern

```dart
// Capability class
class Capability {
  bool hasCamera() {
    // Check if camera API is available
    return Platform.isAndroid || Platform.isIOS;
  }
}

// Policy class
class Policy {
  bool shouldShowCameraFeature() {
    // Business logic - maybe disabled by store policy
    return hasCamera() && !Platform.isIOS;
  }
}
```

Benefits:
- Clear separation of concerns
- Easy to test (mock Capability/Policy independently)
- Simple to update when platforms evolve
- Business logic doesn't depend on device detection

For detailed examples, see [adaptive-capabilities.md](references/adaptive-capabilities.md) and [capability_policy_example.dart](assets/capability_policy_example.dart).

## Examples

### Responsive Navigation

Switch between bottom navigation (small screens) and navigation rail (large screens):

```dart
Widget build(BuildContext context) {
  final width = MediaQuery.sizeOf(context).width;
  
  return width >= 600 
    ? _buildNavigationRailLayout()
    : _buildBottomNavLayout();
}
```

Complete example: [responsive_navigation.dart](assets/responsive_navigation.dart)

### Adaptive Grid

Use `GridView.extent` with responsive maximum width:

```dart
LayoutBuilder(
  builder: (context, constraints) {
    return GridView.extent(
      maxCrossAxisExtent: constraints.maxWidth < 600 ? 150 : 200,
      // ...
    );
  },
)
```

## Resources

### Reference Documentation
- [layout-constraints.md](references/layout-constraints.md) - Complete guide to Flutter's constraint system with 29 examples
- [layout-basics.md](references/layout-basics.md) - Core layout widgets and patterns
- [layout-common-widgets.md](references/layout-common-widgets.md) - Container, GridView, ListView, Stack, Card, ListTile
- [adaptive-workflow.md](references/adaptive-workflow.md) - Detailed 3-step adaptive design approach
- [adaptive-best-practices.md](references/adaptive-best-practices.md) - Design and implementation guidelines
- [adaptive-capabilities.md](references/adaptive-capabilities.md) - Capability/Policy pattern for platform behavior

### Example Code
- [responsive_navigation.dart](assets/responsive_navigation.dart) - NavigationBar ↔ NavigationRail switching
- [capability_policy_example.dart](assets/capability_policy_example.dart) - Capability/Policy class examples

### Scripts
This skill currently has no executable scripts. All guidance is in reference documentation.

### Assets
This skill includes complete Dart example files demonstrating:
- Responsive navigation patterns
- Capability and Policy implementation
- Adaptive layout strategies

These assets can be copied directly into your Flutter project or adapted to your needs.
