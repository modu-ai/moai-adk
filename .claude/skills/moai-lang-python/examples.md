# Python - Practical Examples (10 Examples)

## Example 1: List Comprehensions
```python
squares = [x**2 for x in range(10)]
evens = [x for x in range(20) if x % 2 == 0]
```

## Example 2: Generators
```python
def fibonacci():
    a, b = 0, 1
    while True:
        yield a
        a, b = b, a + b

gen = fibonacci()
print(next(gen))  # 0
print(next(gen))  # 1
```

## Example 3: Decorators
```python
def log(func):
    def wrapper(*args, **kwargs):
        print(f"Calling {func.__name__}")
        return func(*args, **kwargs)
    return wrapper

@log
def greet(name):
    return f"Hello, {name}"
```

## Example 4: Context Managers
```python
with open('file.txt', 'r') as f:
    content = f.read()

class DatabaseConnection:
    def __enter__(self):
        self.conn = connect()
        return self.conn
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        self.conn.close()
```

## Example 5: Async/Await
```python
import asyncio

async def fetch_data(url):
    async with aiohttp.ClientSession() as session:
        async with session.get(url) as response:
            return await response.json()

asyncio.run(fetch_data('https://api.example.com'))
```

## Example 6: Type Hints
```python
def greet(name: str) -> str:
    return f"Hello, {name}"

from typing import List, Dict, Optional

def process(items: List[int]) -> Dict[str, int]:
    return {"count": len(items)}
```

## Example 7: Dataclasses
```python
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    email: str
    
user = User(1, "John", "john@example.com")
```

## Example 8: FastAPI
```python
from fastapi import FastAPI

app = FastAPI()

@app.get("/users/{user_id}")
async def get_user(user_id: int):
    return {"id": user_id, "name": "John"}
```

## Example 9: Pydantic
```python
from pydantic import BaseModel, EmailStr

class User(BaseModel):
    name: str
    email: EmailStr
    age: int

user = User(name="John", email="john@example.com", age=30)
```

## Example 10: Pandas
```python
import pandas as pd

df = pd.DataFrame({'A': [1, 2, 3], 'B': [4, 5, 6]})
filtered = df[df['A'] > 1]
grouped = df.groupby('A').sum()
```
