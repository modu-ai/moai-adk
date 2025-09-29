import { TodoManager } from '../src/index';

describe('TodoManager', () => {
  let todoManager: TodoManager;

  beforeEach(() => {
    todoManager = new TodoManager();
  });

  test('should add a todo', () => {
    const todo = todoManager.addTodo('Test todo');
    expect(todo.text).toBe('Test todo');
    expect(todo.completed).toBe(false);
    expect(todo.id).toBeDefined();
  });

  test('should toggle todo completion', () => {
    const todo = todoManager.addTodo('Test todo');
    todoManager.toggleTodo(todo.id);

    const todos = todoManager.getTodos();
    expect(todos[0].completed).toBe(true);
  });

  test('should return all todos', () => {
    todoManager.addTodo('Todo 1');
    todoManager.addTodo('Todo 2');

    const todos = todoManager.getTodos();
    expect(todos.length).toBe(2);
  });
});