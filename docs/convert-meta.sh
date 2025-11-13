#!/bin/bash

# Convert all _meta.json files to _meta.tsx format for Nextra v3.3+

find /Users/goos/MoAI/MoAI-ADK/docs/pages/ko -name "_meta.json" | while read json_file; do
  tsx_file="${json_file%.json}.tsx"

  echo "Converting: $json_file -> $tsx_file"

  # Read JSON content and convert to TypeScript export
  echo "export default $(cat "$json_file");" > "$tsx_file"

  # Remove the old JSON file
  rm "$json_file"

  echo "✓ Converted and removed: $json_file"
done

echo ""
echo "Conversion complete!"
echo "Total files converted: $(find /Users/goos/MoAI/MoAI-ADK/docs/pages/ko -name "_meta.tsx" | wc -l)"
