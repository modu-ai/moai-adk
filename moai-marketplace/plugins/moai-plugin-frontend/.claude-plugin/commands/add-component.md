# add-component
Generate new React component with TypeScript types and variants.

## Usage
```
/add-component [component-name] [--shadcn] [--test]
```

## Options
- `--shadcn`: Use shadcn/ui as base
- `--test`: Generate unit tests
- `--story`: Generate Storybook story

## What It Does
1. Creates component file with props interface
2. Adds TypeScript types
3. Implements component logic
4. Generates tests (optional)
5. Creates Storybook story (optional)

## Example
```bash
/add-component Button --shadcn --test
```
