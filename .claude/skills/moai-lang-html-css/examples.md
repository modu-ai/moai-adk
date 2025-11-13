# Advanced Examples & Production Patterns

## Example 1: Complete Blog Layout with Semantic HTML

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Tech Blog - Articles & Insights</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <!-- Page Header -->
  <header role="banner">
    <nav role="navigation" aria-label="Main navigation">
      <a href="/" class="logo">TechBlog</a>
      <ul>
        <li><a href="/articles" aria-current="page">Articles</a></li>
        <li><a href="/about">About</a></li>
        <li><a href="/contact">Contact</a></li>
      </ul>
    </nav>
  </header>

  <!-- Skip Link for Accessibility -->
  <a href="#main-content" class="sr-only">Skip to main content</a>

  <!-- Main Content -->
  <main id="main-content">
    <!-- Featured Article -->
    <article class="featured-article">
      <header>
        <h1>Getting Started with Modern Web Development</h1>
        <p class="meta">
          By <strong>Sarah Johnson</strong> on
          <time datetime="2025-11-12">November 12, 2025</time>
        </p>
      </header>
      <img src="featured.jpg" alt="Modern web development tools">
      <section>
        <p>Learn the fundamentals of modern web development with practical examples...</p>
        <a href="#article-full">Read Full Article</a>
      </section>
    </article>

    <!-- Articles Grid -->
    <section aria-labelledby="articles-heading">
      <h2 id="articles-heading">Latest Articles</h2>
      <div class="articles-grid">
        <article class="article-card">
          <header>
            <h3>CSS Grid Mastery</h3>
            <time datetime="2025-11-10">Nov 10</time>
          </header>
          <p>Master CSS Grid layout techniques for modern responsive design...</p>
          <footer>
            <span class="category">CSS</span>
          </footer>
        </article>

        <article class="article-card">
          <header>
            <h3>React Performance Optimization</h3>
            <time datetime="2025-11-08">Nov 8</time>
          </header>
          <p>Optimize your React applications for better performance...</p>
          <footer>
            <span class="category">JavaScript</span>
          </footer>
        </article>
      </div>
    </section>

    <!-- Sidebar with Related Content -->
    <aside aria-label="Related content">
      <h2>Categories</h2>
      <nav>
        <ul>
          <li><a href="/css">CSS</a></li>
          <li><a href="/javascript">JavaScript</a></li>
          <li><a href="/accessibility">Accessibility</a></li>
        </ul>
      </nav>
    </aside>
  </main>

  <!-- Page Footer -->
  <footer role="contentinfo">
    <p>&copy; 2025 TechBlog. All rights reserved.</p>
  </footer>
</body>
</html>
```

```css
/* Responsive Blog Layout */
:root {
  --color-primary: #0ea5e9;
  --color-text: #1f2937;
  --spacing: 1rem;
}

body {
  font-family: system-ui, sans-serif;
  line-height: 1.6;
  color: var(--color-text);
}

/* Navigation */
header {
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: var(--spacing);
  position: sticky;
  top: 0;
}

nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

nav a {
  text-decoration: none;
  color: var(--color-text);
  transition: color 0.3s;
}

nav a:hover {
  color: var(--color-primary);
}

/* Main Layout */
main {
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--spacing);
  display: grid;
  grid-template-columns: 1fr 300px;
  gap: 2rem;
}

/* Featured Article */
.featured-article {
  grid-column: 1 / -1;
  margin-bottom: 2rem;
}

.featured-article h1 {
  font-size: 2.5rem;
  margin: 1rem 0;
}

.featured-article img {
  width: 100%;
  height: auto;
  border-radius: 8px;
}

/* Articles Grid */
.articles-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
  margin: 2rem 0;
}

.article-card {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 1.5rem;
  transition: box-shadow 0.3s;
}

.article-card:hover {
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.article-card h3 {
  margin: 0;
  font-size: 1.25rem;
}

/* Accessibility */
.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0,0,0,0);
  white-space: nowrap;
  border-width: 0;
}

/* Responsive */
@media (max-width: 768px) {
  main {
    grid-template-columns: 1fr;
  }

  .featured-article h1 {
    font-size: 1.75rem;
  }

  aside {
    order: -1;
  }
}
```

## Example 2: Accessible Form Component

```html
<form method="POST" action="/submit" class="contact-form">
  <fieldset>
    <legend>Contact Information</legend>

    <!-- Email Field -->
    <div class="form-group">
      <label for="email">Email Address <span aria-label="required">*</span></label>
      <input
        id="email"
        type="email"
        name="email"
        required
        aria-required="true"
        aria-describedby="email-hint"
      >
      <small id="email-hint">We'll never share your email</small>
    </div>

    <!-- Message Field -->
    <div class="form-group">
      <label for="message">Message <span aria-label="required">*</span></label>
      <textarea
        id="message"
        name="message"
        rows="6"
        required
        aria-required="true"
        aria-describedby="message-hint"
        placeholder="Your message here..."
      ></textarea>
      <small id="message-hint">Maximum 500 characters</small>
    </div>

    <!-- Checkboxes -->
    <fieldset class="checkbox-group">
      <legend>Interests</legend>
      <div class="form-check">
        <input
          type="checkbox"
          id="interest-web"
          name="interests"
          value="web"
        >
        <label for="interest-web">Web Development</label>
      </div>
      <div class="form-check">
        <input
          type="checkbox"
          id="interest-design"
          name="interests"
          value="design"
        >
        <label for="interest-design">Design</label>
      </div>
    </fieldset>

    <!-- Error Message Area -->
    <div role="alert" aria-live="polite" id="form-error" class="hidden">
      <!-- Validation errors appear here -->
    </div>

    <!-- Submit Button -->
    <button type="submit" class="btn btn-primary">
      Send Message
    </button>
  </fieldset>
</form>
```

```css
.contact-form {
  max-width: 500px;
}

.form-group {
  margin-bottom: 1.5rem;
}

label {
  display: block;
  font-weight: 500;
  margin-bottom: 0.5rem;
  color: #374151;
}

input,
textarea,
select {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-family: inherit;
  font-size: 1rem;
}

input:focus,
textarea:focus,
select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.1);
}

input[type="checkbox"] {
  width: auto;
  margin-right: 0.5rem;
}

.form-check {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;
}

[role="alert"] {
  color: #dc2626;
  margin-bottom: 1rem;
}

.hidden {
  display: none;
}
```

## Example 3: Responsive Navigation Menu

```html
<header>
  <nav class="navbar">
    <div class="navbar-container">
      <a href="/" class="navbar-brand">MyBrand</a>

      <!-- Toggle for Mobile Menu -->
      <button
        class="navbar-toggle"
        aria-label="Toggle navigation"
        id="navbar-toggle"
      >
        <span class="hamburger"></span>
      </button>

      <!-- Navigation List -->
      <ul class="navbar-menu" id="navbar-menu">
        <li><a href="/">Home</a></li>
        <li><a href="/services">Services</a></li>
        <li><a href="/portfolio">Portfolio</a></li>
        <li><a href="/contact">Contact</a></li>
      </ul>
    </div>
  </nav>
</header>
```

```css
.navbar {
  background: white;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.navbar-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.navbar-brand {
  font-size: 1.5rem;
  font-weight: bold;
  text-decoration: none;
  color: var(--color-primary);
}

.navbar-menu {
  display: flex;
  list-style: none;
  gap: 2rem;
  margin: 0;
  padding: 0;
}

.navbar-menu a {
  text-decoration: none;
  color: var(--color-text);
  transition: color 0.3s;
}

.navbar-menu a:hover,
.navbar-menu a:focus {
  color: var(--color-primary);
  outline: none;
}

/* Mobile Menu Toggle */
.navbar-toggle {
  display: none;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.5rem;
}

.hamburger {
  display: block;
  width: 25px;
  height: 20px;
  position: relative;
}

.hamburger::before,
.hamburger::after,
.hamburger {
  content: '';
  position: absolute;
  height: 3px;
  background: var(--color-text);
  left: 0;
  transition: all 0.3s;
}

.hamburger::before {
  top: 0;
  width: 100%;
}

.hamburger::after {
  bottom: 0;
  width: 100%;
}

.hamburger {
  top: 50%;
  width: 100%;
  transform: translateY(-50%);
}

/* Mobile Responsive */
@media (max-width: 768px) {
  .navbar-toggle {
    display: block;
  }

  .navbar-menu {
    display: none;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    flex-direction: column;
    background: white;
    gap: 0;
    padding: 1rem 0;
    box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  }

  .navbar-menu.active {
    display: flex;
  }

  .navbar-menu li {
    border-bottom: 1px solid #e5e7eb;
  }

  .navbar-menu a {
    display: block;
    padding: 1rem;
  }
}
```

## Example 4: Card Component System

```html
<div class="card">
  <div class="card-header">
    <h3 class="card-title">Card Title</h3>
  </div>
  <div class="card-body">
    <p>Card content goes here</p>
  </div>
  <div class="card-footer">
    <button class="btn btn-secondary">Learn More</button>
  </div>
</div>
```

```css
.card {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  transition: box-shadow 0.3s;
}

.card:hover {
  box-shadow: 0 4px 6px rgba(0,0,0,0.1);
}

.card-header {
  padding: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
  background: #f9fafb;
}

.card-title {
  margin: 0;
  font-size: 1.25rem;
  color: var(--color-text);
}

.card-body {
  padding: 1.5rem;
}

.card-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
  display: flex;
  gap: 0.5rem;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 4px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary {
  background: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background: #0284c7;
}

.btn-secondary {
  background: #e5e7eb;
  color: var(--color-text);
}

.btn-secondary:hover {
  background: #d1d5db;
}
```
