"""
Test Code Sample for TRUST 5 Validation
Purpose: Validate that trust-checker can load skills and verify code
"""

def calculate_total(items):
    """Calculate total price of items."""
    total = 0
    for item in items:
        total += item['price'] * item['quantity']
    return total


def authenticate_user(username, password):
    """Authenticate user credentials."""
    # WARNING: Potential security issue - no password hashing
    if username == "admin" and password == "admin123":
        return {"status": "success", "user_id": 1}
    return {"status": "failed"}


class UserService:
    """User management service."""
    
    def __init__(self, db_connection):
        self.db = db_connection
    
    def get_user(self, user_id):
        """Get user by ID."""
        # WARNING: Potential SQL injection
        query = f"SELECT * FROM users WHERE id = {user_id}"
        return self.db.execute(query)
    
    def create_user(self, username, email):
        """Create new user."""
        # Missing input validation
        user = {
            "username": username,
            "email": email,
            "created_at": "2025-01-01"
        }
        return self.db.insert("users", user)
