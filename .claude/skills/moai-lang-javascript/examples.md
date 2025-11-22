# JavaScript - Practical Examples (10 Examples)

## Example 1: Async/Await
```javascript
async function fetchUsers() {
    try {
        const response = await fetch('/api/users');
        const users = await response.json();
        return users;
    } catch (error) {
        console.error('Error:', error);
    }
}
```

## Example 2: Promises
```javascript
Promise.all([
    fetch('/api/users'),
    fetch('/api/posts')
]).then(([usersRes, postsRes]) => {
    return Promise.all([usersRes.json(), postsRes.json()]);
}).then(([users, posts]) => {
    console.log(users, posts);
});
```

## Example 3: Array Methods
```javascript
const numbers = [1, 2, 3, 4, 5];
const doubled = numbers.map(n => n * 2);
const evens = numbers.filter(n => n % 2 === 0);
const sum = numbers.reduce((acc, n) => acc + n, 0);
```

## Example 4: Destructuring
```javascript
const { name, age } = user;
const [first, second, ...rest] = array;
```

## Example 5: Spread Operator
```javascript
const newArray = [...oldArray, newItem];
const newObject = { ...oldObject, updated: true };
```

## Example 6: Classes
```javascript
class User {
    constructor(name, email) {
        this.name = name;
        this.email = email;
    }
    
    greet() {
        return `Hello, ${this.name}`;
    }
}
```

## Example 7: Modules
```javascript
// export
export const add = (a, b) => a + b;
export default User;

// import
import User, { add } from './utils.js';
```

## Example 8: Error Handling
```javascript
try {
    JSON.parse(invalidJSON);
} catch (error) {
    console.error('Parse error:', error.message);
} finally {
    cleanup();
}
```

## Example 9: Event Listeners
```javascript
document.querySelector('#btn').addEventListener('click', (e) => {
    e.preventDefault();
    handleClick();
});
```

## Example 10: Fetch API
```javascript
fetch('/api/data', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ key: 'value' })
})
.then(res => res.json())
.then(data => console.log(data));
```
