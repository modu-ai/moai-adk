# Layout Basics in Flutter

## Core Concepts

In Flutter, almost everything is a widget—even layout models are widgets. The images, icons, and text that you see in a Flutter app are all widgets. But things you don't see are also widgets, such as the rows, columns, and grids that arrange, constrain, and align the visible widgets. You create a layout by composing widgets to build more complex widgets.

## Lay out a Single Widget

### Select a layout widget

Choose from a variety of layout widgets based on how you want to align or constrain a visible widget, as these characteristics are typically passed on to the contained widget. For example, you could use the [`Center`][] layout widget to center a visible widget horizontally and vertically.

### Create a visible widget

Choose a visible widget for your app to contain visible elements, such as text, images, or icons.

### Add the visible widget to the layout widget

All layout widgets have either of the following:

* A `child` property if they take a single child—for example, `Center` or `Container`
* A `children` property if they take a list of widgets—for example, `Row`, `Column`, `ListView`, or `Stack`.

### Add the layout widget to the page

A Flutter app is itself a widget, and most widgets have a `build()` method. Instantiating and returning a widget in the app's `build()` method displays the widget.

## Lay Out Multiple Widgets Vertically and Horizontally

One of the most common layout patterns is to arrange widgets vertically or horizontally. You can use a `Row` widget to arrange widgets horizontally, and a `Column` widget to arrange widgets vertically.

To create a row or column in Flutter, add a list of children widgets to a [`Row`][] or [`Column`][] widget. In turn, each child can itself be a row or column, and so on.

### Aligning widgets

You control how a row or column aligns its children using the `mainAxisAlignment` and `crossAxisAlignment` properties. For a row, the main axis runs horizontally and the cross axis runs vertically. For a column, the main axis runs vertically and the cross axis runs horizontally.

The [`MainAxisAlignment`][] and [`CrossAxisAlignment`][] enums offer a variety of constants for controlling alignment.

### Sizing widgets

When a layout is too large to fit a device, a yellow and black striped pattern appears along the affected edge. Widgets can be sized to fit within a row or column by using the [`Expanded`][] widget.

### Packing widgets

By default, a row or column occupies as much space along its main axis as possible, but if you want to pack the children closely together, set its `mainAxisSize` to `MainAxisSize.min`.

### Nesting rows and columns

The layout framework allows you to nest rows and columns inside of rows and columns as deeply as you need. To minimize the visual confusion that can result from heavily nested layout code, implement pieces of the UI in variables and functions.

[Center]: https://api.flutter.dev/flutter/widgets/Center-class.html
[Column]: https://api.flutter.dev/flutter/widgets/Column-class.html
[CrossAxisAlignment]: https://api.flutter.dev/flutter/rendering/CrossAxisAlignment.html
[Expanded]: https://api.flutter.dev/flutter/widgets/Expanded-class.html
[MainAxisAlignment]: https://api.flutter.dev/flutter/rendering/MainAxisAlignment.html
[Row]: https://api.flutter.dev/flutter/widgets/Row-class.html
