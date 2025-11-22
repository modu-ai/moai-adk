# Scala Performance Optimization

_Last updated: 2025-11-22 | Enterprise Production Ready_

---

## JVM Tuning for Scala

### JVM Heap Configuration

```bash
# Development
export JVM_OPTS="-Xms1G -Xmx4G -XX:+UseG1GC"

# Production
export JVM_OPTS="-Xms8G -Xmx16G -XX:+UseG1GC -XX:MaxGCPauseMillis=200"
```

### Garbage Collection Optimization

```bash
# G1GC configuration for low latency
java -Xms8G -Xmx16G \
  -XX:+UseG1GC \
  -XX:MaxGCPauseMillis=200 \
  -XX:InitiatingHeapOccupancyPercent=35 \
  -XX:G1HeapRegionSize=16M \
  -XX:MinMetaspaceSize=512M \
  -XX:MaxMetaspaceSize=2G \
  -XX:+PrintGCDetails \
  -XX:+PrintGCTimeStamps \
  -Xloggc:gc.log
```

---

## Code-Level Optimization

### Value Classes for Zero-Cost Abstraction

```scala
case class UserId(value: Int) extends AnyVal {
  def toString: String = s"User($value)"
}

// Compiled away - no object allocation
val id = UserId(42)
val extracted = id.value  // Direct access
```

### Specialization for Generic Types

```scala
@specialized trait Functor[@specialized F[_]] {
  def map[@specialized A, @specialized B](fa: F[A])(f: A => B): F[B]
}

// Generates specialized versions for primitive types
object IntFunctor extends Functor[Option] {
  def map[A, B](fa: Option[A])(f: A => B): Option[B] = fa.map(f)
}
```

### Tail Recursion

```scala
// Good: Tail recursive (optimized by compiler)
@scala.annotation.tailrec
def factorial(n: Int, acc: Int = 1): Int = {
  if (n <= 1) acc else factorial(n - 1, acc * n)
}

// Bad: Not tail recursive
def sum(n: Int): Int = {
  if (n <= 0) 0 else n + sum(n - 1)
}
```

---

## Collection Performance

### Choosing the Right Collection

```scala
// For frequent lookups
val map: scala.collection.Map[String, Int] = Map("a" -> 1, "b" -> 2)

// For sequential access
val list: List[String] = List("a", "b", "c")

// For set operations
val set: scala.collection.Set[Int] = Set(1, 2, 3)

// For mutable operations
val buffer: scala.collection.mutable.Buffer[String] = scala.collection.mutable.ArrayBuffer()
buffer += "item"
```

### Efficient Traversal

```scala
// Good: Single pass, lazy evaluation
val result = (1 to 1000000)
  .filter(_ % 2 == 0)
  .map(_ * 2)
  .sum

// Better: Use view for lazy evaluation
val lazyResult = (1 to 1000000)
  .view
  .filter(_ % 2 == 0)
  .map(_ * 2)
  .sum
```

---

## Async Performance

### ZIO Fiber Tuning

```scala
import zio.*

def tuneZIOPerformance: ZIO[Any, Nothing, Unit] = {
  // Configure fiber thread pool
  val config = ZIOAppConfig.default.copy(
    blockingExecutor = Executor.makeBlockingExecutor(100)
  )

  // Limit concurrent fibers
  val limitedFibers = for {
    semaphore <- Semaphore.make(100)
    result <- ZIO.foreachParDiscard((1 to 1000).toList) { i =>
      semaphore.withPermit(computeTask(i))
    }
  } yield result

  limitedFibers
}
```

### Play Framework Performance

```scala
// Configure thread pool in application.conf
play {
  akka {
    actor-system = "play"
    server {
      netty {
        transport = "native"
        max-chunk-size = "256kb"
        idle-timeout = 60s
      }
    }
  }

  server {
    provider = "play.core.server.NettyServerProvider"
    http {
      port = 9000
      idleTimeout = 60s
    }
  }
}
```

---

## Profiling Scala Applications

### Using JFR (Java Flight Recorder)

```bash
# Record JFR data
java -XX:+UnlockCommercialFeatures -XX:+FlightRecorder \
  -XX:StartFlightRecording=duration=60s,filename=myapp.jfr \
  -jar myapp.jar

# Analyze with jfr command
jfr print myapp.jfr
jfr dump --output analysis.html myapp.jfr
```

### Using async-profiler

```bash
# Install async-profiler
curl https://github.com/async-profiler/async-profiler/releases/download/v3.0/async-profiler-3.0-linux-x64.tar.gz -o async-profiler.tar.gz
tar xzf async-profiler.tar.gz

# Profile running JVM process
./profiler.sh -d 30 -e alloc -f result.html 12345

# Analyze hot paths
./profiler.sh -d 60 -e cpu -f result.txt 12345
```

---

## Compilation Performance

### Scala Compilation Tuning

```bash
# Incremental compilation
sbt -Dcompile.incremental=true

# Parallel compilation
sbt -Dcompile.parallel=true -J-Xmx4g

# Faster compilation with SBT
sbt -java-home /path/to/jdk
```

### sbt Configuration

```scala
// build.sbt
ThisBuild / scalaVersion := "3.6.0"
ThisBuild / parallelExecution := true
ThisBuild / maxErrors := 5

// Incremental compiler options
scalacOptions ++= Seq(
  "-release", "17",
  "-opt:l:inline",
  "-opt-inline-from:scala/**"
)

// JAR optimization
assembly / assemblyMergeStrategy := {
  case PathContentConflict(_, xs) =>
    xs.last
  case _ => MergeStrategy.first
}
```

---

## Memory Optimization

### Reducing Memory Footprint

```scala
// Use lazy evaluation to defer allocation
lazy val expensiveComputation: List[Int] = {
  (1 to 1000000).toList
}

// Use iterators instead of collections
def processLargeFile(path: String): Iterator[String] = {
  scala.io.Source.fromFile(path).getLines()
}

// Use streaming for large data
import zio.*
import zio.stream.*

def streamLargeFile(path: String): ZStream[Any, Throwable, String] = {
  ZStream.fromFileName(path)
    .via(ZPipeline.splitLines)
}
```

### Object Pooling

```scala
object ConnectionPool {
  private val pool = scala.collection.mutable.Queue[Connection]()

  def borrow: Connection = {
    if (pool.isEmpty) createConnection()
    else pool.dequeue()
  }

  def return(conn: Connection): Unit = {
    if (pool.size < MAX_POOL_SIZE) {
      pool.enqueue(conn)
    } else {
      conn.close()
    }
  }

  private def createConnection(): Connection = {
    new Connection(/* ... */)
  }
}
```

---

## Monitoring Performance

### Metrics Collection with ZIO

```scala
import zio.*
import zio.metrics.*

def captureMetrics[A](operation: ZIO[Any, Any, A]): ZIO[Any, Any, A] = {
  val startTime = System.nanoTime()

  operation.ensuring {
    val endTime = System.nanoTime()
    val duration = (endTime - startTime) / 1_000_000.0  // milliseconds

    ZIO.succeed(println(s"Operation took ${duration}ms"))
  }
}
```

### JVM Metrics

```scala
import java.lang.management.*

object JVMMetrics {
  def heapUsage: Long = {
    val bean = ManagementFactory.getMemoryMXBean
    bean.getHeapMemoryUsage.getUsed
  }

  def gcCount: Long = {
    ManagementFactory.getGarbageCollectorMXBeans
      .map(_.getCollectionCount)
      .sum
  }

  def threadCount: Int = {
    ManagementFactory.getThreadMXBean.getThreadCount
  }
}
```

---

## Best Practices

1. **Profile before optimizing**: Use JFR and async-profiler
2. **Prefer immutable collections**: Safer and optimized for sharing
3. **Use lazy evaluation**: Defer computation with lazy and view
4. **Configure JVM properly**: Tune heap, GC, and thread pools
5. **Monitor in production**: Collect metrics and analyze patterns
6. **Use value classes**: For zero-cost abstractions
7. **Limit concurrency**: Use semaphores to prevent resource exhaustion

---

**Status**: Production Ready | **Updated**: 2025-11-22
