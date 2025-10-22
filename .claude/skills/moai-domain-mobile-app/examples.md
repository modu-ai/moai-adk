# Mobile App Development - Working Examples

> Real-world Flutter and React Native examples

---

## Example 1: Flutter Todo App

### main.dart
```dart
import 'package:flutter/material.dart';

void main() => runApp(MyApp());

class MyApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Todo App',
      theme: ThemeData(primarySwatch: Colors.blue),
      home: TodoList(),
    );
  }
}

class TodoList extends StatefulWidget {
  @override
  _TodoListState createState() => _TodoListState();
}

class _TodoListState extends State<TodoList> {
  List<String> todos = [];

  void _addTodo(String task) {
    setState(() {
      todos.add(task);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Todos')),
      body: ListView.builder(
        itemCount: todos.length,
        itemBuilder: (context, index) {
          return ListTile(title: Text(todos[index]));
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _addTodo('New task'),
        child: Icon(Icons.add),
      ),
    );
  }
}
```

---

**Last Updated**: 2025-10-22
**Framework**: Flutter 3.27.0
