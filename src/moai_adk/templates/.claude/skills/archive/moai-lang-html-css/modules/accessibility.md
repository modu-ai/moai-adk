# HTML5 & CSS3 Accessibility Patterns — WCAG 2.1 AA Compliance

_Last updated: 2025-11-22_

## WCAG 2.1 AA Compliance Checklist

### Perceivable (Usable by all senses)

#### Color Contrast (1.4.3)
```css
/* ✅ 4.5:1 contrast for normal text */
.text-primary {
    color: #000000;
    background-color: #FFFFFF;
    /* 21:1 contrast ratio */
}

/* ✅ 3:1 contrast for large text */
.text-large {
    font-size: 18px;
    font-weight: bold;
    color: #333333;
    background-color: #FFFFFF;
    /* 7:1 contrast ratio */
}

/* ❌ AVOID: Insufficient contrast */
.text-bad {
    color: #888888;
    background-color: #FFFFFF;
    /* 4.5:1 - fails for normal text */
}
```

#### Text Alternatives (1.1.1)
```html
<!-- ✅ Images have alt text -->
<img src="chart.png" alt="Sales growth from Q1 to Q4, reaching $2M revenue">

<!-- ✅ Decorative images use empty alt -->
<img src="separator.png" alt="">

<!-- ✅ Icons with aria-label -->
<button aria-label="Close navigation menu">
    <svg>...</svg>
</button>
```

### Operable (Keyboard navigation)

#### Keyboard Navigation (2.1.1)
```html
<!-- ✅ Proper keyboard order -->
<form>
    <label for="name">Name:</label>
    <input id="name" type="text" required>

    <label for="email">Email:</label>
    <input id="email" type="email" required>

    <button type="submit">Submit</button>
</form>

<!-- ✅ Visible focus indicators -->
<style>
    button:focus {
        outline: 3px solid #4A90E2;
        outline-offset: 2px;
    }

    input:focus {
        outline: 2px solid #4A90E2;
    }
</style>

<!-- ✅ Skip links -->
<a href="#main-content" class="skip-link">Skip to main content</a>

<style>
    .skip-link {
        position: absolute;
        top: -40px;
        left: 0;
        background: #000;
        color: #FFF;
        padding: 8px;
    }

    .skip-link:focus {
        top: 0;
    }
</style>
```

#### No Keyboard Trap (2.1.2)
```html
<!-- ✅ Users can tab out of all elements -->
<div role="dialog" aria-modal="true">
    <button>First Button</button>
    <input type="text">
    <button>Last Button</button>
    <!-- Tab cycles properly -->
</div>

<!-- ✅ Modal focus management -->
<div id="modal" role="dialog" aria-modal="true">
    <button id="close">Close</button>
    <!-- Focus trap within modal -->
</div>

<script>
    const modal = document.getElementById('modal');
    const focusableElements = modal.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    modal.addEventListener('keydown', (e) => {
        if (e.key === 'Tab') {
            if (e.shiftKey && document.activeElement === firstElement) {
                lastElement.focus();
                e.preventDefault();
            } else if (!e.shiftKey && document.activeElement === lastElement) {
                firstElement.focus();
                e.preventDefault();
            }
        }
    });
</script>
```

### Understandable (Clear language)

#### Language (3.1.1)
```html
<!-- ✅ Declare page language -->
<html lang="en">
    <body>
        <!-- ✅ Mark language changes -->
        <p>Welcome to our site. <span lang="es">Bienvenido a nuestro sitio.</span></p>
    </body>
</html>
```

#### Labels (3.3.2)
```html
<!-- ✅ Explicit labels -->
<label for="email">Email Address:</label>
<input id="email" type="email" required>

<!-- ✅ Error messages with aria-describedby -->
<input
    id="password"
    type="password"
    aria-describedby="password-error"
    required
>
<span id="password-error" role="alert">
    Password must be at least 8 characters
</span>
```

### Robust (Works with assistive tech)

#### ARIA Labels (4.1.2)
```html
<!-- ✅ Hidden labels for screen readers -->
<button aria-label="Close menu" class="menu-close-btn">×</button>

<!-- ✅ aria-describedby for additional info -->
<button
    id="delete-btn"
    aria-describedby="delete-warning"
>
    Delete Account
</button>
<div id="delete-warning">This action cannot be undone</div>

<!-- ✅ Live regions for dynamic content -->
<div aria-live="polite" aria-atomic="true" id="status">
    Ready
</div>

<script>
    function updateStatus(message) {
        document.getElementById('status').textContent = message;
        // Screen readers announce changes
    }
</script>

<!-- ✅ Landmark regions -->
<header>
    <nav>Main Navigation</nav>
</header>
<main id="main-content">
    Primary content
</main>
<aside>
    Related content
</aside>
<footer>
    Footer information
</footer>
```

## Accessible Form Patterns

### Complete Accessible Form
```html
<form>
    <!-- ✅ Proper fieldset and legend -->
    <fieldset>
        <legend>Contact Information</legend>

        <div class="form-group">
            <label for="full-name">Full Name:</label>
            <input
                id="full-name"
                type="text"
                required
                aria-required="true"
            >
        </div>

        <div class="form-group">
            <label for="email">Email:</label>
            <input
                id="email"
                type="email"
                required
                aria-required="true"
                aria-describedby="email-format"
            >
            <span id="email-format">Format: example@domain.com</span>
        </div>

        <div class="form-group">
            <fieldset>
                <legend>Message Type</legend>
                <label>
                    <input type="radio" name="type" value="general">
                    General Inquiry
                </label>
                <label>
                    <input type="radio" name="type" value="support">
                    Support Request
                </label>
            </fieldset>
        </div>

        <button type="submit">Send Message</button>
    </fieldset>
</form>
```

## Accessible Table Pattern
```html
<!-- ✅ Proper table structure -->
<table role="table" aria-label="Employee Directory">
    <caption>Employees currently on staff</caption>
    <thead>
        <tr>
            <th scope="col">Name</th>
            <th scope="col">Department</th>
            <th scope="col">Email</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>John Smith</td>
            <td>Engineering</td>
            <td><a href="mailto:john@example.com">john@example.com</a></td>
        </tr>
    </tbody>
</table>
```

---

**Last Updated**: 2025-11-22
**Related**: moai-lang-html-css/SKILL.md, modules/css-optimization.md

