import 'package:flutter/material.dart';
import 'dart:io';

/// Example of Capability and Policy classes
/// for handling platform-specific behavior
class CapabilityPolicyExample extends StatelessWidget {
  const CapabilityPolicyExample({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Capability & Policy Example')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Using Policy class for business logic
            if (Policy().shouldShowPurchaseButton())
              ElevatedButton(
                onPressed: () => Capability().openBrowser(),
                child: const Text('Buy in Browser'),
              )
            else
              const Text('Purchase not available on this platform'),
          ],
        ),
      ),
    );
  }
}

/// Capability class - defines what the code CAN do
class Capability {
  /// Check if browser is available
  bool hasBrowserCapability() {
    return Platform.isAndroid || Platform.isIOS || Platform.isMacOS;
  }

  /// Open browser (implementation depends on platform)
  void openBrowser() {
    if (hasBrowserCapability()) {
      // Launch browser implementation would go here
      print('Opening browser...');
    }
  }
}

/// Policy class - defines what the code SHOULD do
class Policy {
  /// Policy: don't show purchase button on iOS
  bool shouldShowPurchaseButton() {
    return !Platform.isIOS;
  }

  /// Policy: use specific payment provider based on platform
  String getPaymentProvider() {
    if (Platform.isAndroid) {
      return 'Google Play';
    } else if (Platform.isIOS) {
      return 'Apple App Store';
    } else {
      return 'Web Payment';
    }
  }
}
