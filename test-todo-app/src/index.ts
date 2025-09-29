// Main entry point for TODO app
export interface Todo {
  id: number;
  text: string;
  completed: boolean;
}

export class TodoManager {
  private todos: Todo[] = [];

  addTodo(text: string): Todo {
    const todo: Todo = {
      id: Date.now(),
      text,
      completed: false
    };
    this.todos.push(todo);
    return todo;
  }

  toggleTodo(id: number): boolean {
    const todo = this.todos.find(t => t.id === id);
    if (todo) {
      todo.completed = !todo.completed;
      return true;
    }
    return false;
  }

  getTodos(): Todo[] {
    return [...this.todos];
  }
}