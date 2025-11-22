# Scala 3 Advanced Patterns

_Last updated: 2025-11-22 | Enterprise Production Ready_

---

## ZIO Effect System

### Basic ZIO Program

```scala
import zio.*

object HelloWorld extends ZIOAppDefault {
  def run = console.printLine("Hello, World!")
}

// Running ZIO effects
val program: ZIO[Any, IOException, Unit] =
  console.printLine("Enter your name:") *>
    console.readLine.flatMap { name =>
      console.printLine(s"Hello, $name!")
    }
```

### Error Handling with ZIO

```scala
import zio.*

def divide(a: Int, b: Int): ZIO[Any, String, Int] = {
  if (b == 0) ZIO.fail("Division by zero")
  else ZIO.succeed(a / b)
}

// Handle errors
val result: ZIO[Any, Nothing, Int] =
  divide(10, 2).catchAll { error =>
    console.printLine(s"Error: $error") *> ZIO.succeed(0)
  }
```

### Resource Management

```scala
import zio.*
import java.io.*

def openFile(path: String): ZIO[Any, IOException, BufferedReader] = {
  ZIO.succeed(new BufferedReader(new FileReader(path)))
}

def closeFile(reader: BufferedReader): ZIO[Any, IOException, Unit] = {
  ZIO.succeed(reader.close())
}

// Automatic resource management
val program: ZIO[Any, IOException, String] = {
  ZIO.acquireReleaseWith(openFile("file.txt"))(closeFile) { reader =>
    ZIO.succeed(reader.readLine())
  }
}
```

### Concurrent Operations

```scala
import zio.*
import zio.concurrent.*

def fetch(id: Int): ZIO[Any, Nothing, String] = {
  ZIO.sleep(Duration.fromSeconds(1)) *> ZIO.succeed(s"Data-$id")
}

// Run in parallel
val parallel: ZIO[Any, Nothing, List[String]] = {
  ZIO.collectAllPar(
    (1 to 5).map(fetch).toList
  )
}
```

---

## Cats Effect System

### Basic Effect

```scala
import cats.effect.*
import cats.effect.IO

def hello: IO[Unit] = {
  IO.println("Hello from Cats Effect")
}

def greet(name: String): IO[Unit] = {
  IO.print(s"What is your name? ") *>
    IO.println(name)
}
```

### Resource Management with Cats

```scala
import cats.effect.*
import cats.effect.Resource

def fileResource(path: String): Resource[IO, java.nio.file.Files] = {
  Resource.make(
    IO(new java.nio.file.Files)
  )(_ => IO.unit)
}

fileResource("data.txt").use { file =>
  IO.println("Using file")
}
```

### Concurrent Effects

```scala
import cats.effect.*
import cats.effect.std.Queue
import scala.concurrent.duration.*

def producer(queue: Queue[IO, Int]): IO[Unit] = {
  (1 to 10).toList.traverse { i =>
    IO.sleep(100.millis) *> queue.offer(i)
  }.void
}

def consumer(queue: Queue[IO, Int]): IO[Unit] = {
  queue.take.flatMap { value =>
    IO.println(s"Got: $value") *> consumer(queue)
  }
}
```

---

## Type-Level Programming

### Generics and Type Bounds

```scala
// Type bounds
def max[T: Ordering](a: T, b: T): T = {
  val ord = summon[Ordering[T]]
  if (ord.gt(a, b)) a else b
}

// Using given instances
val maxInt = max(5, 10)
val maxStr = max("apple", "zebra")
```

### Context Bounds

```scala
// Using context bounds
def process[T: Show](value: T): String = {
  val shower = summon[Show[T]]
  shower.show(value)
}

trait Show[T] {
  def show(t: T): String
}

given Show[Int] with {
  def show(i: Int) = s"Int($i)"
}
```

### Higher-Kinded Types

```scala
// Higher-kinded type example
trait Functor[F[_]] {
  def map[A, B](fa: F[A])(f: A => B): F[B]
}

given Functor[List] with {
  def map[A, B](fa: List[A])(f: A => B): List[B] = fa.map(f)
}

given Functor[Option] with {
  def map[A, B](fa: Option[A])(f: A => B): Option[B] = fa.map(f)
}
```

---

## Functional Abstractions

### Monoid Pattern

```scala
trait Monoid[T] {
  def empty: T
  def combine(a: T, b: T): T
}

given Monoid[Int] with {
  def empty = 0
  def combine(a: Int, b: Int) = a + b
}

given Monoid[String] with {
  def empty = ""
  def combine(a: String, b: String) = a + b
}

def fold[T](values: List[T])(using m: Monoid[T]): T = {
  values.foldLeft(m.empty)(m.combine)
}
```

### Monad Pattern

```scala
trait Monad[M[_]] {
  def pure[A](a: A): M[A]
  def flatMap[A, B](ma: M[A])(f: A => M[B]): M[B]
}

given Monad[Option] with {
  def pure[A](a: A) = Some(a)
  def flatMap[A, B](ma: Option[A])(f: A => Option[B]) = ma.flatMap(f)
}

// Usage
val result = for {
  x <- Some(5)
  y <- Some(10)
} yield x + y
```

---

## Pattern Matching

### Exhaustive Pattern Matching

```scala
enum TrafficLight {
  case Red, Yellow, Green
}

def action(light: TrafficLight): String = light match {
  case TrafficLight.Red => "Stop"
  case TrafficLight.Yellow => "Slow down"
  case TrafficLight.Green => "Go"
}
```

### GADT Pattern Matching

```scala
enum Expr[T] {
  case IntVal(n: Int) extends Expr[Int]
  case StrVal(s: String) extends Expr[String]
  case Add(l: Expr[Int], r: Expr[Int]) extends Expr[Int]
}

def eval[T](expr: Expr[T]): T = expr match {
  case Expr.IntVal(n) => n
  case Expr.StrVal(s) => s
  case Expr.Add(l, r) => eval(l) + eval(r)
}
```

---

## DSL Creation

### Simple DSL Example

```scala
class SqlBuilder {
  private var query = ""

  def select(cols: String*): this.type = {
    query = s"SELECT ${cols.mkString(", ")}"
    this
  }

  def from(table: String): this.type = {
    query += s" FROM $table"
    this
  }

  def where(condition: String): this.type = {
    query += s" WHERE $condition"
    this
  }

  def build: String = query
}

val sql = SqlBuilder()
  .select("id", "name")
  .from("users")
  .where("age > 18")
  .build
```

---

## Best Practices

1. **Use ZIO for async**: Leverage ZIO for all concurrent operations
2. **Leverage type system**: Use given/using for implicit parameters
3. **Prefer immutability**: Design data structures as immutable
4. **Use pattern matching**: Exhaust all cases explicitly
5. **Compose effects**: Chain operations with flatMap
6. **Error as values**: Model errors as types, not exceptions
7. **Test with property-based testing**: Use ScalaCheck for exhaustive testing

---

**Status**: Production Ready | **Updated**: 2025-11-22
