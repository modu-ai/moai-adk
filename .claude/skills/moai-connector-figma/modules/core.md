  );
}

if (config.needsTokens) {
  requests.push(
    mcp__figma__get_variable_defs({
      fileId: "abc123xyz"
    })
  );
}

const results = await Promise.all(requests);
```

**Benefit**: Reduce API calls by 30-50% based on actual needs

### Parameter Guidelines & Validation

#### Required Parameters

**`dirForAssetWrites`** (CRITICAL - Common Error Source)

```typescript
// ❌ WRONG: Will cause 400 Bad Request
const context = await mcp__figma__get_design_context({
  nodeId: "689:1242",
  clientLanguages: "typescript"
  // Missing dirForAssetWrites!
});

// ✅ CORRECT: Specify asset output directory
const context = await mcp__figma__get_design_context({
  nodeId: "689:1242",
  clientLanguages: "typescript",
  dirForAssetWrites: "/tmp/figma-assets" // Required even if not using assets
});
```

**Why**: MCP tool needs to know where to write exported assets (even if not used)

#### NodeId Format Validation

```typescript
// NodeId examples (format: "parent-id:component-id")
const validNodeIds = [
  "689:1242",        // Simple component
  "0:1",             // Page/root
  "689:1242:5678",   // Nested instance
  "I123:456:789"     // Copy instance
];

// Validation pattern
function validateNodeId(nodeId: string): boolean {
  // Format: alphanumeric:digit, optionally nested
  return /^[a-zA-Z0-9]+:[0-9]+(:[0-9a-zA-Z:]+)?$/.test(nodeId);
}

if (!validateNodeId(nodeId)) {
  throw new Error(`Invalid nodeId format: ${nodeId}`);
}
```

#### ClientLanguages/Frameworks Auto-Detection

```typescript
// Auto-detect from project context
function detectFramework(projectPath: string): string {
  const packageJson = require(`${projectPath}/package.json`);

  if (packageJson.dependencies?.react) {
    return packageJson.dependencies.typescript ? "typescript" : "javascript";
  }

  if (packageJson.dependencies?.vue) {
    return "typescript"; // Vue 3 + TS recommended
  }

  if (packageJson.dependencies?.["@angular/core"]) {
    return "typescript"; // Angular always TS
  }

  return "typescript"; // Default
}

// Usage
const clientLanguages = detectFramework("./");
const context = await mcp__figma__get_design_context({
  nodeId: "689:1242",
  clientLanguages, // Auto-detected
  dirForAssetWrites: "./src/generated"
});
```


## Level 3: Advanced Patterns & Error Handling

### Error Handling Strategies

#### Common Error: 400 Bad Request - Missing dirForAssetWrites

```typescript
// Error symptoms:
// - 400 Bad Request
// - "dirForAssetWrites is required"
// - Asset export fails silently

// Solution
try {
  const context = await mcp__figma__get_design_context({
    nodeId: "689:1242",
    clientLanguages: "typescript",
    dirForAssetWrites: "./src/generated/figma-assets" // Add this!
  });
} catch (error) {
  if (error.message.includes("dirForAssetWrites")) {
    console.error("Missing dirForAssetWrites parameter");
    // Provide default asset directory
    return await mcp__figma__get_design_context({
      nodeId: "689:1242",
      clientLanguages: "typescript",
      dirForAssetWrites: "/tmp/figma-assets" // Fallback
    });
  }
  throw error;
}
```

#### Rule: Separate get_screenshot and get_variable_defs

**Do NOT call in sequence unless necessary** - Use parallel calls instead:

```typescript
// ❌ INEFFICIENT: Sequential calls
const screenshot1 = await mcp__figma__get_screenshot({nodeId: "id1"});
const vars1 = await mcp__figma__get_variable_defs({fileId: "file1"});
const screenshot2 = await mcp__figma__get_screenshot({nodeId: "id2"});
const vars2 = await mcp__figma__get_variable_defs({fileId: "file1"});
// Total time: 16-20s (sequential)

// ✅ EFFICIENT: Parallel calls grouped by type
const [screenshots, variables] = await Promise.all([
  Promise.all([
    mcp__figma__get_screenshot({nodeId: "id1"}),
    mcp__figma__get_screenshot({nodeId: "id2"})
  ]),
  mcp__figma__get_variable_defs({fileId: "file1"})
]);
// Total time: 3-4s (parallel)
```

**Benefits**: 4-5x faster for batch operations

#### Rate Limiting Handling

```typescript
// Exponential backoff for rate-limited requests
async function callWithBackoff(
  fn: () => Promise<any>,
  maxRetries = 3,
  initialDelay = 1000
): Promise<any> {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await fn();
    } catch (error) {
      if (error.status === 429) { // Rate limited
        const delay = initialDelay * Math.pow(2, attempt);
        console.log(`Rate limited. Retrying in ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
      } else {
        throw error;
      }
    }
  }
  throw new Error(`Max retries exceeded`);
}

// Usage
const screenshot = await callWithBackoff(() =>
  mcp__figma__get_screenshot({nodeId: "689:1242"})
);
```

### Performance Optimization Tips

#### Caching Strategy

```typescript
// Cache metadata with different TTLs based on change frequency
const cacheConfig = {
  metadata: { ttl: 72 * 3600 }, // Design rarely changes (72h)
  variables: { ttl: 24 * 3600 }, // Tokens updated daily (24h)
  screenshots: { ttl: 6 * 3600 }, // Visual assets change frequently (6h)
  components: { ttl: 48 * 3600 } // Component structure stable (48h)
};

// Implementation
const cache = new Map();

async function getWithCache(key: string, fetcher: () => Promise<any>, ttl: number) {
  const cached = cache.get(key);
  const now = Date.now();

  if (cached && (now - cached.timestamp) < (ttl * 1000)) {
