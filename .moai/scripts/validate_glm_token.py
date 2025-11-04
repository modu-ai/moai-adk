#!/usr/bin/env python3
"""
GLM API Token Validation Script

This script validates the GLM API token to ensure it works correctly
and hasn't expired.
"""

import json
import os
import sys
import requests
from pathlib import Path

def get_glm_token():
    """Get GLM token from environment variable or settings"""
    # First check environment variable directly
    token = os.environ.get('GLM_API_KEY')
    if token:
        print(f"✅ Found GLM_API_KEY environment variable")
        print(f"Token format: {token[:20]}...{token[-20:] if len(token) > 40 else token}")
        return token

    # If not found, try to get from settings files
    settings_paths = [
        Path.home() / '.claude' / 'settings.json',
        Path('.claude' / 'settings.json'),
        Path('.moai' / 'config.json')
    ]

    for settings_path in settings_paths:
        if settings_path.exists():
            try:
                with open(settings_path, 'r', encoding='utf-8') as f:
                    settings = json.load(f)

                env_vars = settings.get('env', {})
                auth_token = env_vars.get('ANTHROPIC_AUTH_TOKEN', '')

                if auth_token.startswith('${') and auth_token.endswith('}'):
                    var_name = auth_token[2:-1]
                    token = os.environ.get(var_name)
                    if token:
                        print(f"✅ Found token from {var_name} in {settings_path}")
                        print(f"Token format: {token[:20]}...{token[-20:] if len(token) > 40 else token}")
                        return token

                elif auth_token and not auth_token.startswith('${'):
                    print(f"⚠️  Found hardcoded token in {settings_path}")
                    print(f"Token format: {auth_token[:20]}...{auth_token[-20:] if len(auth_token) > 40 else auth_token}")
                    return auth_token

            except Exception as e:
                print(f"❌ Error reading {settings_path}: {e}")
                continue

    print("❌ No GLM token found in environment variables or settings files")
    return None

def validate_token(token):
    """Validate token by making a test API call"""
    if not token:
        print("❌ No token provided for validation")
        return False

    print("\n🔍 Testing GLM API token...")

    # Test API endpoint
    url = "https://api.z.ai/api/anthropic/v1/messages"
    headers = {
        "Content-Type": "application/json",
        "x-api-key": token,
        "anthropic-version": "2023-06-01"
    }

    # Simple test request
    test_payload = {
        "model": "claude-3-haiku-20240307",
        "max_tokens": 10,
        "messages": [
            {"role": "user", "content": "test"}
        ]
    }

    try:
        response = requests.post(url, headers=headers, json=test_payload, timeout=30)

        if response.status_code == 200:
            print("✅ Token is valid and working correctly!")
            print(f"Response status: {response.status_code}")
            return True
        else:
            print(f"❌ Token validation failed")
            print(f"Response status: {response.status_code}")
            print(f"Response body: {response.text[:200]}...")
            return False

    except requests.exceptions.Timeout:
        print("❌ Request timed out - check network connectivity")
        return False
    except requests.exceptions.RequestException as e:
        print(f"❌ Request failed: {e}")
        return False

def main():
    """Main validation function"""
    print("🎯 GLM API Token Validator")
    print("=" * 50)

    # Get token
    token = get_glm_token()

    if not token:
        print("\n🚨 Token not found!")
        print("\nTo fix this, set your GLM_API_KEY environment variable:")
        print("export GLM_API_KEY='your_token_here'")
        print("\nOr add it to your shell profile (e.g., ~/.zshrc, ~/.bashrc)")
        sys.exit(1)

    # Validate token
    is_valid = validate_token(token)

    if is_valid:
        print("\n🎉 Token validation successful!")
        print("Your GLM API configuration should now work correctly.")
        sys.exit(0)
    else:
        print("\n🚨 Token validation failed!")
        print("\nPossible solutions:")
        print("1. Check if your GLM_API_KEY is set correctly")
        print("2. Verify your token hasn't expired")
        print("3. Contact GLM API provider for a new token")
        print("4. Check the token format (no extra whitespace)")
        sys.exit(1)

if __name__ == "__main__":
    main()