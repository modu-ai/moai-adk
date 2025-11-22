# Python Advanced Patterns

## Metaclasses
```python
class Meta(type):
    def __new__(cls, name, bases, attrs):
        attrs['added'] = True
        return super().__new__(cls, name, bases, attrs)

class MyClass(metaclass=Meta):
    pass
```

## Descriptors
```python
class Positive:
    def __set_name__(self, owner, name):
        self.name = name
    
    def __get__(self, obj, objtype=None):
        return obj.__dict__.get(self.name, 0)
    
    def __set__(self, obj, value):
        if value < 0:
            raise ValueError("Must be positive")
        obj.__dict__[self.name] = value
```

## Context Variables
```python
from contextvars import ContextVar

request_id = ContextVar('request_id')

async def process():
    req_id = request_id.get()
    # Use req_id
```

## Weak References
```python
import weakref

cache = weakref.WeakValueDictionary()
cache[key] = obj
```
