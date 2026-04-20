# TDD Examples

## Example 1: New API Endpoint

### RED - Write Failing Test

```python
def test_create_user_returns_201_with_valid_data():
    response = client.post("/users", json={
        "email": "test@example.com",
        "name": "Test User"
    })
    assert response.status_code == 201
    assert response.json()["email"] == "test@example.com"
```

### GREEN - Minimal Implementation

```python
@app.post("/users", status_code=201)
def create_user(user: UserCreate):
    return {"email": user.email, "name": user.name}
```

### REFACTOR - Improve Code

```python
@app.post("/users", status_code=201)
def create_user(user: UserCreate, db: Session = Depends(get_db)):
    db_user = User(**user.dict())
    db.add(db_user)
    db.commit()
    return db_user
```

## Example 2: New React Component

### RED - Write Failing Test

```typescript
test('Button displays label text', () => {
  render(<Button label="Click me" />);
  expect(screen.getByText('Click me')).toBeInTheDocument();
});
```

### GREEN - Minimal Implementation

```typescript
function Button({ label }: { label: string }) {
  return <button>{label}</button>;
}
```

### REFACTOR - Add Styling

```typescript
function Button({ label, variant = 'primary' }: ButtonProps) {
  return (
    <button className={`btn btn-${variant}`}>
      {label}
    </button>
  );
}
```

## Example 3: New Utility Function

### RED - Write Failing Test

```go
func TestFormatCurrency(t *testing.T) {
    result := FormatCurrency(1234.56, "USD")
    assert.Equal(t, "$1,234.56", result)
}
```

### GREEN - Minimal Implementation

```go
func FormatCurrency(amount float64, currency string) string {
    return fmt.Sprintf("$%.2f", amount)
}
```

### REFACTOR - Handle Localization

```go
func FormatCurrency(amount float64, currency string) string {
    p := message.NewPrinter(language.English)
    symbol := currencySymbols[currency]
    return p.Sprintf("%s%.2f", symbol, amount)
}
```
