# Scala 3.4 Functional Programming

Comprehensive coverage of Scala 3.4 functional programming features including Given/Using, Type Classes, Enums, Opaque Types, and Extension Methods.

## Context7 Library Mappings

```
/scala/scala3 - Scala 3.4 language reference
/scala/scala-library - Standard library
/typelevel/cats - Cats 2.10 functional abstractions
```

---

## Extension Methods

Type-safe extensions without implicit classes:

### Basic Extensions

```scala
extension (s: String)
  def words: List[String] = s.split("\\s+").toList
  def truncate(maxLen: Int): String =
    if s.length <= maxLen then s else s.take(maxLen - 3) + "..."
  def isBlank: Boolean = s.trim.isEmpty
  def toSlug: String = s.toLowerCase.replaceAll("\\s+", "-")

// Usage
val text = "Hello World"
text.words      // List("Hello", "World")
text.truncate(8) // "Hello..."
text.toSlug     // "hello-world"
```

### Generic Extensions

```scala
extension [A](list: List[A])
  def second: Option[A] = list.drop(1).headOption
  def penultimate: Option[A] = list.dropRight(1).lastOption
  def intersperse(separator: A): List[A] =
    list.flatMap(a => List(separator, a)).drop(1)

extension [A](opt: Option[A])
  def orElseThrow(msg: => String): A =
    opt.getOrElse(throw new NoSuchElementException(msg))
  def toEither[E](error: => E): Either[E, A] =
    opt.fold(Left(error))(Right(_))
```

### Conditional Extensions

```scala
extension [A](value: A)
  def applyIf(cond: Boolean)(f: A => A): A =
    if cond then f(value) else value
  def let[B](f: A => B): B = f(value)
  def also(f: A => Unit): A = { f(value); value }

// Usage
val result = "hello"
  .applyIf(shouldUppercase)(_.toUpperCase)
  .let(s => s"[$s]")
  .also(println)
```

---

## Given and Using (Context Parameters)

Modern replacement for Scala 2 implicits:

### Type Class Pattern

```scala
trait JsonEncoder[A]:
  def encode(value: A): String

trait JsonDecoder[A]:
  def decode(json: String): Either[String, A]

// Type class instances
given JsonEncoder[String] with
  def encode(value: String): String = s"\"$value\""

given JsonEncoder[Int] with
  def encode(value: Int): String = value.toString

given JsonEncoder[Boolean] with
  def encode(value: Boolean): String = value.toString

// Derived instances
given [A](using encoder: JsonEncoder[A]): JsonEncoder[List[A]] with
  def encode(value: List[A]): String =
    value.map(encoder.encode).mkString("[", ",", "]")

given [A](using encoder: JsonEncoder[A]): JsonEncoder[Option[A]] with
  def encode(value: Option[A]): String =
    value.fold("null")(encoder.encode)

given [K, V](using 
  keyEnc: JsonEncoder[K], 
  valEnc: JsonEncoder[V]
): JsonEncoder[Map[K, V]] with
  def encode(value: Map[K, V]): String =
    val pairs = value.map { case (k, v) =>
      s"${keyEnc.encode(k)}:${valEnc.encode(v)}"
    }
    pairs.mkString("{", ",", "}")
```

### Using Context Functions

```scala
def toJson[A](value: A)(using encoder: JsonEncoder[A]): String =
  encoder.encode(value)

def fromJson[A](json: String)(using decoder: JsonDecoder[A]): Either[String, A] =
  decoder.decode(json)

// Context functions for cleaner APIs
type Encoded[A] = JsonEncoder[A] ?=> String

def encodeAll[A](values: List[A]): Encoded[A] =
  values.map(v => summon[JsonEncoder[A]].encode(v)).mkString(",")

// Usage
val json = toJson(List(1, 2, 3))  // "[1,2,3]"
val encoded = encodeAll(List("a", "b"))  // "\"a\",\"b\""
```

### Given Priority and Imports

```scala
// Import all givens from a companion
import JsonEncoder.given

// Import specific givens
import JsonEncoder.{given JsonEncoder[String]}

// Define priority with explicit types
given lowPriority: JsonEncoder[Any] = ???
given highPriority: JsonEncoder[String] = ???

// Using clauses for multiple contexts
def process[A](value: A)(using 
  enc: JsonEncoder[A],
  ord: Ordering[A],
  show: Show[A]
): String = ???
```

---

## Enum Types and ADTs

Algebraic data types with exhaustive pattern matching:

### Simple Enums

```scala
enum Color:
  case Red, Green, Blue

enum DayOfWeek:
  case Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday
  
  def isWeekend: Boolean = this match
    case Saturday | Sunday => true
    case _ => false
```

### Parameterized Enums

```scala
enum Color(val hex: String):
  case Red extends Color("#FF0000")
  case Green extends Color("#00FF00")
  case Blue extends Color("#0000FF")
  case Custom(override val hex: String) extends Color(hex)

enum HttpStatus(val code: Int, val message: String):
  case Ok extends HttpStatus(200, "OK")
  case Created extends HttpStatus(201, "Created")
  case BadRequest extends HttpStatus(400, "Bad Request")
  case NotFound extends HttpStatus(404, "Not Found")
  case InternalError extends HttpStatus(500, "Internal Server Error")
```

### Generic Enums (ADTs)

```scala
enum Result[+E, +A]:
  case Success(value: A)
  case Failure(error: E)

  def map[B](f: A => B): Result[E, B] = this match
    case Success(a) => Success(f(a))
    case Failure(e) => Failure(e)

  def flatMap[E2 >: E, B](f: A => Result[E2, B]): Result[E2, B] = this match
    case Success(a) => f(a)
    case Failure(e) => Failure(e)

  def fold[B](onFailure: E => B, onSuccess: A => B): B = this match
    case Success(a) => onSuccess(a)
    case Failure(e) => onFailure(e)

enum Tree[+A]:
  case Leaf(value: A)
  case Branch(left: Tree[A], right: Tree[A])

  def map[B](f: A => B): Tree[B] = this match
    case Leaf(a) => Leaf(f(a))
    case Branch(l, r) => Branch(l.map(f), r.map(f))

  def fold[B](onLeaf: A => B)(onBranch: (B, B) => B): B = this match
    case Leaf(a) => onLeaf(a)
    case Branch(l, r) => onBranch(l.fold(onLeaf)(onBranch), r.fold(onLeaf)(onBranch))
```

---

## Opaque Types

Zero-cost type abstractions:

### Basic Opaque Types

```scala
object UserId:
  opaque type UserId = Long
  
  def apply(id: Long): UserId = id
  def fromString(s: String): Option[UserId] = s.toLongOption
  
  extension (id: UserId)
    def value: Long = id
    def asString: String = id.toString

export UserId.UserId
```

### Validated Opaque Types

```scala
object Email:
  opaque type Email = String
  
  private val emailRegex = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$".r
  
  def apply(email: String): Either[String, Email] =
    if emailRegex.matches(email) then Right(email)
    else Left(s"Invalid email format: $email")
  
  def unsafeApply(email: String): Email = email
  
  extension (email: Email)
    def value: String = email
    def domain: String = email.split("@").last
    def localPart: String = email.split("@").head

object NonEmptyString:
  opaque type NonEmptyString = String
  
  def apply(s: String): Option[NonEmptyString] =
    Option.when(s.nonEmpty)(s)
  
  def unsafeApply(s: String): NonEmptyString =
    require(s.nonEmpty, "String must not be empty")
    s
  
  extension (nes: NonEmptyString)
    def value: String = nes
    def length: Int = nes.length
```

### Refined Types Pattern

```scala
object PositiveInt:
  opaque type PositiveInt = Int
  
  def apply(n: Int): Either[String, PositiveInt] =
    if n > 0 then Right(n)
    else Left(s"$n is not positive")
  
  extension (n: PositiveInt)
    def value: Int = n
    def +(other: PositiveInt): PositiveInt = n + other
    def *(other: PositiveInt): PositiveInt = n * other

object Percentage:
  opaque type Percentage = Double
  
  def apply(value: Double): Either[String, Percentage] =
    if value >= 0 && value <= 100 then Right(value)
    else Left(s"$value is not a valid percentage (0-100)")
  
  extension (p: Percentage)
    def value: Double = p
    def asFraction: Double = p / 100.0
```

---

## Union and Intersection Types

### Union Types

```scala
// Type-safe unions without wrapper types
type StringOrInt = String | Int

def describe(value: StringOrInt): String = value match
  case s: String => s"String: $s"
  case i: Int => s"Int: $i"

// Complex unions
type JsonPrimitive = String | Int | Double | Boolean | Null
type JsonValue = JsonPrimitive | List[JsonValue] | Map[String, JsonValue]

def processJson(json: JsonValue): String = json match
  case s: String => s"string: $s"
  case n: Int => s"int: $n"
  case d: Double => s"double: $d"
  case b: Boolean => s"bool: $b"
  case null => "null"
  case l: List[?] => s"array[${l.size}]"
  case m: Map[?, ?] => s"object[${m.size}]"
```

### Intersection Types

```scala
trait HasName:
  def name: String

trait HasAge:
  def age: Int

trait HasEmail:
  def email: String

// Combining traits
type Person = HasName & HasAge
type Contact = HasName & HasEmail
type FullProfile = HasName & HasAge & HasEmail

def greet(person: Person): String =
  s"Hello ${person.name}, age ${person.age}"

def sendEmail(contact: Contact): Unit =
  println(s"Sending to ${contact.email}")

// Implementation
case class User(name: String, age: Int, email: String) 
  extends HasName, HasAge, HasEmail

val user = User("John", 30, "john@example.com")
greet(user)      // Works - User is a Person
sendEmail(user)  // Works - User is a Contact
```

---

## Match Types

Type-level computation and pattern matching:

```scala
// Basic match types
type Elem[X] = X match
  case String => Char
  case Array[t] => t
  case List[t] => t
  case Iterable[t] => t

val charElem: Elem[String] = 'a'
val intElem: Elem[List[Int]] = 42

// Recursive match types
type Flatten[T] = T match
  case List[List[t]] => Flatten[List[t]]
  case List[t] => List[t]
  case t => t

// Head and Tail for tuples
type Head[T <: Tuple] = T match
  case h *: t => h
  case EmptyTuple => Nothing

type Tail[T <: Tuple] = T match
  case h *: t => t
  case EmptyTuple => EmptyTuple

// Concat for tuples
type Concat[T <: Tuple, U <: Tuple] = T match
  case EmptyTuple => U
  case h *: t => h *: Concat[t, U]
```

---

## Inline and Compile-Time Evaluation

```scala
// Inline methods
inline def debug(inline msg: String): Unit =
  if isDebugEnabled then println(s"[DEBUG] $msg")

inline def choose[A](inline cond: Boolean, inline a: A, inline b: A): A =
  if cond then a else b

// Compile-time operations
import scala.compiletime.*

inline def typeNameOf[T]: String = constValue[T].toString

inline def sumValues[T <: Tuple]: Int = inline erasedValue[T] match
  case _: EmptyTuple => 0
  case _: (h *: t) => constValue[h & Int] + sumValues[t]

// Selective compilation
inline def platformCode(): Unit =
  inline if constValue[scala.util.Properties.isWin] then
    println("Windows")
  else
    println("Unix-like")
```

---

## Best Practices

Extension Methods:
- Group related extensions in a single extension block
- Use descriptive names that read naturally with the type
- Consider providing both safe (Option-returning) and unsafe variants

Given/Using:
- Prefer given instances in companion objects for automatic discovery
- Use explicit type annotations for complex derived instances
- Import givens selectively to avoid ambiguity

Enums:
- Use simple enums for finite sets of values
- Use parameterized enums when values carry data
- Prefer pattern matching over isInstanceOf checks

Opaque Types:
- Always provide a smart constructor that validates input
- Export the type alias from the companion object
- Use extensions for all instance methods

---

Last Updated: 2026-01-06
Version: 2.0.0
