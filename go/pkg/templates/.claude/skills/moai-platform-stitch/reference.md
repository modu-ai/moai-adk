# Google Stitch Prompt Template Library

Comprehensive collection of production-ready prompt templates for Google Stitch MCP organized by category, complexity level, and use case.

---

## Table of Contents

- Authentication Screens
- Data Screens
- Navigation Patterns
- Form Screens
- Modals and Overlays
- E-commerce Patterns
- Analytics Dashboards
- Fitness and Health Apps
- Finance and Banking Apps
- Music and Media Apps
- Design System Integration
- Accessibility Patterns (WCAG 2.2)
- Progressive Complexity Examples

---

## Authentication Screens

### Basic Login Screen

```
Login screen with centered card layout (400px max-width). Email input with email icon prefix, password input with show/hide eye icon, login button (primary), forgot password link (below button), divider with "OR" text, social login buttons (Google, GitHub). Use white background, subtle shadow, Inter font family. Mobile: full-width card, stack vertically.
```

**Complexity**: Basic
**Components**: 8
**Responsive**: Yes

### Registration with Multi-Step Form

```
Multi-step registration wizard with progress indicator at top showing 4 steps (Account, Profile, Preferences, Complete). Step 1: Email, password, confirm password. Step 2: Full name, phone number, profile picture upload. Step 3: Interest checkboxes, newsletter subscription toggle. Step 4: Success message with "Get Started" button. Use Next/Back navigation at bottom. Center in 600px card. Highlight current step in progress bar with accent color.
```

**Complexity**: Advanced
**Components**: 15+
**Responsive**: Yes
**Interactive**: Yes (progress tracking)

### Forgot Password Screen

```
Forgot password screen with centered card layout. Page title "Reset Password", subtitle "Enter your email address and we'll send you a link to reset your password". Email input field with "Send Reset Link" primary button. Secondary link "Remember your password? Sign in" below button. Use minimal design with plenty of whitespace. Mobile: full-width card.
```

**Complexity**: Basic
**Components**: 5
**Responsive**: Yes

### Social Login Options

```
Authentication screen with social login focus. Large "Continue with Google" button with Google logo, "Continue with Apple" button with Apple logo, "Continue with GitHub" button with GitHub logo. Divider line with text "or continue with email". Email input, password input, "Sign In" button. Footer text "Don't have an account? Sign up". Stack all buttons vertically with 12px spacing. Use brand colors for each social button.
```

**Complexity**: Intermediate
**Components**: 9
**Responsive**: Yes

### Two-Factor Authentication

```
Two-factor authentication screen with centered layout. Title "Two-Factor Authentication", subtitle "Enter the 6-digit code from your authenticator app". Six individual input boxes (one digit each) with auto-focus next input. Timer showing "Resend code in 30 seconds" link. "Verify" primary button, "Back to login" secondary link. Use large touch targets (48px height). Mobile: stack vertically, center align.
```

**Complexity**: Intermediate
**Components**: 10
**Interactive**: Yes (auto-focus inputs)

---

## Data Screens

### Data Table with Sorting and Filtering

```
Data table with sticky header showing columns: ID (number), Name (text with avatar), Email (text), Status (badge: Active/Inactive), Role (dropdown badge), Created Date (date), Actions (kebab menu). Include sort indicators on sortable columns (ID, Name, Created Date). Filter bar above table with search input, status dropdown filter, role multi-select filter. Row hover effect with background color change. Pagination controls at bottom right (Previous, Page numbers, Next, Rows per page). Use compact spacing (32px row height). Mobile: hide columns, show card view.
```

**Complexity**: Advanced
**Components**: 20+
**Interactive**: Yes (sorting, filtering, pagination)
**Responsive**: Yes (table to card transformation)

### Analytics Dashboard with Metrics

```
Analytics dashboard with sidebar navigation (Dashboard, Reports, Settings). Top header with page title "Analytics Overview" and date range picker dropdown. Main content area with 4 metric cards in row: Total Revenue (large number, green upward trend icon), Active Users (large number, blue user icon), Conversion Rate (percentage, orange chart icon), Avg Session Duration (time, purple clock icon). Below metrics, line chart showing revenue trend over 30 days. Below chart, data table with top 5 performing pages. Use card-based layout with 8px gaps. Mobile: hide sidebar, stack metric cards in 2 columns.
```

**Complexity**: Advanced
**Components**: 25+
**Data Visualization**: Yes (charts, metrics)
**Responsive**: Yes

### List with Cards Layout

```
User management list with card layout. Top bar with search input, "Add User" primary button, filter dropdown. Grid of user cards (3 columns on desktop, 1 on mobile). Each card: avatar image (left), name (bold), email (gray, small), role badge (top right), status indicator (green dot), "View Profile" link button. Card hover effect with shadow increase and slight lift. Load more button at bottom when scrolling. Use consistent card height (200px). Mobile: single column, full-width cards.
```

**Complexity**: Intermediate
**Components**: 15
**Responsive**: Yes (grid reflow)
**Interactive**: Yes (search, filter, load more)

### Charts and Data Visualization

```
Data visualization dashboard with multiple chart types. Top row: 3 donut charts showing user distribution by Device Type, Location, Browser. Middle row: 2 bar charts comparing weekly revenue vs expenses. Bottom row: 1 large line chart showing user growth over 12 months. Each chart has title, legend, download button (export as PNG). Tooltip on hover showing exact values. Use consistent color palette across all charts (primary blue, secondary purple, accent green). Sidebar navigation. Mobile: stack charts vertically, hide sidebar.
```

**Complexity**: Advanced
**Components**: 20+
**Data Visualization**: Yes (donut, bar, line charts)
**Interactive**: Yes (tooltips, export)

### Kanban Board Layout

```
Kanban board with 4 columns: Backlog, In Progress, Review, Done. Each column has title with task count in badge, "Add Task" button. Task cards show title, priority indicator (high/medium/low), assignee avatar, due date, labels. Drag and drop functionality between columns. Horizontal scroll for columns on mobile. Filter bar at top with assignee dropdown, priority dropdown, search input. Use gray background for board, white cards with shadow. Mobile: single column view with column toggle tabs.
```

**Complexity**: Advanced
**Components**: 18
**Interactive**: Yes (drag and drop)
**Responsive**: Yes (horizontal scroll, column toggle)

---

## Navigation Patterns

### Sidebar Navigation

```
Admin dashboard with left sidebar navigation. Sidebar has logo at top, navigation links (Dashboard, Users, Reports, Settings) with icons, active state highlight (background color + left border), collapsed toggle button at bottom. User profile section at bottom with avatar, name, logout button. Main content area with page title and breadcrumb navigation. Sidebar 240px wide on desktop, collapses to icons only on tablet, hidden on mobile with hamburger menu. Use dark sidebar (#1a1a1a) with white text.
```

**Complexity**: Intermediate
**Components**: 12
**Responsive**: Yes (collapse, hide)
**Navigation Pattern**: Sidebar

### Top Navigation Bar

```
Website with top navigation bar. Logo on left, navigation links in center (Home, Features, Pricing, About, Contact), "Get Started" primary button on right. Sticky header with shadow on scroll. Mobile: hamburger menu that opens full-screen overlay with navigation links stacked vertically. Active link underline with accent color. Hover effect on links with color change. Use white background, 64px height. Secondary nav below main for sub-pages.
```

**Complexity**: Intermediate
**Components**: 10
**Responsive**: Yes (hamburger overlay)
**Navigation Pattern**: Top bar

### Bottom Navigation (Mobile)

```
Mobile app with bottom tab navigation. 5 tabs: Home, Search, Create, Notifications, Profile. Each tab has icon, label text below icon. Active tab with accent color fill, inactive with gray. Floating action button for Create tab (elevated, larger). Main content area above navigation. Safe area padding for iOS devices. Tab switch with smooth slide animation. Use fixed position at bottom, 56px height.
```

**Complexity**: Intermediate
**Components**: 8
**Platform**: Mobile-first
**Navigation Pattern**: Bottom tabs

### Breadcrumb Trails

```
Documentation page with breadcrumb navigation at top. Breadcrumb format: Home > Library > Documentation > Current Page. Each breadcrumb is clickable link, current page is non-clickable text. Separator is ">" icon. Hover effect on links with underline. Mobile: truncate long breadcrumbs with ellipsis (...). Page title below breadcrumbs. Side navigation for sub-sections on right. Use gray text for breadcrumbs, small font size (14px).
```

**Complexity**: Basic
**Components**: 5
**Navigation Pattern**: Breadcrumbs
**Responsive**: Yes (truncate)

### Tab Systems

```
Settings page with tab navigation. Horizontal tabs at top: General, Security, Notifications, Billing, Integrations. Active tab with bottom border (3px, accent color), inactive tabs with gray text. Tab content area below with smooth fade-in animation when switching. Tab panels are pre-rendered but hidden. Save changes button at bottom right (sticky). Use 48px tab height, 16px padding. Mobile: horizontal scroll for tabs if overflow.
```

**Complexity**: Intermediate
**Components**: 10
**Interactive**: Yes (tab switching, animations)
**Navigation Pattern**: Tabs

---

## Form Screens

### Single Column Form

```
Contact form with single column layout. Page title "Contact Us", subtitle "We'd love to hear from you". Form fields: Name (text input), Email (email input), Subject (dropdown: General, Support, Sales), Message (textarea with 4 rows), checkbox "I agree to the privacy policy". Submit button at bottom (full-width). Field labels above inputs, required field indicator (asterisk). Validation errors in red below fields. Use 400px max-width card, center on page. Mobile: full-width card.
```

**Complexity**: Basic
**Components**: 8
**Form Pattern**: Single column
**Responsive**: Yes

### Multi-Column Form

```
User registration form with 2-column grid layout. Left column: First Name (text), Last Name (text), Email (email), Phone (tel). Right column: Address (text), City (text), State (dropdown), Zip Code (text). Full-width below: Bio (textarea), Newsletter checkbox. Submit and Cancel buttons at bottom right. Label and input pairs in grid cells. Use 600px max-width, 16px gap between columns. Mobile: single column stack.
```

**Complexity**: Intermediate
**Components**: 12
**Form Pattern**: Multi-column grid
**Responsive**: Yes (2 columns to 1)

### Multi-Step Wizard

```
Onboarding wizard with 5 steps. Progress stepper at top showing all steps with numbers, current step highlighted with accent color, completed steps with checkmark. Step 1: Welcome screen with "Get Started" button. Step 2: Profile setup (name, avatar upload, bio). Step 3: Preferences (interest multi-select, language dropdown, timezone). Step 4: Connect accounts (social media toggle cards). Step 5: Complete with success animation, "Go to Dashboard" button. Back/Next navigation at bottom. Use 800px max-width card. Mobile: single column, full-width.
```

**Complexity**: Advanced
**Components**: 20+
**Form Pattern**: Wizard
**Interactive**: Yes (progress tracking, validation)

### Search and Filter Interface

```
Product search interface with filter sidebar and results grid. Left sidebar (280px): Category checkboxes (expandable sections), Price range slider, Brand multi-select, Rating stars filter, Color swatches. Right main content: Search bar at top with autocomplete, Sort dropdown (Relevance, Price, Rating), Clear all filters link. Product grid with 12 products per page. No results state with illustration. Active filters shown as removable tags. Use sticky filters on scroll. Mobile: filters in slide-out drawer.
```

**Complexity**: Advanced
**Components**: 25+
**Form Pattern**: Search + Filter sidebar
**Interactive**: Yes (autocomplete, dynamic filtering)

### Inline Edit Form

```
Data table with inline editing capabilities. Table shows user rows. Each row has: avatar, name (text), email (text), role (badge), status (badge), actions (Edit button). Clicking Edit transforms row to edit mode: name becomes input field, email becomes input field, role becomes dropdown, status becomes toggle, Edit button changes to Save/Cancel buttons. Validation on save with error messages. Save updates row with success flash animation. Use compact spacing (40px row height). Mobile: stack fields in edit mode.
```

**Complexity**: Advanced
**Components**: 15
**Form Pattern**: Inline editing
**Interactive**: Yes (edit mode toggle, validation)

---

## Modals and Overlays

### Dialog Modal

```
Confirmation dialog modal with overlay backdrop (semi-transparent black). Modal card centered on screen (500px max-width). Title "Delete Account", warning text "Are you sure you want to delete your account? This action cannot be undone.", warning icon. "Cancel" secondary button (left), "Delete Account" danger button (right, red background). Close X button at top right. Fade-in animation from scale 0.9 to 1.0. Click outside closes modal. ESC key closes modal. Use white card, 16px padding, 24px border-radius.
```

**Complexity**: Intermediate
**Components**: 6
**Modal Type**: Dialog
**Interactive**: Yes (close on click outside, ESC)

### Slide-Out Drawer

```
Settings drawer sliding from right side. Trigger button in header opens drawer. Drawer 400px wide, full viewport height, overlay backdrop. Drawer sections: Account (avatar, name, email), Notifications (toggle switches), Privacy (radio buttons), Appearance (color picker, theme dropdown), Danger Zone (red delete button). Close button at top right. Slide-in animation from right (300ms ease-out). Click outside closes drawer. Use white background, left border (1px gray). Mobile: full-width drawer.
```

**Complexity**: Intermediate
**Components**: 15
**Modal Type**: Drawer (slide-out)
**Interactive**: Yes (slide animation, close on overlay)

### Dropdown Menu

```
User avatar dropdown menu in header. Clicking avatar opens dropdown menu below. Menu items: Profile (with user icon), Settings (gear icon), Billing (credit card icon), Divider line, Help (question icon), Logout (sign out icon). Each item has icon on left, text, keyboard shortcut hint on right. Hover effect on items with background color. Active state tracking. Click outside closes dropdown. Fade-in animation (150ms). Use white card, shadow, 8px border-radius, 200px min-width.
```

**Complexity**: Intermediate
**Components**: 8
**Modal Type**: Dropdown menu
**Interactive**: Yes (hover, click outside close)

### Toast Notification

```
Toast notification appearing at top-right of screen. Four types: Success (green checkmark icon), Error (red X icon), Warning (yellow exclamation icon), Info (blue info icon). Message text on right of icon. Close X button on right. Auto-dismiss after 5 seconds with progress bar at bottom. Hover pauses auto-dismiss. Stacking for multiple toasts (vertical). Slide-in animation from right (300ms bounce). Use white card with colored left border (4px) indicating type. Mobile: full-width toasts at top.
```

**Complexity**: Intermediate
**Components**: 5 per toast
**Modal Type**: Toast notification
**Interactive**: Yes (auto-dismiss, pause on hover, stacking)

### Tooltip

```
Tooltip appearing on hover over elements. Trigger element: "Info" icon with question mark. Tooltip shows above element with arrow pointing down. Tooltip content: "This field requires a valid email address". Use dark background (#333), white text, 8px border-radius, 200px max-width. Fade-in animation (150ms delay before showing). Disappear on mouse leave. Keyboard accessible (focus shows tooltip, ESC closes). Use 4px offset from trigger element.
```

**Complexity**: Basic
**Components**: 2
**Modal Type**: Tooltip
**Interactive**: Yes (hover, keyboard accessible)

---

## E-commerce Patterns

### Product Card Component

```
E-commerce product card with image at top (aspect ratio 1:1), product title below (2 lines max, truncate with ellipsis), price in accent color (large font), original price crossed out in gray (if on sale), "Add to Cart" primary button (full-width), rating stars (yellow), review count in gray. Card hover effect: shadow increase, slight lift (translate Y -4px), image zoom. Use white card, 8px border-radius. Grid layout with 4 columns on desktop, 2 on tablet, 1 on mobile.
```

**Complexity**: Intermediate
**Components**: 10
**Domain**: E-commerce
**Interactive**: Yes (hover effects)

### Checkout Flow

```
Multi-step checkout with 3 steps. Step 1 Shipping: shipping form (name, address, city, state, zip, country), saved address cards (selectable), shipping method radio cards (Standard $5, Express $15, Overnight $25) with delivery estimates. Step 2 Payment: credit card form (number, expiry, CVV, visual card preview), saved payment methods, PayPal button. Step 3 Review: order summary (itemized list with images, quantities, prices), shipping address review, payment method review, "Place Order" button. Progress stepper at top. Use 800px max-width, gray background.
```

**Complexity**: Advanced
**Components**: 30+
**Domain**: E-commerce
**Interactive**: Yes (multi-step, saved selections)

### Shopping Cart

```
Shopping cart page with 2-column layout. Left column (70%): cart items list. Each item row: product thumbnail (left), product details (title, variant options), quantity stepper (- 1 +), remove button (X icon), price (right). Right column (30%): order summary card (subtotal, shipping cost, tax, total), promo code input field, "Proceed to Checkout" button. Empty cart state with illustration and "Continue Shopping" button. Use sticky order summary on scroll. Mobile: stack columns, order summary at bottom.
```

**Complexity**: Advanced
**Components**: 20+
**Domain**: E-commerce
**Interactive**: Yes (quantity stepper, remove, promo code)

### Product Detail Page

```
Product detail page with 2-column layout. Left column: product image gallery (main image with zoom, thumbnail strip below, click to change main image). Right column: product title (h1), rating stars with review count link, price (large), sale badge (if applicable), variant selectors (color swatches, size buttons), quantity stepper, "Add to Cart" primary button, "Buy Now" secondary button, accordion below (Description, Specifications, Reviews). Use sticky add to cart bar on mobile. Mobile: stack columns, image gallery first.
```

**Complexity**: Advanced
**Components**: 25+
**Domain**: E-commerce
**Interactive**: Yes (image zoom, variant selection, accordion)

---

## Analytics Dashboards

### Metrics Overview Dashboard

```
Analytics dashboard with metric cards at top. Row of 4 cards: Total Users (number with +12% growth indicator), Active Sessions (number with live pulse animation), Conversion Rate (percentage with mini sparkline chart), Revenue (currency with +8% growth). Each card has icon (colored background), trend indicator (up/down arrow with color), subtitle text. Below metrics, main content area with 2 columns: left column has line chart (user growth over time), right column has donut chart (user distribution by device). Use card-based layout, 8px gaps, consistent color palette.
```

**Complexity**: Advanced
**Components**: 20+
**Domain**: Analytics
**Data Visualization**: Yes (sparklines, line chart, donut chart)

### Real-Time Monitoring Dashboard

```
Real-time monitoring dashboard with live data. Top bar: auto-refresh toggle (5s, 30s, 1m), last updated timestamp. Main grid: 6 gauge charts showing CPU, Memory, Disk, Network, Database, Cache usage. Each gauge has percentage value, color-coded zones (green < 50%, yellow 50-80%, red > 80%). Log stream at bottom with auto-scroll, colored by severity (INFO blue, WARN yellow, ERROR red). Alert panel on right with active alerts (timestamp, severity, message, acknowledge button). Use dark theme (#0f172a), card layout.
```

**Complexity**: Advanced
**Components**: 25+
**Domain**: Analytics (monitoring)
**Real-Time**: Yes (auto-refresh, live gauges, log stream)

### Funnel Analysis Dashboard

```
Conversion funnel dashboard showing user journey steps. Funnel visualization: 5 horizontal bars (Page View, Sign Up, Email Verify, Profile Complete, First Purchase). Each bar shows step name, user count, percentage, drop-off percentage from previous step. Connecting arrows between bars. Below funnel, breakdown cards: Traffic sources (pie chart), Device types (donut chart), Geographic distribution (map with heat overlay). Date range selector at top right. Export buttons (CSV, PDF). Use card layout, consistent colors across funnel stages.
```

**Complexity**: Advanced
**Components**: 15
**Domain**: Analytics (funnel analysis)
**Data Visualization**: Yes (funnel, pie chart, map)

---

## Fitness and Health Apps

### Workout Tracking Screen

```
Workout tracking screen with exercise list. Top section: workout summary (total time, calories burned, exercises completed). Main list: exercise cards with exercise name, sets x reps, weight input field, checkmark completion button. Each card has expandable section for notes and rest timer. Bottom bar: "Start Workout" primary button (sticky). Empty state with "No exercises yet" illustration and "Add Exercise" button. Use card layout, green accent color for fitness theme. Mobile: full-width cards, sticky bottom bar.
```

**Complexity**: Intermediate
**Components**: 15
**Domain**: Fitness
**Interactive**: Yes (completion toggle, expand, timer)

### Meal Logging Interface

```
Meal logging interface with daily breakdown. Date selector at top (previous day, today date, next day). Meal sections: Breakfast, Lunch, Dinner, Snacks. Each section has "Add Food" button, list of food items (name, serving size, calories). Nutrition summary card at top: total calories, protein, carbs, fat with progress bars towards daily goals. Food search modal with autocomplete, recent items, favorites. Use fresh color palette (green, orange). Mobile: single column, sticky summary card.
```

**Complexity**: Intermediate
**Components**: 18
**Domain**: Health (nutrition)
**Interactive**: Yes (date navigation, food search, add items)

### Progress Tracker

```
Fitness progress tracker with charts and metrics. Top section: current weight with goal weight, progress percentage, remaining to goal. Main chart: line graph showing weight over time with goal line. Secondary metrics: body measurements table (date, weight, body fat, muscle mass), workout frequency bar chart (weekly), before/after photo gallery (2 columns). Action buttons: "Log Weight", "Upload Photo", "View History". Use motivating color scheme (blue progress, green success). Mobile: stack charts vertically, photo gallery 1 column.
```

**Complexity**: Advanced
**Components**: 20+
**Domain**: Fitness (progress tracking)
**Data Visualization**: Yes (line chart, bar chart, photo gallery)

---

## Finance and Banking Apps

### Account Dashboard

```
Banking dashboard with account overview. Top section: total balance (large, masked), account type badge, "Show Balance" toggle button. Main cards: checking account (balance, account number last 4), savings account (balance, interest rate), credit card (balance, available credit, payment due date). Transaction list below: date, merchant icon, merchant name, category badge, amount (green for income, red for expense). Filter bar: all, income, expenses tabs. Search input. Use trustworthy color palette (blue, green). Mobile: hide credit card details behind tap-to-reveal.
```

**Complexity**: Intermediate
**Components**: 18
**Domain**: Finance (banking)
**Interactive**: Yes (balance toggle, filters, search)

### Transaction History with Filters

```
Transaction history page with advanced filtering. Top bar: search input (search by merchant or amount), date range picker. Filter sidebar (left): category checkboxes (Food, Transport, Shopping, etc.), amount range slider, account type multi-select. Main content: transaction list grouped by date (Today, Yesterday, This Week, etc.). Each transaction: merchant logo, name, category, time, amount with color coding, swipe to reveal actions (split, categorize, receipt). Export button (CSV, PDF). Use card layout, sticky filters on scroll. Mobile: filters in slide-out drawer.
```

**Complexity**: Advanced
**Components**: 25+
**Domain**: Finance (transactions)
**Interactive**: Yes (filters, search, swipe actions)

### Budget Management

```
Budget management screen with category budgets. Top section: monthly budget summary (total budget, spent, remaining, percentage). Budget categories list: each category card shows icon, name, budget amount, spent amount, progress bar (color-coded: green < 70%, yellow 70-90%, red > 90%), "Edit" button. Click category opens detail modal: transaction breakdown, spending trend chart, adjust budget slider. Add category button at bottom. Use progress bars prominently, alert colors for over-budget. Mobile: full-width category cards.
```

**Complexity**: Intermediate
**Components**: 15
**Domain**: Finance (budgeting)
**Interactive**: Yes (edit, adjust slider, detail modal)

---

## Music and Media Apps

### Music Player Interface

```
Music player screen with album art (large square, rounded corners). Track info below: song title (bold), artist name (gray). Playback controls: shuffle button (toggle), previous track button, play/pause button (large, circular), next track button, repeat button (toggle). Progress bar: scrubber with current time / total duration labels. Volume slider with speaker icon. Lyrics button (shows lyrics sheet). Add to favorites heart button (toggle). Use blurred album art as background, gradient overlay. Mobile: full-screen player with swipe down to dismiss.
```

**Complexity**: Intermediate
**Components**: 12
**Domain**: Music (player)
**Interactive**: Yes (playback controls, progress scrub, volume)

### Playlist Management

```
Playlist management screen with playlist list. Top bar: "Create Playlist" primary button, search input. Playlist cards: cover image (3 collage thumbnails), playlist name, owner name, song count, duration. Grid layout (3 columns desktop, 2 tablet, 1 mobile). Hover effect: play button overlay on cover. Click card opens playlist detail: song list with drag-to-reorder, context menu (remove from playlist, add to queue), shuffle play button at top. Use dark theme with album art gradients. Mobile: single column list, swipe for actions.
```

**Complexity**: Intermediate
**Components**: 15
**Domain**: Music (playlists)
**Interactive**: Yes (create, search, drag-reorder, context menu)

### Video Streaming Interface

```
Video streaming interface with video player. Main area: video player with controls (play/pause, progress bar, volume, fullscreen, settings, captions). Video info below: thumbnail, title, channel avatar, channel name, subscriber count, views, upload date. Action buttons: like (thumb up), dislike (thumb down), share, save, more. Recommended videos sidebar: video cards (thumbnail, title, channel name, views, duration). Comments section below: top comment, sort by (top, newest), comment input. Use dark theme, red accent color. Mobile: hide sidebar, stack below video.
```

**Complexity**: Advanced
**Components**: 25+
**Domain**: Media (video streaming)
**Interactive**: Yes (player controls, like, share, comments)

---

## Design System Integration

### Design Token Reference Syntax

When extracting design context from existing screens, Stitch returns structured design tokens that can be referenced in prompts:

**Color Tokens Reference:**
```
"Use primary brand color for main button, secondary color for secondary button, accent color for highlights. Use neutral-500 for borders, neutral-800 for text. Use success color for confirmation states, warning color for alerts."
```

**Typography Tokens Reference:**
```
"Use heading font family for page titles (h1, h2), body font family for content. Use font-size-large for headlines, font-size-medium for subheadings, font-size-small for captions. Use font-weight-bold for emphasis, font-weight-regular for body text."
```

**Spacing Tokens Reference:**
```
"Use spacing-sm for tight layouts (4px), spacing-md for normal spacing (8px), spacing-lg for spacious layouts (16px). Use component-padding-md for card interiors, component-padding-lg for sections."
```

### Token Usage in Prompts

**Example with Token References:**
```
Login screen using extracted design context: primary-blue color for submit button, neutral-gray-100 for card background, font-family-Inter for all text, spacing-md (8px) between form fields. Use component-border-radius for card corners. Text color neutral-gray-800 for headings, neutral-gray-600 for body text.
```

### Multi-Screen Consistency

**Consistency Strategy:**
```
Extract design context from home screen first, then pass to all subsequent screen generations. This ensures color palette, typography scale, and spacing system remain consistent across entire application.
```

**Example Workflow:**
1. Generate home screen
2. Extract design context from home screen
3. Generate all other screens with design_context parameter

### Brand Guideline Adherence

**Brand-Specific Prompts:**
```
"Use brand color #FF5722 for primary actions, brand color #00BCD4 for secondary actions. Use brand font 'Roboto' for all text. Maintain brand spacing system (4px base unit, multiples of 4). Follow brand component styles: rounded buttons with 4px radius, cards with 8px radius, subtle 2px shadows."
```

---

## Accessibility Patterns (WCAG 2.2)

### ARIA Labels for Icon Buttons

**Pattern:**
```
"All icon buttons must have aria-label. Search button: 'Search products', Close button: 'Close dialog', Menu button: 'Open navigation menu', Settings button: 'Open settings'."
```

**Example Prompt:**
```
Header with search icon button (aria-label 'Search search'), notification bell icon (aria-label 'View notifications'), user avatar (aria-label 'User menu'). Ensure all interactive elements have accessible labels.
```

### Keyboard Navigation Patterns

**Pattern:**
```
"All interactive elements must be keyboard accessible. Tab order follows visual flow (left to right, top to bottom). Focus indicators visible (2px solid outline, offset 2px). Enter and Space activate buttons and links. Escape closes modals and dropdowns. Arrow keys navigate within components (list boxes, tabs)."
```

**Example Prompt:**
```
Navigation menu with keyboard support. Tab moves between menu items. Enter activates link. Arrow keys open submenus. Escape closes open menu. Focus visible with blue outline. Skip to main content link at top of page.
```

### Focus Indicators

**Pattern:**
```
"Custom focus styles for all interactive elements. Focus ring: 2px solid, offset 2px from element, color matches brand (usually blue). Ensure focus indicator has 3:1 contrast ratio against background. Remove default browser outline only if replacing with custom style."
```

**Example Prompt:**
```
Form inputs with custom focus styles. Default state: gray border. Focus state: blue border (2px), blue shadow (0 0 0 3px rgba(59, 130, 246, 0.1)). Error state: red border with red focus ring. Ensure keyboard users see focus indicator clearly.
```

### Color Contrast Requirements

**Pattern:**
```
"All text must meet WCAG AA contrast ratios. Normal text: 4.5:1 contrast against background. Large text (18pt+): 3:1 contrast. UI components: 3:1 contrast against adjacent colors. Interactive elements must have distinct focus and hover states with sufficient contrast."
```

**Example Prompt:**
```
Form with high contrast labels. Labels: dark gray (#1f2937) on white background (contrast ratio 15:1, exceeds AA). Required field asterisk: red (#dc2626) for visibility. Error messages: red (#dc2626) on white background. Link text: blue (#2563eb) with underline on hover.
```

### Screen Reader Optimization

**Pattern:**
```
"Semantic HTML structure. Use heading levels h1-h6 in correct order (h1 for page title, h2 for section titles). Use landmark elements: header, nav, main, section, article, aside, footer. Use lists for navigation (ul > li > a). Form labels associated with inputs (for attribute or aria-labelledby)."
```

**Example Prompt:**
```
Blog post page with semantic structure. h1: article title. Article element for main content. Section elements for introduction, body, conclusion. h2 for section headings. Aside for sidebar (related posts). Footer for page footer. Navigation in nav element with ul/li/a structure.
```

---

## Progressive Complexity Examples

### Level 1: Basic Screen (Simple)

**Prompt:**
```
Login form with email input, password input, login button. Center on page. Use white background, blue button.
```

**Components**: 3
**Complexity**: Beginner
**Use Case**: Learning Stitch basics

### Level 2: Standard Screen (Intermediate)

**Prompt:**
```
User registration form with centered card layout (500px). Email input, password input with show/hide toggle, confirm password input, full name input. Register button (primary), sign in link (secondary). Page title "Create Account", subtitle "Join us today". Validation errors below fields. Use white card, shadow, Inter font.
```

**Components**: 8
**Complexity**: Intermediate
**Use Case**: Standard registration flow

### Level 3: Complex Screen (Advanced)

**Prompt:**
```
Multi-step registration wizard with progress indicator (3 steps: Account, Profile, Complete). Step 1: Email, password, confirm password, show/hide toggle. Step 2: Full name, phone input with country code dropdown, profile picture upload (drag and drop zone). Step 3: Success message with checkmark animation, "Go to Dashboard" button. Next/Back navigation buttons. Current step highlighted in progress bar. Center card (600px max-width). Use extracted design context. Mobile: single column, full-width card.
```

**Components**: 18
**Complexity**: Advanced
**Use Case**: Complex onboarding flow

### Level 4: Expert Screen (Expert)

**Prompt:**
```
Admin dashboard with sidebar navigation (logo, nav links with icons, user profile at bottom). Main content: header with page title "User Management", breadcrumb, search input, filter dropdown, "Add User" primary button. Data table with columns: ID, Name (avatar + text), Email, Role (badge), Status (colored dot + text), Actions (kebab menu). Sort indicators on headers. Row hover effect. Pagination at bottom. Sticky table header. Use design context from existing screens. Mobile: hide sidebar (hamburger menu), table transforms to card layout. Accessible: ARIA labels on all icon buttons, keyboard navigation for table.
```

**Components**: 30+
**Complexity**: Expert
**Use Case**: Production-ready admin interface

---

## Prompt Optimization Techniques

### Incremental Refinement

**Technique:** Start with basic layout, then add complexity in follow-up prompts.

**Iteration 1 (Foundation):**
```
Product listing page with 3-column grid of product cards.
```

**Iteration 2 (Add Components):**
```
Add product details to each card: image, title, price, add to cart button.
```

**Iteration 3 (Add Interactions):**
```
Add hover effects on cards (shadow increase, lift), focus ring on button.
```

**Iteration 4 (Add Layout):**
```
Add filter sidebar on left with category links, sort dropdown at top.
```

**Iteration 5 (Refine):**
```
Mobile: single column layout, hide sidebar. Desktop: 3-column grid with sidebar.
```

### Component Breakdown

**Technique:** Generate individual components separately, then combine.

**Step 1: Generate Navigation**
```
Top navigation bar with logo, nav links (Home, Products, About), contact button.
```

**Step 2: Generate Hero Section**
```
Hero section with headline text, subheadline, primary CTA button, background image.
```

**Step 3: Generate Features Grid**
```
Features section with 3 cards in row, each card has icon, title, description.
```

**Step 4: Combine in Full Page**
```
Landing page with navigation bar, hero section below, features section below hero. Sticky header. Mobile: stack sections vertically.
```

---

## Quick Prompt Templates

### Copy-Paste Ready Templates

**Basic Form:**
```
[Form name] form with [field list]. Center in [width]px card. [Button text] button. Use [color scheme].
```

**Data Table:**
```
Data table with columns: [column list]. [Feature list: sort, filter, pagination]. Sticky header. Row hover effect. [Mobile behavior].
```

**Dashboard:**
```
[Dashboard type] dashboard with [number] metric cards at top. [Chart types] below. Sidebar navigation. [Mobile behavior].
```

**Modal:**
```
[Modal type] modal with [width]px card. [Title], [message text]. [Primary button text], [secondary button text]. Overlay backdrop. [Animation].
```

**Card Grid:**
```
[Card type] cards in [number]-column grid. Each card has [content list]. [Hover effects]. [Mobile: column count].
```

---

Status: Reference Complete
Version: 1.0.0
Last Updated: 2026-01-23
Total Templates: 40+ examples
