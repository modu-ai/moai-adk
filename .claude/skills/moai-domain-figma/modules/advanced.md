    console.log(`Cache hit for ${key}`);
    return cached.value;
  }

  const value = await fetcher();
  cache.set(key, { value, timestamp: now });
  return value;
}

// Usage
const variables = await getWithCache(
  `variables:abc123`,
  () => mcp__figma__get_variable_defs({fileId: "abc123"}),
  cacheConfig.variables.ttl
);
```

#### Batch Processing Optimization

```typescript
// Process components in optimal batch sizes (10-20 per batch)
async function exportComponentsBatch(
  nodeIds: string[],
  batchSize = 15
): Promise<any[]> {
  const results = [];

  for (let i = 0; i < nodeIds.length; i += batchSize) {
    const batch = nodeIds.slice(i, i + batchSize);

    // Parallel requests within batch
    const batchResults = await Promise.all(
      batch.map(nodeId =>
        mcp__figma__get_design_context({
          nodeId,
          clientLanguages: "typescript",
          dirForAssetWrites: "./src/generated"
        })
      )
    );

    results.push(...batchResults);

    // Small delay between batches to respect rate limits
    if (i + batchSize < nodeIds.length) {
      await new Promise(resolve => setTimeout(resolve, 100));
    }
  }

  return results;
}

// Usage: Export 150 components in 10 parallel requests
const allComponents = await exportComponentsBatch(
  Array.from({length: 150}, (_, i) => `component:${i}`)
);
```


## Level 4: Design System Integration Workflow

### Complete Design-to-Code Pipeline

```typescript
import * as fs from "fs";
import * as path from "path";

interface DesignSystemConfig {
  figmaFileId: string;
  figmaTeamId?: string;
  outputDir: string;
  componentNodeIds: string[];
  clientLanguages: "typescript" | "javascript";
}

async function syncDesignSystem(config: DesignSystemConfig) {
  const {figmaFileId, outputDir, componentNodeIds, clientLanguages} = config;

  console.log(`Starting design system sync for ${componentNodeIds.length} components...`);

  // Phase 1: Extract design tokens
  console.log("Phase 1: Extracting design tokens...");
  const variables = await mcp__figma__get_variable_defs({
    fileId: figmaFileId,
    teamId: config.figmaTeamId
  });

  const tokensOutput = path.join(outputDir, "tokens.json");
  fs.writeFileSync(tokensOutput, JSON.stringify(variables, null, 2));
  console.log(`✓ Tokens exported to ${tokensOutput}`);

  // Phase 2: Generate component code (parallel batch processing)
  console.log("Phase 2: Generating component code...");
  const components = await exportComponentsBatch(componentNodeIds, 15);

  const componentsDir = path.join(outputDir, "components");
  fs.mkdirSync(componentsDir, {recursive: true});

  components.forEach((component, index) => {
    const componentFile = path.join(
      componentsDir,
      `${component.componentName || `Component-${index}`}.ts`
    );
    fs.writeFileSync(componentFile, component.generatedCode || "");
  });

  console.log(`✓ Generated ${components.length} components`);

  // Phase 3: Export visual assets (parallel)
  console.log("Phase 3: Exporting visual assets...");
  const screenshots = await Promise.all(
    componentNodeIds.slice(0, 10).map(nodeId => // Limit to 10 for performance
      mcp__figma__get_screenshot({
        nodeId,
        format: "png",
        scale: 2
      })
    )
  );

  const assetsDir = path.join(outputDir, "assets");
  fs.mkdirSync(assetsDir, {recursive: true});

  screenshots.forEach((screenshot, index) => {
    const assetFile = path.join(assetsDir, `component-${index}.png`);
    fs.writeFileSync(assetFile, screenshot.imageData);
  });

  console.log(`✓ Exported ${screenshots.length} component previews`);

  // Phase 4: Generate documentation
  console.log("Phase 4: Generating documentation...");
  const docMarkdown = generateComponentDocs(components, variables);
  const docsFile = path.join(outputDir, "COMPONENTS.md");
  fs.writeFileSync(docsFile, docMarkdown);

  console.log(`✓ Documentation generated at ${docsFile}`);
  console.log(`\nDesign system sync complete!`);
  console.log(`Output directory: ${outputDir}`);
}

// Helper function
function generateComponentDocs(components: any[], tokens: any[]): string {
  let md = "# Auto-Generated Component Documentation\n\n";
  md += `Generated: ${new Date().toISOString()}\n\n`;

  md += "## Design Tokens\n\n";
  tokens.forEach(token => {
    md += `- \`${token.name}\`: ${token.value}\n`;
  });

  md += "\n## Components\n\n";
  components.forEach((comp, i) => {
    md += `### ${comp.componentName || `Component ${i}`}\n`;
    md += `Path: \`${comp.nodePath}\`\n`;
    md += "```typescript\n";
    md += comp.generatedCode?.substring(0, 200) + "...\n";
    md += "```\n\n";
  });

  return md;
}
```


## Design Token Management

### Token Extraction & Export

```typescript
// Extract tokens from Figma and export to multiple formats
async function exportTokens(figmaFileId: string, outputDir: string) {
  const variables = await mcp__figma__get_variable_defs({
    fileId: figmaFileId
  });

  // Format 1: CSS Custom Properties
  let cssContent = ":root {\n";
  variables.forEach(token => {
    cssContent += `  --${token.name}: ${token.value};\n`;
  });
  cssContent += "}\n";
  fs.writeFileSync(path.join(outputDir, "tokens.css"), cssContent);

  // Format 2: JSON (for JavaScript)
  const jsonTokens = Object.fromEntries(
    variables.map(token => [token.name, token.value])
  );
  fs.writeFileSync(
    path.join(outputDir, "tokens.json"),
    JSON.stringify(jsonTokens, null, 2)
  );

  // Format 3: SCSS Variables
  let scssContent = "";
  variables.forEach(token => {
    scssContent += `$${token.name}: ${token.value};\n`;
  });
  fs.writeFileSync(path.join(outputDir, "tokens.scss"), scssContent);

  console.log(`✓ Exported ${variables.length} tokens in 3 formats`);
}
```


