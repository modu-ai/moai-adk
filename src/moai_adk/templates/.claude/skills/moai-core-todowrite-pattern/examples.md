# TodoWrite Pattern Examples

Practical implementations demonstrating incremental development with the TodoWrite pattern.

---

## Basic TodoWrite Pattern

### Simple API Implementation

```python
# Phase 1: Scaffold with TodoWrite placeholders
class UserController:
    """User management controller."""

    def __init__(self, db):
        self.db = db

    def create_user(self, email: str, password: str, name: str) -> dict:
        """Create a new user account."""
        TodoWrite("Validate email format")
        TodoWrite("Check if email already exists")
        TodoWrite("Validate password strength")
        TodoWrite("Hash password with bcrypt")
        TodoWrite("Create user record in database")
        TodoWrite("Send welcome email")
        TodoWrite("Return user object (excluding password)")

    def get_user(self, user_id: int) -> dict:
        """Retrieve user by ID."""
        TodoWrite("Query database for user")
        TodoWrite("Handle user not found error")
        TodoWrite("Return sanitized user object")

    def update_user(self, user_id: int, updates: dict) -> dict:
        """Update user information."""
        TodoWrite("Validate user exists")
        TodoWrite("Validate update fields")
        TodoWrite("Apply updates to database")
        TodoWrite("Return updated user")

    def delete_user(self, user_id: int) -> bool:
        """Soft delete user account."""
        TodoWrite("Validate user exists")
        TodoWrite("Mark user as deleted (soft delete)")
        TodoWrite("Invalidate user sessions")
        TodoWrite("Return success status")
```

```python
# Phase 2: Implement critical path (create_user validation)
class UserController:
    def __init__(self, db):
        self.db = db

    def create_user(self, email: str, password: str, name: str) -> dict:
        """Create a new user account."""

        # ✅ DONE: Validate email format
        import re
        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, email):
            raise ValueError(f"Invalid email format: {email}")

        # ✅ DONE: Check if email already exists
        existing_user = self.db.query("SELECT id FROM users WHERE email = ?", (email,))
        if existing_user:
            raise ValueError(f"Email already registered: {email}")

        # ✅ DONE: Validate password strength
        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters")
        if not any(c.isupper() for c in password):
            raise ValueError("Password must contain uppercase letter")
        if not any(c.isdigit() for c in password):
            raise ValueError("Password must contain a number")

        TodoWrite("Hash password with bcrypt")
        TodoWrite("Create user record in database")
        TodoWrite("Send welcome email")
        TodoWrite("Return user object (excluding password)")
```

```python
# Phase 3: Implement security and persistence
class UserController:
    def __init__(self, db):
        self.db = db

    def create_user(self, email: str, password: str, name: str) -> dict:
        """Create a new user account."""

        # ✅ DONE: Validation (from Phase 2)
        import re
        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, email):
            raise ValueError(f"Invalid email format: {email}")

        existing_user = self.db.query("SELECT id FROM users WHERE email = ?", (email,))
        if existing_user:
            raise ValueError(f"Email already registered: {email}")

        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters")

        # ✅ DONE: Hash password
        import bcrypt
        hashed_password = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())

        # ✅ DONE: Create user record
        user_id = self.db.execute(
            "INSERT INTO users (email, password_hash, name, created_at) VALUES (?, ?, ?, ?)",
            (email, hashed_password.decode('utf-8'), name, datetime.utcnow())
        )

        TodoWrite("Send welcome email")
        TodoWrite("Return user object (excluding password)")


# Phase 4: Complete implementation
class UserController:
    def __init__(self, db, email_service):
        self.db = db
        self.email_service = email_service

    def create_user(self, email: str, password: str, name: str) -> dict:
        """Create a new user account."""

        # ✅ DONE: Validation
        import re
        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, email):
            raise ValueError(f"Invalid email format: {email}")

        existing_user = self.db.query("SELECT id FROM users WHERE email = ?", (email,))
        if existing_user:
            raise ValueError(f"Email already registered: {email}")

        if len(password) < 8:
            raise ValueError("Password must be at least 8 characters")

        # ✅ DONE: Hash password
        import bcrypt
        hashed_password = bcrypt.hashpw(password.encode('utf-8'), bcrypt.gensalt())

        # ✅ DONE: Create user record
        user_id = self.db.execute(
            "INSERT INTO users (email, password_hash, name, created_at) VALUES (?, ?, ?, ?)",
            (email, hashed_password.decode('utf-8'), name, datetime.utcnow())
        )

        # ✅ DONE: Send welcome email
        self.email_service.send_welcome_email(email, name)

        # ✅ DONE: Return user object
        return {
            'id': user_id,
            'email': email,
            'name': name,
            'created_at': datetime.utcnow().isoformat()
        }
```

---

## Complex Feature Example

### E-commerce Checkout System

```typescript
// Phase 1: High-level scaffold
class CheckoutService {
  constructor(
    private cartService: CartService,
    private paymentGateway: PaymentGateway,
    private inventoryService: InventoryService,
    private orderService: OrderService,
    private emailService: EmailService
  ) {}

  async processCheckout(
    userId: string,
    paymentInfo: PaymentInfo
  ): Promise<Order> {
    TodoWrite("Retrieve user's cart");
    TodoWrite("Validate cart items are still available");
    TodoWrite("Calculate total with taxes and shipping");
    TodoWrite("Reserve inventory for cart items");
    TodoWrite("Process payment");
    TodoWrite("Create order record");
    TodoWrite("Update inventory");
    TodoWrite("Clear user's cart");
    TodoWrite("Send order confirmation email");
    TodoWrite("Return order confirmation");
  }
}

// Phase 2: Implement cart and validation logic
class CheckoutService {
  async processCheckout(
    userId: string,
    paymentInfo: PaymentInfo
  ): Promise<Order> {
    // ✅ DONE: Retrieve cart
    const cart = await this.cartService.getCart(userId);

    if (!cart || cart.items.length === 0) {
      throw new Error("Cart is empty");
    }

    // ✅ DONE: Validate availability
    for (const item of cart.items) {
      const available = await this.inventoryService.checkAvailability(
        item.productId,
        item.quantity
      );

      if (!available) {
        throw new Error(`Product ${item.productId} is out of stock`);
      }
    }

    // ✅ DONE: Calculate total
    const subtotal = cart.items.reduce(
      (sum, item) => sum + item.price * item.quantity,
      0
    );
    const tax = subtotal * 0.1; // 10% tax
    const shipping = cart.items.length > 0 ? 10 : 0; // $10 flat shipping
    const total = subtotal + tax + shipping;

    TodoWrite("Reserve inventory for cart items");
    TodoWrite("Process payment");
    TodoWrite("Create order record");
    TodoWrite("Update inventory");
    TodoWrite("Clear user's cart");
    TodoWrite("Send order confirmation email");

    return { total, items: cart.items };
  }
}

// Phase 3: Implement payment and order creation
class CheckoutService {
  async processCheckout(
    userId: string,
    paymentInfo: PaymentInfo
  ): Promise<Order> {
    // ✅ DONE: Cart retrieval and validation (from Phase 2)
    const cart = await this.cartService.getCart(userId);

    if (!cart || cart.items.length === 0) {
      throw new Error("Cart is empty");
    }

    for (const item of cart.items) {
      const available = await this.inventoryService.checkAvailability(
        item.productId,
        item.quantity
      );

      if (!available) {
        throw new Error(`Product ${item.productId} is out of stock`);
      }
    }

    const subtotal = cart.items.reduce(
      (sum, item) => sum + item.price * item.quantity,
      0
    );
    const tax = subtotal * 0.1;
    const shipping = cart.items.length > 0 ? 10 : 0;
    const total = subtotal + tax + shipping;

    // ✅ DONE: Reserve inventory
    const reservationId = await this.inventoryService.reserveItems(
      cart.items.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
      }))
    );

    try {
      // ✅ DONE: Process payment
      const paymentResult = await this.paymentGateway.charge({
        amount: total,
        currency: "USD",
        paymentMethod: paymentInfo,
      });

      if (!paymentResult.success) {
        throw new Error(`Payment failed: ${paymentResult.error}`);
      }

      // ✅ DONE: Create order
      const order = await this.orderService.createOrder({
        userId,
        items: cart.items,
        subtotal,
        tax,
        shipping,
        total,
        paymentId: paymentResult.transactionId,
        status: "confirmed",
      });

      TodoWrite("Update inventory (commit reservation)");
      TodoWrite("Clear user's cart");
      TodoWrite("Send order confirmation email");

      return order;
    } catch (error) {
      // Rollback inventory reservation on failure
      await this.inventoryService.releaseReservation(reservationId);
      throw error;
    }
  }
}

// Phase 4: Complete implementation
class CheckoutService {
  async processCheckout(
    userId: string,
    paymentInfo: PaymentInfo
  ): Promise<Order> {
    // ✅ ALL DONE: Full implementation
    const cart = await this.cartService.getCart(userId);

    if (!cart || cart.items.length === 0) {
      throw new Error("Cart is empty");
    }

    for (const item of cart.items) {
      const available = await this.inventoryService.checkAvailability(
        item.productId,
        item.quantity
      );

      if (!available) {
        throw new Error(`Product ${item.productId} is out of stock`);
      }
    }

    const subtotal = cart.items.reduce(
      (sum, item) => sum + item.price * item.quantity,
      0
    );
    const tax = subtotal * 0.1;
    const shipping = cart.items.length > 0 ? 10 : 0;
    const total = subtotal + tax + shipping;

    const reservationId = await this.inventoryService.reserveItems(
      cart.items.map((item) => ({
        productId: item.productId,
        quantity: item.quantity,
      }))
    );

    try {
      const paymentResult = await this.paymentGateway.charge({
        amount: total,
        currency: "USD",
        paymentMethod: paymentInfo,
      });

      if (!paymentResult.success) {
        throw new Error(`Payment failed: ${paymentResult.error}`);
      }

      const order = await this.orderService.createOrder({
        userId,
        items: cart.items,
        subtotal,
        tax,
        shipping,
        total,
        paymentId: paymentResult.transactionId,
        status: "confirmed",
      });

      // ✅ DONE: Commit inventory changes
      await this.inventoryService.commitReservation(reservationId);

      // ✅ DONE: Clear cart
      await this.cartService.clearCart(userId);

      // ✅ DONE: Send confirmation email
      await this.emailService.sendOrderConfirmation(userId, order);

      return order;
    } catch (error) {
      await this.inventoryService.releaseReservation(reservationId);
      throw error;
    }
  }
}
```

---

## TodoWrite Tracking

### Tracking Progress Script

```python
#!/usr/bin/env python3
"""
Track TodoWrite progress across the codebase.
"""

import os
import re
from pathlib import Path
from collections import defaultdict


def find_todowrite_calls(directory: Path) -> dict:
    """Find all TodoWrite calls in Python/TypeScript files."""

    todowrite_pattern = re.compile(r'TodoWrite\(["\'](.+?)["\']\)', re.IGNORECASE)
    results = defaultdict(list)

    for ext in ['*.py', '*.ts', '*.tsx', '*.js', '*.jsx']:
        for filepath in directory.rglob(ext):
            try:
                content = filepath.read_text()

                for line_num, line in enumerate(content.splitlines(), 1):
                    match = todowrite_pattern.search(line)
                    if match:
                        todo_description = match.group(1)
                        results[str(filepath)].append({
                            'line': line_num,
                            'description': todo_description,
                            'code': line.strip()
                        })
            except Exception as e:
                print(f"Error reading {filepath}: {e}")

    return results


def print_report(results: dict):
    """Print TodoWrite report."""
    total_todos = sum(len(todos) for todos in results.values())

    print("=" * 70)
    print(f"📝 TodoWrite Progress Report")
    print("=" * 70)
    print(f"\nTotal TodoWrite calls: {total_todos}")
    print(f"Files with TODOs: {len(results)}\n")

    if total_todos == 0:
        print("✅ All TODOs completed!")
        return

    for filepath, todos in sorted(results.items()):
        print(f"\n📄 {filepath} ({len(todos)} TODOs)")
        print("-" * 70)

        for todo in todos:
            print(f"  Line {todo['line']}: {todo['description']}")
            print(f"    → {todo['code']}\n")


def main():
    """Main entry point."""
    project_root = Path.cwd()
    src_dir = project_root / 'src'

    if not src_dir.exists():
        print(f"Error: {src_dir} not found")
        return 1

    results = find_todowrite_calls(src_dir)
    print_report(results)

    return 0 if not results else 1


if __name__ == "__main__":
    exit(main())
```

**Usage**:

```bash
# Run from project root
python scripts/track-todowrite.py

# Output:
# ======================================================================
# 📝 TodoWrite Progress Report
# ======================================================================
#
# Total TodoWrite calls: 12
# Files with TODOs: 3
#
# 📄 src/controllers/user.py (5 TODOs)
# ----------------------------------------------------------------------
#   Line 45: Hash password with bcrypt
#     → TodoWrite("Hash password with bcrypt")
#
#   Line 46: Create user record in database
#     → TodoWrite("Create user record in database")
# ...
```

---

## Pre-commit Hook for TodoWrite

### Prevent Committing TODOs to Main Branch

```bash
#!/bin/bash
# .git/hooks/pre-commit
# Prevent commits to main with TodoWrite placeholders

BRANCH=$(git rev-parse --abbrev-ref HEAD)

if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
  # Check for TodoWrite in staged files
  STAGED_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(py|ts|tsx|js|jsx)$')

  if [ -n "$STAGED_FILES" ]; then
    TODO_COUNT=$(git diff --cached | grep -c "TodoWrite" || true)

    if [ "$TODO_COUNT" -gt 0 ]; then
      echo "❌ ERROR: Cannot commit TodoWrite placeholders to $BRANCH branch"
      echo "Found $TODO_COUNT TodoWrite call(s) in staged files"
      echo ""
      echo "Please complete all TodoWrite placeholders before merging to $BRANCH"
      echo "Or commit to a feature branch instead"
      exit 1
    fi
  fi
fi

exit 0
```

---

## MVP vs Full Implementation

### Example: Blog Platform

```python
# MVP (Week 1): Core blogging functionality
class BlogPost:
    def __init__(self, db):
        self.db = db

    def create_post(self, title: str, content: str, author_id: int) -> dict:
        """Create a new blog post."""

        # ✅ MVP: Basic validation and creation
        if not title or not content:
            raise ValueError("Title and content required")

        post_id = self.db.execute(
            "INSERT INTO posts (title, content, author_id, created_at) VALUES (?, ?, ?, ?)",
            (title, content, author_id, datetime.utcnow())
        )

        # 🔜 Future features (post-MVP)
        TodoWrite("Add support for rich text formatting (Markdown)")
        TodoWrite("Implement post categories and tags")
        TodoWrite("Add featured image upload")
        TodoWrite("Enable post scheduling (publish_at)")
        TodoWrite("Implement SEO metadata (title, description, keywords)")

        return {'id': post_id, 'title': title, 'status': 'published'}

    def get_posts(self, limit: int = 10) -> list:
        """Get recent blog posts."""

        # ✅ MVP: Simple query
        posts = self.db.query(
            "SELECT id, title, content, author_id, created_at FROM posts ORDER BY created_at DESC LIMIT ?",
            (limit,)
        )

        # 🔜 Future features
        TodoWrite("Add pagination support")
        TodoWrite("Implement full-text search")
        TodoWrite("Add filtering by category/tag")

        return posts


# Full Release (Week 4): All features implemented
class BlogPost:
    def __init__(self, db, storage, search_engine):
        self.db = db
        self.storage = storage
        self.search = search_engine

    def create_post(
        self,
        title: str,
        content: str,
        author_id: int,
        categories: list[str] = None,
        featured_image: bytes = None,
        publish_at: datetime = None,
        seo_metadata: dict = None
    ) -> dict:
        """Create a new blog post with full features."""

        # ✅ ALL DONE: Complete implementation
        if not title or not content:
            raise ValueError("Title and content required")

        # Markdown processing
        import markdown
        html_content = markdown.markdown(content)

        # Upload featured image
        image_url = None
        if featured_image:
            image_url = self.storage.upload(featured_image, f"posts/{uuid4()}.jpg")

        # Determine publish status
        status = 'scheduled' if publish_at and publish_at > datetime.utcnow() else 'published'

        # Create post
        post_id = self.db.execute(
            """INSERT INTO posts
               (title, content, html_content, author_id, featured_image,
                publish_at, status, seo_title, seo_description, created_at)
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
            (title, content, html_content, author_id, image_url,
             publish_at or datetime.utcnow(), status,
             seo_metadata.get('title') if seo_metadata else title,
             seo_metadata.get('description') if seo_metadata else content[:160],
             datetime.utcnow())
        )

        # Add categories
        if categories:
            for category in categories:
                self.db.execute(
                    "INSERT INTO post_categories (post_id, category) VALUES (?, ?)",
                    (post_id, category)
                )

        # Index for search
        self.search.index_document(post_id, title, content)

        return {
            'id': post_id,
            'title': title,
            'status': status,
            'image_url': image_url,
            'categories': categories or []
        }
```

---

## Best Practices Summary

### ✅ DO

```python
# ✅ Descriptive TodoWrite messages
TodoWrite("Validate email format with regex")
TodoWrite("Hash password using bcrypt with cost factor 12")
TodoWrite("Save user to database with transaction")

# ✅ Group related TODOs
def signup(email, password):
    # Validation group
    TodoWrite("Validate email format")
    TodoWrite("Validate password strength")
    TodoWrite("Check email not already registered")

    # Security group
    TodoWrite("Hash password with bcrypt")
    TodoWrite("Generate email verification token")

    # Persistence group
    TodoWrite("Save user to database")
    TodoWrite("Send verification email")

# ✅ Track progress in documentation
"""
## TodoWrite Progress

### Phase 1: Validation (COMPLETED ✅)
- [x] Validate email format
- [x] Validate password strength
- [x] Check email uniqueness

### Phase 2: Security (IN PROGRESS 🔄)
- [x] Hash password
- [ ] Generate verification token

### Phase 3: Persistence (PENDING ⏳)
- [ ] Save to database
- [ ] Send email
"""
```

### ❌ DON'T

```python
# ❌ Generic TodoWrite messages
TodoWrite("TODO")
TodoWrite("Fix this")
TodoWrite("Complete later")

# ❌ Too granular (over-scaffolding)
def calculate_total(items):
    TodoWrite("Initialize total variable")
    TodoWrite("Create for loop")
    TodoWrite("Add item price to total")
    TodoWrite("Return total")

# Instead, just implement it (simple logic):
def calculate_total(items):
    return sum(item.price for item in items)

# ❌ Leaving TODOs indefinitely
# TodoWrite calls should be temporary placeholders, not permanent markers
# Use proper TODO comments for long-term reminders:
# TODO(username): Implement caching layer (Q2 2025)
```

---

**See also**: [reference.md](./reference.md) for advanced TodoWrite patterns and workflow integration
