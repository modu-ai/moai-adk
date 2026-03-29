name: moai-library-shadcn-theming
description: shadcn/ui theme system and design token customization

## Theme System Implementation

### Advanced Theme Provider

```typescript
// Advanced theme system with CSS variables and dark mode
import { createContext, useContext, useEffect, useState } from "react";

type Theme = "dark" | "light" | "system";

interface ThemeProviderProps {
 children: React.ReactNode;
 defaultTheme?: Theme;
 storageKey?: string;
 attribute?: string;
 enableSystem?: boolean;
}

interface ThemeProviderState {
 theme: Theme;
 setTheme: (theme: Theme) => void;
}

const ThemeProviderContext = createContext<ThemeProviderState | undefined>(undefined);

export function ThemeProvider({
 children,
 defaultTheme = "system",
 storageKey = "ui-theme",
 attribute = "class",
 enableSystem = true,
 ...props
}: ThemeProviderProps) {
 const [theme, setTheme] = useState<Theme>(() => {
 if (typeof window !== "undefined") {
 return (localStorage.getItem(storageKey) as Theme) || defaultTheme;
 }
 return defaultTheme;
 });

 useEffect(() => {
 const root = window.document.documentElement;

 root.classList.remove("light", "dark");

 if (theme === "system" && enableSystem) {
 const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
 .matches
 ? "dark"
 : "light";

 root.classList.add(systemTheme);
 return;
 }

 root.classList.add(theme);
 }, [theme, enableSystem, attribute]);

 const value = {
 theme,
 setTheme: (theme: Theme) => {
 localStorage.setItem(storageKey, theme);
 setTheme(theme);
 },
 };

 return (
 <ThemeProviderContext.Provider {...props} value={value}>
 {children}
 </ThemeProviderContext.Provider>
 );
}

export const useTheme = () => {
 const context = useContext(ThemeProviderContext);

 if (context === undefined)
 throw new Error("useTheme must be used within a ThemeProvider");

 return context;
};
```

### Design Token Configuration

```typescript
// Theme configuration with design tokens
export const theme = {
 light: {
 background: "hsl(0 0% 100%)",
 foreground: "hsl(240 10% 3.9%)",
 card: "hsl(0 0% 100%)",
 cardForeground: "hsl(240 10% 3.9%)",
 popover: "hsl(0 0% 100%)",
 popoverForeground: "hsl(240 10% 3.9%)",
 primary: "hsl(240 9% 10%)",
 primaryForeground: "hsl(0 0% 98%)",
 secondary: "hsl(240 4.8% 95.9%)",
 secondaryForeground: "hsl(240 3.8% 46.1%)",
 muted: "hsl(240 4.8% 95.9%)",
 mutedForeground: "hsl(240 3.8% 46.1%)",
 accent: "hsl(240 4.8% 95.9%)",
 accentForeground: "hsl(240 5.9% 10%)",
 destructive: "hsl(0 72.22% 50.59%)",
 destructiveForeground: "hsl(0 0% 98%)",
 border: "hsl(240 5.9% 90%)",
 input: "hsl(240 5.9% 90%)",
 ring: "hsl(240 5.9% 10%)",
 },
 dark: {
 background: "hsl(240 10% 3.9%)",
 foreground: "hsl(0 0% 98%)",
 card: "hsl(240 10% 3.9%)",
 cardForeground: "hsl(0 0% 98%)",
 popover: "hsl(240 10% 3.9%)",
 popoverForeground: "hsl(0 0% 98%)",
 primary: "hsl(0 0% 98%)",
 primaryForeground: "hsl(240 9% 10%)",
 secondary: "hsl(240 3.7% 15.9%)",
 secondaryForeground: "hsl(0 0% 98%)",
 muted: "hsl(240 3.7% 15.9%)",
 mutedForeground: "hsl(240 5% 64.9%)",
 accent: "hsl(240 3.7% 15.9%)",
 accentForeground: "hsl(0 0% 98%)",
 destructive: "hsl(0 62.8% 30.6%)",
 destructiveForeground: "hsl(0 0% 98%)",
 border: "hsl(240 3.7% 15.9%)",
 input: "hsl(240 3.7% 15.9%)",
 ring: "hsl(240 4.9% 83.9%)",
 },
};

// Apply theme CSS variables
export function applyThemeCSS() {
 const root = document.documentElement;
 
 Object.entries(theme.light).forEach(([key, value]) => {
 root.style.setProperty(`--${key}`, value);
 });
}
```

### CSS Variables in globals.css

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
 :root {
 --background: 0 0% 100%;
 --foreground: 240 10% 3.9%;
 --card: 0 0% 100%;
 --card-foreground: 240 10% 3.9%;
 --popover: 0 0% 100%;
 --popover-foreground: 240 10% 3.9%;
 --primary: 240 9% 10%;
 --primary-foreground: 0 0% 98%;
 --secondary: 240 4.8% 95.9%;
 --secondary-foreground: 240 3.8% 46.1%;
 --muted: 240 4.8% 95.9%;
 --muted-foreground: 240 3.8% 46.1%;
 --accent: 240 4.8% 95.9%;
 --accent-foreground: 240 5.9% 10%;
 --destructive: 0 72.22% 50.59%;
 --destructive-foreground: 0 0% 98%;
 --border: 240 5.9% 90%;
 --input: 240 5.9% 90%;
 --ring: 240 5.9% 10%;
 --radius: 0.5rem;
 }

 .dark {
 --background: 240 10% 3.9%;
 --foreground: 0 0% 98%;
 --card: 240 10% 3.9%;
 --card-foreground: 0 0% 98%;
 --popover: 240 10% 3.9%;
 --popover-foreground: 0 0% 98%;
 --primary: 0 0% 98%;
 --primary-foreground: 240 9% 10%;
 --secondary: 240 3.7% 15.9%;
 --secondary-foreground: 0 0% 98%;
 --muted: 240 3.7% 15.9%;
 --muted-foreground: 240 5% 64.9%;
 --accent: 240 3.7% 15.9%;
 --accent-foreground: 0 0% 98%;
 --destructive: 0 62.8% 30.6%;
 --destructive-foreground: 0 0% 98%;
 --border: 240 3.7% 15.9%;
 --input: 240 3.7% 15.9%;
 --ring: 240 4.9% 83.9%;
 }
}

@layer base {
 * {
 @apply border-border;
 }
 body {
 @apply bg-background text-foreground;
 }
}
```

### Custom Brand Theme

```typescript
// Create custom brand theme
export function createBrandTheme(brandColors: {
 primary: string;
 secondary: string;
 accent: string;
}) {
 return {
 light: {
 ...theme.light,
 primary: brandColors.primary,
 secondary: brandColors.secondary,
 accent: brandColors.accent,
 },
 dark: {
 ...theme.dark,
 primary: brandColors.primary,
 secondary: brandColors.secondary,
 accent: brandColors.accent,
 },
 };
}

// Apply custom brand theme
export function applyBrandTheme(brandTheme: typeof theme) {
 const root = document.documentElement;
 const currentTheme = root.classList.contains("dark") ? "dark" : "light";
 
 Object.entries(brandTheme[currentTheme]).forEach(([key, value]) => {
 root.style.setProperty(`--${key}`, value);
 });
}
```

### Theme Toggle Component

```typescript
// Theme toggle component
import { Moon, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
 DropdownMenu,
 DropdownMenuContent,
 DropdownMenuItem,
 DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useTheme } from "@/components/theme-provider";

export function ThemeToggle() {
 const { setTheme } = useTheme();

 return (
 <DropdownMenu>
 <DropdownMenuTrigger asChild>
 <Button variant="outline" size="icon">
 <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
 <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
 <span className="sr-only">Toggle theme</span>
 </Button>
 </DropdownMenuTrigger>
 <DropdownMenuContent align="end">
 <DropdownMenuItem onClick={() => setTheme("light")}>
 Light
 </DropdownMenuItem>
 <DropdownMenuItem onClick={() => setTheme("dark")}>
 Dark
 </DropdownMenuItem>
 <DropdownMenuItem onClick={() => setTheme("system")}>
 System
 </DropdownMenuItem>
 </DropdownMenuContent>
 </DropdownMenu>
 );
}
```

### Tailwind Configuration

```javascript
// tailwind.config.js
/ @type {import('tailwindcss').Config} */
module.exports = {
 darkMode: ["class"],
 content: [
 './pages//*.{ts,tsx}',
 './components//*.{ts,tsx}',
 './app//*.{ts,tsx}',
 './src//*.{ts,tsx}',
 ],
 theme: {
 container: {
 center: true,
 padding: "2rem",
 screens: {
 "2xl": "1400px",
 },
 },
 extend: {
 colors: {
 border: "hsl(var(--border))",
 input: "hsl(var(--input))",
 ring: "hsl(var(--ring))",
 background: "hsl(var(--background))",
 foreground: "hsl(var(--foreground))",
 primary: {
 DEFAULT: "hsl(var(--primary))",
 foreground: "hsl(var(--primary-foreground))",
 },
 secondary: {
 DEFAULT: "hsl(var(--secondary))",
 foreground: "hsl(var(--secondary-foreground))",
 },
 destructive: {
 DEFAULT: "hsl(var(--destructive))",
 foreground: "hsl(var(--destructive-foreground))",
 },
 muted: {
 DEFAULT: "hsl(var(--muted))",
 foreground: "hsl(var(--muted-foreground))",
 },
 accent: {
 DEFAULT: "hsl(var(--accent))",
 foreground: "hsl(var(--accent-foreground))",
 },
 popover: {
 DEFAULT: "hsl(var(--popover))",
 foreground: "hsl(var(--popover-foreground))",
 },
 card: {
 DEFAULT: "hsl(var(--card))",
 foreground: "hsl(var(--card-foreground))",
 },
 },
 borderRadius: {
 lg: "var(--radius)",
 md: "calc(var(--radius) - 2px)",
 sm: "calc(var(--radius) - 4px)",
 },
 keyframes: {
 "accordion-down": {
 from: { height: 0 },
 to: { height: "var(--radix-accordion-content-height)" },
 },
 "accordion-up": {
 from: { height: "var(--radix-accordion-content-height)" },
 to: { height: 0 },
 },
 },
 animation: {
 "accordion-down": "accordion-down 0.2s ease-out",
 "accordion-up": "accordion-up 0.2s ease-out",
 },
 },
 },
 plugins: [require("tailwindcss-animate")],
}
```

---

## OKLCH Color System

shadcn/ui has adopted OKLCH as its color system, replacing HSL for improved perceptual uniformity and wider gamut support. OKLCH provides more consistent perceived lightness across different hues, making it easier to create harmonious color palettes.

### OKLCH Format

OKLCH values use three channels: Lightness (0-1), Chroma (0-0.4), and Hue (0-360). CSS variables in the OKLCH system follow the pattern: oklch(L C H) where L is lightness, C is chroma, and H is hue angle.

### Migration from HSL

Existing HSL-based themes continue to work. New projects initialized with shadcn/cli v4 default to OKLCH. The preset system generates OKLCH-based color schemes. For manual migration, convert HSL values to OKLCH equivalents in the CSS variable definitions.

### Color Variable Pattern

The CSS variable pattern remains the same structure (--background, --foreground, --primary, etc.) but values change from HSL to OKLCH format. This means component code requires no changes — only the CSS variable definitions in globals.css need updating.

## Tailwind CSS v4 Compatibility

shadcn/ui fully supports Tailwind CSS v4 with CSS-first configuration. Key differences from v3:

### Configuration

Tailwind v4 uses CSS-based configuration instead of tailwind.config.js. Theme customization happens directly in CSS using @theme directives. The shadcn/cli auto-detects the Tailwind version and generates appropriate configuration.

### CSS Variables

CSS variable integration is native in Tailwind v4. Color values defined as CSS variables work directly without the hsl() wrapper function. This simplifies the theming layer and aligns with the OKLCH color system.

### Dark Mode

Dark mode configuration uses the same class-based approach. The .dark selector in CSS defines dark theme variable overrides.

## Preset-Based Theming

The shadcn preset system bundles complete theme configurations into a single shareable code. A preset includes color palette, typography (font families), icon library, and border radius settings.

Presets are generated through the shadcn/create builder with live preview on real components. Apply a preset during project initialization or switch presets by re-initializing.

For custom presets, define the complete set of CSS variables, font imports, and configuration in a registry:base item. The build command packages it for distribution.

---

Last Updated: 2026-03-28
Related: [Main Skill](../SKILL.md), [Component Architecture](component-architecture.md)
