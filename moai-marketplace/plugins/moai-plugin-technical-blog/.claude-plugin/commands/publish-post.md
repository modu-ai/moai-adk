# publish-post
Publish blog post to multiple platforms.

## Usage
```
/publish-post [file] [--platforms PLATFORMS] [--schedule DATE]
```

## Parameters
- `file`: Markdown blog post file
- `--platforms`: Comma-separated list (wordpress,medium,dev-to,naver)
- `--schedule`: Publish date/time

## What It Does
1. Parses markdown content
2. Converts to platform-specific formats
3. Publishes to selected platforms
4. Schedules posts
5. Tracks publishing status

## Example
```bash
/publish-post blog/my-article.md --platforms wordpress,medium
```
