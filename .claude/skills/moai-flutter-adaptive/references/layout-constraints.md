# Layout Constraints in Flutter

## Core Rule

**Constraints go down. Sizes go up. Parent sets position.**

Flutter layout can't be understood without knowing this rule.

In more detail:

* A widget gets its **constraints** from its **parent**. A constraint is just a set of 4 doubles: a minimum and maximum width, and a minimum and maximum height.
* Then the widget goes through its own list of **children**. One by one, the widget tells its children what their **constraints** are, and then asks each child what size it wants to be.
* Then, widget positions its **children** (horizontally in the x axis, and vertically in the y axis), one by one.
* And, finally, widget tells its parent about its own **size** (within the original constraints, of course).

## Limitations

Flutter's layout engine is designed to be a one-pass process:

* A widget can decide its own size only within the constraints given to it by its parent. This means a widget usually **can't have any size it wants**.
* A widget **can't know and doesn't decide its own position in the screen**, since it's the widget's parent who decides the position of the widget.
* Since the parent's size and position also depends on its own parent, it's impossible to precisely define the size and position of any widget without taking into consideration the tree as a whole.
* If a child wants a different size from its parent and parent doesn't have enough information to align it, then the child's size might be ignored. **Be specific when defining alignment.**

## Widget Types

Generally, there are three kinds of boxes, in terms of how they handle their constraints:

* Those that try to be as big as possible (e.g., [`Center`][], [`ListView`][])
* Those that try to be same size as their children (e.g., [`Transform`][], [`Opacity`][])
* Those that try to be a particular size (e.g., [`Image`][], [`Text`][])

Some widgets vary from type to type based on their constructor arguments (e.g., [`Container`][] varies based on width/height parameters).

## Examples

### Example 1: Container fills screen

```dart
Container(color: Colors.red)
```

The screen is parent of `Container`, and it forces `Container` to be exactly same size as screen. So `Container` fills screen and paints it red.

### Example 2: Container with fixed size in screen

```dart
Container(width: 100, height: 100, color: Colors.red)
```

The red `Container` wants to be 100×100, but it can't, because the screen forces it to be exactly the same size as the screen. So `Container` fills the screen.

### Example 3: Centered Container

```dart
Center(child: Container(width: 100, height: 100, color: Colors.red))
```

The screen forces the `Center` to be exactly the same size as the screen, so `Center` fills the screen. The `Center` tells the `Container` that it can be any size it wants, but not bigger than the screen. Now the `Container` can indeed be 100×100.

### Example 4: Aligned Container

```dart
Align(
  alignment: Alignment.bottomRight,
  child: Container(width: 100, height: 100, color: Colors.red),
)
```

`Align` also tells the `Container` that it can be any size it wants, but if there is empty space it won't center the `Container`. Instead, it aligns the container to the bottom-right of the available space.

### Example 5: Infinite Container

```dart
Center(
  child: Container(
    width: double.infinity,
    height: double.infinity,
    color: Colors.red,
  ),
)
```

The screen forces the `Center` to be exactly the same size as the screen, so `Center` fills the screen. The `Center` tells the `Container` that it can be any size it wants, but not bigger than the screen. The `Container` wants to be of infinite size, but since it can't be bigger than the screen, it just fills the screen.

### Example 6: Empty Container

```dart
Center(child: Container(color: Colors.red))
```

The screen forces the `Center` to be exactly the same size as the screen, so `Center` fills the screen. The `Center` tells the `Container` that it can be any size it wants, but not bigger than the screen. Since the `Container` has no child and no fixed size, it decides it wants to be as big as possible, so it fills the whole screen.

### Example 7: Nested Containers

```dart
Center(
  child: Container(
    color: Colors.red,
    child: Container(color: Colors.green, width: 30, height: 30),
  ),
)
```

The screen forces the `Center` to be exactly the same size as the screen, so `Center` fills the screen. The `Center` tells the red `Container` that it can be any size it wants, but not bigger than the screen. Since the red `Container` has no size but has a child, it decides it wants to be the same size as its child. The red `Container` tells its child that it can be any size it wants, but not bigger than the screen. The child is a green `Container` that wants to be 30×30. The red `Container` sizes itself to the size of its child, so it is also 30×30. The red color isn't visible because the green `Container` entirely covers all of the red `Container`.

### Example 8: Container with padding

```dart
Center(
  child: Container(
    padding: const EdgeInsets.all(20),
    color: Colors.red,
    child: Container(color: Colors.green, width: 30, height: 30),
  ),
)
```

The red `Container` sizes itself to its children's size, but it takes its own padding into consideration. So it is also 30×30 plus padding. The red color is visible because of the padding, and the green `Container` has the same size as in the previous example.

### Example 9: ConstrainedBox without Center

```dart
ConstrainedBox(
  constraints: const BoxConstraints(
    minWidth: 70,
    minHeight: 70,
    maxWidth: 150,
    maxHeight: 150,
  ),
  child: Container(color: Colors.red, width: 10, height: 10),
)
```

You might guess that the `Container` has to be between 70 and 150 pixels, but you would be wrong. The `ConstrainedBox` only imposes **additional** constraints from those it receives from its parent. Here, the screen forces the `ConstrainedBox` to be exactly the same size as the screen, so it tells its child `Container` to also assume the size of the screen, thus ignoring its 'constraints' parameter.

### Example 10: ConstrainedBox with Center

```dart
Center(
  child: ConstrainedBox(
    constraints: const BoxConstraints(
      minWidth: 70,
      minHeight: 70,
      maxWidth: 150,
      maxHeight: 150,
    ),
    child: Container(color: Colors.red, width: 10, height: 10),
  ),
)
```

Now, `Center` allows `ConstrainedBox` to be any size up to the screen size. The `ConstrainedBox` imposes **additional** constraints from its 'constraints' parameter onto its child. The `Container` must be between 70 and 150 pixels. It wants to have 10 pixels, so it will end up having 70 (the minimum).

### Example 11: ConstrainedBox with large Container

```dart
Center(
  child: ConstrainedBox(
    constraints: const BoxConstraints(
      minWidth: 70,
      minHeight: 70,
      maxWidth: 150,
      maxHeight: 150,
    ),
    child: Container(color: Colors.red, width: 1000, height: 1000),
  ),
)
```

`Center` allows `ConstrainedBox` to be any size up to the screen size. The `ConstrainedBox` imposes **additional** constraints from its 'constraints' parameter onto its child. The `Container` must be between 70 and 150 pixels. It wants to have 1000 pixels, so it ends up having 150 (the maximum).

### Example 12: ConstrainedBox with correct size

```dart
Center(
  child: ConstrainedBox(
    constraints: const BoxConstraints(
      minWidth: 70,
      minHeight: 70,
      maxWidth: 150,
      maxHeight: 150,
    ),
    child: Container(color: Colors.red, width: 100, height: 100),
  ),
)
```

`Center` allows `ConstrainedBox` to be any size up to the screen size. `ConstrainedBox` imposes **additional** constraints from its 'constraints' parameter onto its child. The `Container` must be between 70 and 150 pixels. It wants to have 100 pixels, and that's the size it has, since that's between 70 and 150.

### Example 13: UnconstrainedBox

```dart
UnconstrainedBox(
  child: Container(color: Colors.red, width: 20, height: 50),
)
```

The screen forces the `UnconstrainedBox` to be exactly the same size as the screen. However, the `UnconstrainedBox` lets its child `Container` be any size it wants.

### Example 14: UnconstrainedBox with overflow

```dart
UnconstrainedBox(
  child: Container(color: Colors.red, width: 4000, height: 50),
)
```

The screen forces the `UnconstrainedBox` to be exactly the same size as the screen, and `UnconstrainedBox` lets its child `Container` be any size it wants. Unfortunately, in this case `Container` has 4000 pixels of width and is too big to fit in the `UnconstrainedBox`, so the `UnconstrainedBox` displays the much dreaded "overflow warning".

### Example 15: OverflowBox

```dart
OverflowBox(
  minWidth: 0,
  minHeight: 0,
  maxWidth: double.infinity,
  maxHeight: double.infinity,
  child: Container(color: Colors.red, width: 4000, height: 50),
)
```

The screen forces the `OverflowBox` to be exactly the same size as the screen, and `OverflowBox` lets its child `Container` be any size it wants. `OverflowBox` is similar to `UnconstrainedBox`, and difference is that it won't display any warnings if child doesn't fit space. In this case `Container` is 4000 pixels wide, and is too big to fit in the `OverflowBox`, but the `OverflowBox` simply shows as much as it can, with no warnings given.

### Example 16: UnconstrainedBox with infinite Container

```dart
UnconstrainedBox(
  child: Container(color: Colors.red, width: double.infinity, height: 100),
)
```

This won't render anything, and you'll see an error in the console. The `UnconstrainedBox` lets its child be any size it wants, however its child is a `Container` with infinite size. Flutter can't render infinite sizes, so it throws an error with following message: "BoxConstraints forces an infinite width."

### Example 17: LimitedBox

```dart
UnconstrainedBox(
  child: LimitedBox(
    maxWidth: 100,
    child: Container(
      color: Colors.red,
      width: double.infinity,
      height: 100,
    ),
  ),
)
```

Here you won't get an error anymore, because when the `LimitedBox` is given an infinite size by the `UnconstrainedBox`, it passes a maximum width of 100 down to its child. If you swap the `UnconstrainedBox` for a `Center` widget, the `LimitedBox` won't apply its limit anymore (since its limit is only applied when it gets infinite constraints), and the width of the `Container` is allowed to grow past 100.

### Example 18: FittedBox

```dart
FittedBox(child: Text('Some Example Text.'))
```

The screen forces the `FittedBox` to be exactly the same size as the screen. The `Text` has some natural width (also called its intrinsic width) that depends on the amount of text, its font size, and so on. The `FittedBox` lets the `Text` be any size it wants, but after the `Text` tells its size to the `FittedBox`, the `FittedBox` scales the `Text` until it fills all of the available width.

### Example 19: FittedBox in Center

```dart
Center(child: FittedBox(child: Text('Some Example Text.')))
```

The `Center` lets the `FittedBox` be any size it wants, up to the screen size. The `FittedBox` then sizes itself to the `Text`, and lets the `Text` be any size it wants. Since both `FittedBox` and `Text` have the same size, no scaling happens.

### Example 20: FittedBox with large text

```dart
Center(
  child: FittedBox(
    child: Text(
      'This is some very very very large text that is too big to fit a regular screen in a single line.',
    ),
  ),
)
```

`FittedBox` tries to size itself to the `Text`, but it can't be bigger than the screen. It then assumes the screen size, and resizes `Text` so that it fits the screen, too.

### Example 21: Large text without FittedBox

```dart
Center(
  child: Text(
    'This is some very very very large text that is too big to fit a regular screen in a single line.',
  ),
)
```

If you remove the `FittedBox`, the `Text` gets its maximum width from the screen, and breaks the line so that it fits the screen.

### Example 22: FittedBox with unbounded Container

```dart
FittedBox(
  child: Container(
    height: 20,
    width: double.infinity,
  ),
)
```

`FittedBox` can only scale a widget that is BOUNDED (has non-infinite width and height). Otherwise, it won't render anything, and you'll see an error in the console.

### Example 23: Row with text

```dart
Row(
  children: [
    Container(
      color: Colors.red,
      child: const Text('Hello!', style: TextStyle(fontSize: 30)),
    ),
    Container(
      color: Colors.green,
      child: const Text('Goodbye!', style: TextStyle(fontSize: 30)),
    ),
  ],
)
```

The screen forces the `Row` to be exactly the same size as the screen. Just like an `UnconstrainedBox`, the `Row` won't impose any constraints onto its children, and instead lets them be any size they want. The `Row` then puts them side-by-side, and any extra space remains empty.

### Example 24: Row with overflow

```dart
Row(
  children: [
    Container(
      color: Colors.red,
      child: const Text(
        'This is a very long text that won\'t fit the line.',
        style: TextStyle(fontSize: 30),
      ),
    ),
    Container(
      color: Colors.green,
      child: const Text('Goodbye!', style: TextStyle(fontSize: 30)),
    ),
  ],
)
```

Since the `Row` won't impose any constraints onto its children, it's quite possible that children might be too big to fit the available width of `Row`. In this case, just like an `UnconstrainedBox`, the `Row` displays the "overflow warning".

### Example 25: Row with Expanded

```dart
Row(
  children: [
    Expanded(
      child: Center(
        child: Container(
          color: Colors.red,
          child: const Text(
            'This is a very long text that won\'t fit the line.',
            style: TextStyle(fontSize: 30),
          ),
        ),
      ),
    ),
    Container(
      color: Colors.green,
      child: const Text('Goodbye!', style: TextStyle(fontSize: 30)),
    ),
  ],
)
```

When a `Row`'s child is wrapped in an `Expanded` widget, the `Row` won't let this child define its own width anymore. Instead, it defines the `Expanded` width according to the other children, and only then the `Expanded` widget forces the original child to have the `Expanded`'s width. In other words, once you use `Expanded`, the original child's width becomes irrelevant, and is ignored.

### Example 26: Row with two Expanded

```dart
Row(
  children: [
    Expanded(
      child: Container(
        color: Colors.red,
        child: const Text(
          'This is a very long text that won\'t fit the line.',
          style: TextStyle(fontSize: 30),
        ),
      ),
    ),
    Expanded(
      child: Container(
        color: Colors.green,
        child: const Text('Goodbye!', style: TextStyle(fontSize: 30)),
      ),
    ),
  ],
)
```

If all of `Row`'s children are wrapped in `Expanded` widgets, each `Expanded` has a size proportional to its flex parameter, and only then each `Expanded` widget forces its child to have the `Expanded`'s width. In other words, `Expanded` ignores the preferred width of its children.

### Example 27: Row with Flexible

```dart
Row(
  children: [
    Flexible(
      child: Container(
        color: Colors.red,
        child: const Text(
          'This is a very long text that won\'t fit the line.',
          style: TextStyle(fontSize: 30),
        ),
      ),
    ),
    Flexible(
      child: Container(
        color: Colors.green,
        child: const Text('Goodbye!', style: TextStyle(fontSize: 30)),
      ),
    ),
  ],
)
```

The only difference if you use `Flexible` instead of `Expanded`, is that `Flexible` lets its child be SMALLER than the `Flexible` width, while `Expanded` forces its child to have the same width of the `Expanded`. But both `Expanded` and `Flexible` ignore their children's width when sizing themselves.

### Example 28: Scaffold with Column

```dart
Scaffold(
  body: Container(
    color: Colors.blue,
    child: const Column(
      children: [Text('Hello!'), Text('Goodbye!')],
    ),
  ),
)
```

The screen forces the `Scaffold` to be exactly the same size as the screen, so `Scaffold` fills the screen. The `Scaffold` tells the `Container` that it can be any size it wants, but not bigger than the screen. When a widget tells its child that it can be smaller than a certain size, we say the widget supplies "loose" constraints to its child.

### Example 29: Scaffold with expanded Column

```dart
Scaffold(
  body: SizedBox.expand(
    child: Container(
      color: Colors.blue,
      child: const Column(
        children: [Text('Hello!'), Text('Goodbye!')],
      ),
    ),
  ),
)
```

If you want the `Scaffold`'s child to be exactly the same size as the `Scaffold` itself, you can wrap its child with `SizedBox.expand`. When a widget tells its child that it must be of a certain size, we say the widget supplies "tight" constraints to its child.

[Center]: https://api.flutter.dev/flutter/widgets/Center-class.html
[Container]: https://api.flutter.dev/flutter/widgets/Container-class.html
[FittedBox]: https://api.flutter.dev/flutter/widgets/FittedBox-class.html
[Image]: https://api.flutter.dev/flutter/dart-ui/Image-class.html
[ListTile]: https://api.flutter.dev/flutter/material/ListTile-class.html
[ListView]: https://api.flutter.dev/flutter/widgets/ListView-class.html
[Opacity]: https://api.flutter.dev/flutter/widgets/Opacity-class.html
[Row]: https://api.flutter.dev/flutter/widgets/Row-class.html
[Text]: https://api.flutter.dev/flutter/widgets/Text-class.html
[Transform]: https://api.flutter.dev/flutter/widgets/Transform-class.html
