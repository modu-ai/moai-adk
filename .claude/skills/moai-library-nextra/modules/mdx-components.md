# MDX Components

## Overview

Reusable MDX components for enhanced documentation.

## Core Components

### Callout Component

```mdx
<Callout type="info">
 Important information here
</Callout>

<Callout type="warning">
 Warning message
</Callout>
```

### Code Block Component

```mdx
<CodeBlock language="python" highlight="2,4-6">
{`
def example():
 # Highlighted line
 result = process()
 # More highlighted lines
 return result
`}
</CodeBlock>
```

## Custom Components

### Tabs Component

```mdx
<Tabs>
 <Tab label="Python">
 Python code example
 </Tab>
 <Tab label="JavaScript">
 JavaScript code example
 </Tab>
</Tabs>
```

---
Last Updated: 2025-11-23
Status: Production Ready
