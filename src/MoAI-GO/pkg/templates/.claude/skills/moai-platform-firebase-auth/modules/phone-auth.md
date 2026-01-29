# Phone Authentication Module

Comprehensive guide for implementing SMS-based phone number authentication with Firebase, including international support, reCAPTCHA verification, and platform-specific implementations.

---

## Overview

Phone authentication enables users to sign in using their phone number via SMS verification. Firebase handles SMS delivery, verification code generation, and user account management.

Key Features:
- International phone number support with E.164 format
- Automatic reCAPTCHA verification for web
- Auto-verification on Android devices
- Rate limiting and abuse protection built-in

---

## Web Implementation

### Basic Phone Authentication Flow

```typescript
import {
  getAuth,
  RecaptchaVerifier,
  signInWithPhoneNumber,
  PhoneAuthProvider,
  signInWithCredential
} from 'firebase/auth';

const auth = getAuth();

// Step 1: Set up reCAPTCHA verifier
const setupRecaptcha = () => {
  const recaptchaVerifier = new RecaptchaVerifier(auth, 'recaptcha-container', {
    size: 'normal',
    callback: (response: string) => {
      console.log('reCAPTCHA solved');
    },
    'expired-callback': () => {
      console.log('reCAPTCHA expired');
    }
  });
  return recaptchaVerifier;
};

// Step 2: Send verification code
const sendVerificationCode = async (phoneNumber: string) => {
  const recaptchaVerifier = setupRecaptcha();

  try {
    const confirmationResult = await signInWithPhoneNumber(
      auth,
      phoneNumber,
      recaptchaVerifier
    );
    return confirmationResult;
  } catch (error: any) {
    if (error.code === 'auth/invalid-phone-number') {
      throw new Error('Invalid phone number format');
    }
    throw error;
  }
};

// Step 3: Verify the code
const verifyCode = async (confirmationResult: any, code: string) => {
  try {
    const result = await confirmationResult.confirm(code);
    return result.user;
  } catch (error: any) {
    if (error.code === 'auth/invalid-verification-code') {
      throw new Error('Invalid verification code');
    }
    throw error;
  }
};
```

### Invisible reCAPTCHA

```typescript
const setupInvisibleRecaptcha = () => {
  const recaptchaVerifier = new RecaptchaVerifier(auth, 'sign-in-button', {
    size: 'invisible',
    callback: () => {
      onSignInSubmit();
    }
  });
  return recaptchaVerifier;
};
```

---

## Flutter Implementation

### Complete Phone Auth Flow

```dart
import 'package:firebase_auth/firebase_auth.dart';

class PhoneAuthService {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  String? _verificationId;
  int? _resendToken;

  Future<void> sendVerificationCode({
    required String phoneNumber,
    required Function(String verificationId) onCodeSent,
    required Function(PhoneAuthCredential credential) onAutoVerify,
    required Function(String error) onError,
  }) async {
    await _auth.verifyPhoneNumber(
      phoneNumber: phoneNumber,

      verificationCompleted: (PhoneAuthCredential credential) async {
        onAutoVerify(credential);
        await _auth.signInWithCredential(credential);
      },

      verificationFailed: (FirebaseAuthException e) {
        if (e.code == 'invalid-phone-number') {
          onError('Invalid phone number format');
        } else if (e.code == 'too-many-requests') {
          onError('Too many attempts. Please try again later.');
        } else {
          onError(e.message ?? 'Verification failed');
        }
      },

      codeSent: (String verificationId, int? resendToken) {
        _verificationId = verificationId;
        _resendToken = resendToken;
        onCodeSent(verificationId);
      },

      codeAutoRetrievalTimeout: (String verificationId) {
        _verificationId = verificationId;
      },

      timeout: const Duration(seconds: 60),
      forceResendingToken: _resendToken,
    );
  }

  Future<UserCredential> verifyCode(String smsCode) async {
    if (_verificationId == null) {
      throw Exception('No verification ID');
    }

    final credential = PhoneAuthProvider.credential(
      verificationId: _verificationId!,
      smsCode: smsCode,
    );

    return await _auth.signInWithCredential(credential);
  }
}
```

---

## iOS Native Implementation

```swift
import FirebaseAuth

class PhoneAuthManager {
    private var verificationID: String?

    func sendVerificationCode(to phoneNumber: String, completion: @escaping (Result<Void, Error>) -> Void) {
        PhoneAuthProvider.provider().verifyPhoneNumber(phoneNumber, uiDelegate: nil) { [weak self] verificationID, error in
            if let error = error {
                completion(.failure(error))
                return
            }

            self?.verificationID = verificationID
            completion(.success(()))
        }
    }

    func verifyCode(_ code: String, completion: @escaping (Result<User, Error>) -> Void) {
        guard let verificationID = verificationID else {
            completion(.failure(NSError(domain: "PhoneAuth", code: -1,
                                        userInfo: [NSLocalizedDescriptionKey: "No verification ID"])))
            return
        }

        let credential = PhoneAuthProvider.provider().credential(
            withVerificationID: verificationID,
            verificationCode: code
        )

        Auth.auth().signIn(with: credential) { authResult, error in
            if let error = error {
                completion(.failure(error))
                return
            }

            if let user = authResult?.user {
                completion(.success(user))
            }
        }
    }
}
```

---

## Android Native Implementation

```kotlin
import com.google.firebase.auth.*
import java.util.concurrent.TimeUnit

class PhoneAuthManager(private val activity: Activity) {
    private var verificationId: String? = null
    private var resendToken: PhoneAuthProvider.ForceResendingToken? = null

    private val callbacks = object : PhoneAuthProvider.OnVerificationStateChangedCallbacks() {
        override fun onVerificationCompleted(credential: PhoneAuthCredential) {
            signInWithCredential(credential)
        }

        override fun onVerificationFailed(e: FirebaseException) {
            when (e) {
                is FirebaseAuthInvalidCredentialsException -> { /* Invalid number */ }
                is FirebaseTooManyRequestsException -> { /* Quota exceeded */ }
            }
        }

        override fun onCodeSent(
            verificationId: String,
            token: PhoneAuthProvider.ForceResendingToken
        ) {
            this@PhoneAuthManager.verificationId = verificationId
            this@PhoneAuthManager.resendToken = token
        }
    }

    fun sendVerificationCode(phoneNumber: String) {
        val options = PhoneAuthOptions.newBuilder(FirebaseAuth.getInstance())
            .setPhoneNumber(phoneNumber)
            .setTimeout(60L, TimeUnit.SECONDS)
            .setActivity(activity)
            .setCallbacks(callbacks)
            .build()

        PhoneAuthProvider.verifyPhoneNumber(options)
    }

    fun verifyCode(code: String) {
        verificationId?.let { id ->
            val credential = PhoneAuthProvider.getCredential(id, code)
            signInWithCredential(credential)
        }
    }

    private fun signInWithCredential(credential: PhoneAuthCredential) {
        FirebaseAuth.getInstance().signInWithCredential(credential)
            .addOnCompleteListener(activity) { task ->
                if (task.isSuccessful) {
                    val user = task.result?.user
                }
            }
    }
}
```

---

## Phone Number Formatting

Always use E.164 format for phone numbers:

```typescript
// E.164 Format Examples
const validFormats = [
  '+12025551234',      // US
  '+442071234567',     // UK
  '+81312345678',      // Japan
  '+8210123456789',    // South Korea
];

const isValidE164 = (phone: string): boolean => {
  const e164Regex = /^\+[1-9]\d{1,14}$/;
  return e164Regex.test(phone);
};

const formatToE164 = (phone: string, countryCode: string): string => {
  const digits = phone.replace(/\D/g, '');
  const normalized = digits.startsWith('0') ? digits.slice(1) : digits;
  return `+${countryCode}${normalized}`;
};
```

---

## Error Handling

```typescript
const handlePhoneAuthError = (error: any): string => {
  switch (error.code) {
    case 'auth/invalid-phone-number':
      return 'Please enter a valid phone number';
    case 'auth/too-many-requests':
      return 'Too many attempts. Please try again later';
    case 'auth/quota-exceeded':
      return 'SMS quota exceeded';
    case 'auth/invalid-verification-code':
      return 'Invalid verification code';
    case 'auth/code-expired':
      return 'Verification code expired';
    case 'auth/captcha-check-failed':
      return 'reCAPTCHA verification failed';
    default:
      return 'An error occurred';
  }
};
```

---

## Testing with Firebase Emulator

```typescript
import { connectAuthEmulator } from 'firebase/auth';

if (process.env.NODE_ENV === 'development') {
  connectAuthEmulator(auth, 'http://localhost:9099');
}

// Test phone numbers (emulator only)
const testPhoneNumbers = ['+15551234567', '+15559876543'];
```

---

## Best Practices

Rate Limiting and Abuse Prevention:
- Firebase automatically rate limits SMS sending
- Implement client-side rate limiting for better UX
- Show countdown timer after sending code
- Limit resend attempts per session

User Experience Guidelines:
- Pre-fill country code based on device locale
- Show clear instructions for phone number format
- Provide countdown timer for code expiration
- Allow code resend after timeout period

Security Considerations:
- Validate phone number format before sending
- Implement proper error handling for all edge cases
- Use invisible reCAPTCHA for seamless experience
- Consider SIM swap attack mitigation for sensitive operations

Cost Management:
- SMS messages incur costs per message sent
- Implement proper rate limiting to control costs
- Use test phone numbers during development
- Monitor SMS usage in Firebase Console

---

Version: 1.0.0
Last Updated: 2025-12-07
Parent Skill: moai-platform-firebase-auth
