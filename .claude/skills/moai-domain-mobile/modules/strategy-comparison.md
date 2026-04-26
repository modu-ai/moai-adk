# Mobile Paradigm Strategy Comparison

Objective selection guide for choosing between iOS native, Android native, React Native, and Flutter. This guide presents facts and trade-offs without advocating for any specific paradigm.

## Comparison Table

| Criterion | iOS Native | Android Native | React Native | Flutter |
|-----------|------------|----------------|--------------|---------|
| **Performance** | Best (metal access, no bridge) | Best (JNI, no bridge) | Good (New Arch, JSI bridge) | Very Good (Skia/Impeller, no JS) |
| **Development cost** | High (iOS-only) | High (Android-only) | Medium (shared JS/TS) | Low (shared Dart) |
| **Team skill required** | Swift / Objective-C | Kotlin / Java | JavaScript / TypeScript | Dart (new language) |
| **Platform features** | Full (all Apple APIs) | Full (all Google APIs) | Partial (via modules/plugins) | Full (via plugins) |
| **Code reuse %** | 0% (iOS only) | 0% (Android only) | 80-90% (iOS + Android) | 90-95% (iOS + Android + Web*) |
| **Hot reload** | No (re-compile needed) | No (re-compile needed) | Yes (Fast Refresh) | Yes (Hot Reload / Hot Restart) |
| **UI consistency** | Platform-native look | Platform-native look | Platform-native (RN components) | Pixel-perfect custom (Skia) |
| **App size** | Small | Small | Medium (+JS bundle) | Medium (+Dart + Skia) |
| **Community size** | Large (Apple developer) | Large (Google developer) | Very large (JS ecosystem) | Growing (Google-backed) |
| **Learning curve** | High | High | Medium (if JS known) | Medium (Dart is learnable) |

*Flutter web is production-ready but performance varies; not a replacement for dedicated web apps.

---

## Selection Criteria

### Criterion 1: Target Platform

- **iOS only** (iPhone, iPad, Apple Watch, Vision Pro) → iOS Native
- **Android only** (Android phones, Wear OS, Android TV) → Android Native
- **Both iOS + Android** → React Native or Flutter (see other criteria)
- **iOS + Android + Web** → Flutter (with caveats) or separate codebases

### Criterion 2: Team Existing Skill

- Swift/Objective-C expertise → iOS Native
- Kotlin/Java expertise → Android Native
- JavaScript/TypeScript expertise → React Native
- No mobile experience, willing to learn → Flutter (fastest ramp-up from scratch)
- Mixed team with no mobile experience → React Native if web devs, Flutter if starting fresh

### Criterion 3: Performance Requirements

- 3D graphics, custom rendering, complex animations → Native (iOS or Android)
- Standard UI with smooth 60 fps → All options viable; Flutter closest to native
- Large lists, background processing, camera/AR → Native or Flutter preferred
- Simple CRUD app, forms, content display → All options viable

### Criterion 4: Platform Feature Depth

- Deep Apple Watch / VisionOS integration → iOS Native
- Deep Wear OS / AndroidTV integration → Android Native
- Standard notifications, camera, location → All options via plugins
- Custom hardware integration (Bluetooth LE, NFC, ARCore, ARKit) → Native preferred; RN/Flutter via custom modules

### Criterion 5: Code Reuse and Maintenance Cost

- Long-term: two separate native codebases cost ~2x maintenance
- React Native: 80-90% shared code; platform-specific code for UI nuances
- Flutter: 90-95% shared code; platform channels only for deep native access
- For MVPs and startups: React Native or Flutter reduces time-to-market significantly

---

## Decision Flowchart

```
Start
  ↓
Single platform required?
  ├── Yes, iOS-first → iOS Native
  ├── Yes, Android-first → Android Native
  └── No, cross-platform needed
        ↓
        Existing JS/TS team?
          ├── Yes → React Native (Expo)
          └── No
                ↓
                Pixel-perfect custom UI?
                  ├── Yes → Flutter
                  └── No, standard UI
                        ↓
                        Team willing to learn Dart?
                          ├── Yes → Flutter
                          └── No → React Native
```

---

## Known Trade-offs

### React Native Trade-offs

Advantages:
- Massive JS ecosystem (npm packages)
- Existing web developers can contribute
- Expo simplifies setup and distribution

Disadvantages:
- Bridge overhead (mitigated by New Architecture / JSI)
- Platform-specific bug hunting (iOS vs Android behavior differences)
- React Native upgrades can be disruptive

### Flutter Trade-offs

Advantages:
- Consistent pixel-perfect UI across platforms
- No JS bridge — compiled to native ARM
- Strong typing with Dart (sound null safety)

Disadvantages:
- Dart is a new language for most teams
- Larger app binary size vs native
- Some platform APIs require platform channel code

### Native Trade-offs

Advantages:
- Best performance, full platform API access
- No abstraction layer — direct OS integration
- Platform-idiomatic UX

Disadvantages:
- Two separate codebases (iOS + Android)
- Higher development and maintenance cost
- No code reuse between platforms

---

## Recommendation Protocol

This guide intentionally avoids making a single recommendation. The right choice depends on project-specific constraints. When consulting this guide:

1. List the project's top 3 selection criteria
2. Score each paradigm against those criteria
3. Present trade-offs to stakeholders before finalizing
4. Document the decision rationale in the SPEC

For expert consultation, invoke `expert-mobile` with the project requirements and it will apply this guide systematically.
