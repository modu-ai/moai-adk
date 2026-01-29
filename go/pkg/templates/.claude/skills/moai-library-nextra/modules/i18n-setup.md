# Internationalization Setup

## Overview

Multi-language documentation setup and management.

## Configuration

### I18n Config

```javascript
// next.config.js
module.exports = {
 i18n: {
 locales: ['en', 'ko', 'ja', 'zh'],
 defaultLocale: 'en'
 }
}
```

### Directory Structure

```
docs/
 en/
 index.md
 guide.md
 ko/
 index.md
 guide.md
 ja/
 index.md
 guide.md
```

## Translation Workflow

### Translation Files

```json
{
 "en": {
 "welcome": "Welcome",
 "guide": "Guide"
 },
 "ko": {
 "welcome": "",
 "guide": ""
 }
}
```

---
Last Updated: 2025-11-23
Status: Production Ready
