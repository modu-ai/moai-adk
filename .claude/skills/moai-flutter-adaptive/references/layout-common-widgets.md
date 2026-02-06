# Common Layout Widgets

## Container

Adds padding, margins, borders, background color, or other decorations to a widget.

**Summary:**
* Add padding, margins, borders
* Change background color or image
* Contains a single child widget, but that child can be a `Row`, `Column`, or even the root of a widget tree

## GridView

Lays widgets out as a two-dimensional list. `GridView` provides two pre-fabricated lists, or you can build your own custom grid. When a `GridView` detects that its contents are too long to fit the render box, it automatically scrolls.

**Summary:**
* Lays widgets out in a grid
* Detects when the column content exceeds the render box and automatically provides scrolling
* Build your own custom grid, or use one of the provided grids:
  * `GridView.count` allows you to specify the number of columns
  * `GridView.extent` allows you to specify the maximum pixel width of a tile

**Example using GridView.extent:**

```dart
Widget _buildGrid() => GridView.extent(
  maxCrossAxisExtent: 150,
  padding: const EdgeInsets.all(4),
  mainAxisSpacing: 4,
  crossAxisSpacing: 4,
  children: _buildGridTileList(30),
);

List<Widget> _buildGridTileList(int count) =>
    List.generate(count, (i) => Image.asset('images/pic$i.jpg'));
```

## ListView

[`ListView`][], a column-like widget, automatically provides scrolling when its content is too long for its render box.

**Summary:**
* A specialized [`Column`][] for organizing a list of boxes
* Can be laid out horizontally or vertically
* Detects when its content won't fit and provides scrolling
* Less configurable than `Column`, but easier to use and supports scrolling

**Example:**

```dart
Widget _buildList() {
  return ListView(
    children: [
      _tile('CineArts at the Empire', '85 W Portal Ave', Icons.theaters),
      _tile('The Castro Theater', '429 Castro St', Icons.theaters),
      _tile('Alamo Drafthouse Cinema', '2550 Mission St', Icons.theaters),
      const Divider(),
      _tile('K\'s Kitchen', '757 Monterey Blvd', Icons.restaurant),
      _tile('Emmy\'s Restaurant', '1923 Ocean Ave', Icons.restaurant),
      _tile('Chaiya Thai Restaurant', '272 Claremont Blvd', Icons.restaurant),
      _tile('La Ciccia', '291 30th St', Icons.restaurant),
    ],
  );
}

ListTile _tile(String title, String subtitle, IconData icon) {
  return ListTile(
    title: Text(
      title,
      style: const TextStyle(fontWeight: FontWeight.w500, fontSize: 20),
    ),
    subtitle: Text(subtitle),
    leading: Icon(icon, color: Colors.blue[500]),
  );
}
```

## Stack

Use [`Stack`][] to arrange widgets on top of a base widget—often an image. The widgets can completely or partially overlap the base widget.

**Summary:**
* Use for widgets that overlap another widget
* The first widget in the list of children is the base widget; subsequent children are overlaid on top of that base widget
* A `Stack`'s content can't scroll
* You can choose to clip children that exceed the render box

**Example:**

```dart
Widget _buildStack() {
  return Stack(
    alignment: const Alignment(0.6, 0.6),
    children: [
      const CircleAvatar(
        backgroundImage: AssetImage('images/pic.jpg'),
        radius: 100,
      ),
      Container(
        decoration: const BoxDecoration(color: Colors.black45),
        child: const Text(
          'Mia B',
          style: TextStyle(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
    ],
  );
}
```

## Card

A [`Card`][], from the [Material library][], contains related nuggets of information and can be composed of almost any widget, but is often used with [`ListTile`][]. `Card` has a single child, but its child can be a column, row, list, grid, or other widget that supports multiple children. By default, a `Card` shrinks its size to 0 by 0 pixels. You can use [`SizedBox`][] to constrain the size of a card.

**Summary:**
* Implements a Material card
* Used for presenting related nuggets of information
* Accepts a single child, but that child can be a `Row`, `Column`, or other widget that holds a list of children
* Displayed with rounded corners and a drop shadow
* A `Card`'s content can't scroll
* From the Material library

**Example:**

```dart
Widget _buildCard() {
  return SizedBox(
    height: 210,
    child: Card(
      child: Column(
        children: [
          ListTile(
            title: const Text(
              '1625 Main Street',
              style: TextStyle(fontWeight: FontWeight.w500),
            ),
            subtitle: const Text('My City, CA 99984'),
            leading: Icon(Icons.restaurant_menu, color: Colors.blue[500]),
          ),
          const Divider(),
          ListTile(
            title: const Text(
              '(408) 555-1212',
              style: TextStyle(fontWeight: FontWeight.w500),
            ),
            leading: Icon(Icons.contact_phone, color: Colors.blue[500]),
          ),
          ListTile(
            title: const Text('costa@example.com'),
            leading: Icon(Icons.contact_mail, color: Colors.blue[500]),
          ),
        ],
      ),
    ),
  );
}
```

## ListTile

Use [`ListTile`][], a specialized row widget from the [Material library][], for an easy way to create a row containing up to 3 lines of text and optional leading and trailing icons. `ListTile` is most commonly used in [`Card`][] or [`ListView`][], but can be used elsewhere.

**Summary:**
* A specialized row that contains up to 3 lines of text and optional icons
* Less configurable than `Row`, but easier to use
* From the Material library

[Card]: https://api.flutter.dev/flutter/material/Card-class.html
[Column]: https://api.flutter.dev/flutter/widgets/Column-class.html
[ListTile]: https://api.flutter.dev/flutter/material/ListTile-class.html
[ListView]: https://api.flutter.dev/flutter/widgets/ListView-class.html
[Material library]: https://api.flutter.dev/flutter/material/material-library.html
[Stack]: https://api.flutter.dev/flutter/widgets/Stack-class.html
[SizedBox]: https://api.flutter.dev/flutter/widgets/SizedBox-class.html
