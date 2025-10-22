# Mobile App Development Reference

> Official documentation for Flutter and React Native

---

## Official Documentation Links

| Framework | Version | Documentation | Status |
|-----------|---------|--------------|--------|
| **Flutter** | 3.27.0 | https://docs.flutter.dev/ | ✅ Current (2025) |
| **React Native** | 0.76.0 | https://reactnative.dev/docs | ✅ Current (2025) |
| **Expo** | 52.0.0 | https://docs.expo.dev/ | ✅ Current (2025) |

---

## Flutter Best Practices

### StatelessWidget
```dart
class MyButton extends StatelessWidget {
  final String label;
  final VoidCallback onPressed;

  const MyButton({
    Key? key,
    required this.label,
    required this.onPressed,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: onPressed,
      child: Text(label),
    );
  }
}
```

### State Management (Provider)
```dart
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class Counter with ChangeNotifier {
  int _count = 0;
  int get count => _count;

  void increment() {
    _count++;
    notifyListeners();
  }
}

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider(
      create: (_) => Counter(),
      child: MaterialApp(home: CounterScreen()),
    );
  }
}
```

---

## React Native Best Practices

### Functional Component
```typescript
import React, { useState } from 'react';
import { View, Text, Button } from 'react-native';

export function Counter() {
  const [count, setCount] = useState(0);

  return (
    <View>
      <Text>Count: {count}</Text>
      <Button title="Increment" onPress={() => setCount(c => c + 1)} />
    </View>
  );
}
```

---

**Last Updated**: 2025-10-22
**Versions**: Flutter 3.27.0, React Native 0.76.0
