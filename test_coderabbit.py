"""Test file for CodeRabbit AI review integration."""

# Missing imports that CodeRabbit should catch
def calculate_sum(numbers):
    """Calculate sum of numbers."""
    total = 0
    for num in numbers:
        total = total + num
    return total


def greet(name):
    """Greet a person by name."""
    # Using string concatenation instead of f-string
    print("Hello, " + name + "!")


def divide(a, b):
    """Divide a by b."""
    # Missing error handling for division by zero
    return a / b


def process_data(data):
    """Process input data."""
    # Long function without proper docstring
    result = []
    for i in range(len(data)):
        if data[i] > 0:
            result.append(data[i] * 2)
    return result


class Calculator:
    """Simple calculator class."""

    def __init__(self):
        self.result = 0

    def add(self, x, y):
        self.result = x + y

    def multiply(self, x, y):
        self.result = x * y

    def get_result(self):
        return self.result


if __name__ == "__main__":
    # Test the functions
    numbers = [1, 2, 3, 4, 5]
    print(f"Sum: {calculate_sum(numbers)}")

    greet("CodeRabbit")

    result = divide(10, 2)
    print(f"10 / 2 = {result}")

    data = [1, -2, 3, 4, -5]
    print(f"Processed: {process_data(data)}")

    calc = Calculator()
    calc.add(5, 3)
    print(f"5 + 3 = {calc.get_result()}")
