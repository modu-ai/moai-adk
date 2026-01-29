# Akka Typed Actors 2.9

Comprehensive coverage of Akka Typed Actors 2.9 including Actors, Streams, and Clustering patterns.

## Context7 Library Mappings

```
/akka/akka - Akka 2.9 typed actors and streams
/akka/akka-http - Akka HTTP REST APIs
/akka/alpakka - Akka connectors
/akka/akka-persistence - Event sourcing
/akka/akka-cluster - Clustering and sharding
```

---

## Typed Actor Basics

### Defining Behaviors

```scala
import akka.actor.typed.*
import akka.actor.typed.scaladsl.*

object Counter:
  // Message protocol
  sealed trait Command
  case class Increment(amount: Int) extends Command
  case class Decrement(amount: Int) extends Command
  case class GetCount(replyTo: ActorRef[Int]) extends Command
  case object Reset extends Command

  // Behavior definition
  def apply(count: Int = 0): Behavior[Command] =
    Behaviors.receiveMessage {
      case Increment(amount) =>
        Counter(count + amount)
      case Decrement(amount) =>
        Counter(count - amount)
      case GetCount(replyTo) =>
        replyTo ! count
        Behaviors.same
      case Reset =>
        Counter(0)
    }
```

### Using Behaviors.receive

```scala
object UserActor:
  sealed trait Command
  case class GetUser(id: Long, replyTo: ActorRef[Option[User]]) extends Command
  case class CreateUser(request: CreateUserRequest, replyTo: ActorRef[User]) extends Command
  case class UpdateUser(id: Long, name: String, replyTo: ActorRef[Option[User]]) extends Command

  def apply(repository: UserRepository): Behavior[Command] =
    Behaviors.receive { (context, message) =>
      message match
        case GetUser(id, replyTo) =>
          replyTo ! repository.findById(id)
          Behaviors.same
        
        case CreateUser(request, replyTo) =>
          val user = repository.save(User.from(request))
          context.log.info(s"Created user: ${user.id}")
          replyTo ! user
          Behaviors.same
        
        case UpdateUser(id, name, replyTo) =>
          val updated = repository.findById(id).map { user =>
            repository.save(user.copy(name = name))
          }
          replyTo ! updated
          Behaviors.same
    }
```

### Stateful Behaviors

```scala
object ShoppingCart:
  sealed trait Command
  case class AddItem(item: Item, replyTo: ActorRef[Cart]) extends Command
  case class RemoveItem(itemId: String, replyTo: ActorRef[Cart]) extends Command
  case class GetCart(replyTo: ActorRef[Cart]) extends Command
  case class Checkout(replyTo: ActorRef[Order]) extends Command

  case class Cart(items: List[Item] = Nil):
    def add(item: Item): Cart = copy(items = item :: items)
    def remove(id: String): Cart = copy(items = items.filterNot(_.id == id))
    def total: BigDecimal = items.map(_.price).sum

  def apply(): Behavior[Command] = active(Cart())

  private def active(cart: Cart): Behavior[Command] =
    Behaviors.receiveMessage {
      case AddItem(item, replyTo) =>
        val newCart = cart.add(item)
        replyTo ! newCart
        active(newCart)
      
      case RemoveItem(itemId, replyTo) =>
        val newCart = cart.remove(itemId)
        replyTo ! newCart
        active(newCart)
      
      case GetCart(replyTo) =>
        replyTo ! cart
        Behaviors.same
      
      case Checkout(replyTo) =>
        val order = Order.from(cart)
        replyTo ! order
        empty()
    }

  private def empty(): Behavior[Command] = apply()
```

---

## Actor Lifecycle and Supervision

### Lifecycle Signals

```scala
def apply(): Behavior[Command] =
  Behaviors.setup { context =>
    context.log.info("Actor starting")
    
    Behaviors.receiveMessage[Command] { msg =>
      // Handle message
      Behaviors.same
    }.receiveSignal {
      case (ctx, PreRestart) =>
        ctx.log.info("Actor restarting")
        Behaviors.same
      case (ctx, PostStop) =>
        ctx.log.info("Actor stopped")
        Behaviors.same
    }
  }
```

### Supervision Strategies

```scala
import akka.actor.typed.SupervisorStrategy
import scala.concurrent.duration.*

def supervisedBehavior: Behavior[Command] =
  Behaviors.supervise(Counter())
    .onFailure[IllegalArgumentException](SupervisorStrategy.resume)

def restartWithBackoff: Behavior[Command] =
  Behaviors.supervise(Counter())
    .onFailure[RuntimeException](
      SupervisorStrategy.restartWithBackoff(
        minBackoff = 1.second,
        maxBackoff = 30.seconds,
        randomFactor = 0.2
      ).withMaxRestarts(10)
    )

def stopOnFailure: Behavior[Command] =
  Behaviors.supervise(Counter())
    .onFailure[Exception](SupervisorStrategy.stop)
```

### Child Actors

```scala
object Parent:
  sealed trait Command
  case class CreateChild(name: String) extends Command
  case class MessageChild(name: String, msg: Child.Command) extends Command

  def apply(): Behavior[Command] =
    Behaviors.setup { context =>
      var children = Map.empty[String, ActorRef[Child.Command]]

      Behaviors.receiveMessage {
        case CreateChild(name) =>
          val child = context.spawn(Child(), name)
          context.watch(child)
          children += (name -> child)
          Behaviors.same
        
        case MessageChild(name, msg) =>
          children.get(name).foreach(_ ! msg)
          Behaviors.same
      }.receiveSignal {
        case (ctx, Terminated(ref)) =>
          ctx.log.info(s"Child ${ref.path.name} terminated")
          children -= ref.path.name
          Behaviors.same
      }
    }
```

---

## Akka Streams

### Basic Stream Operations

```scala
import akka.stream.*
import akka.stream.scaladsl.*
import akka.NotUsed

given system: ActorSystem[Nothing] = ???

// Source, Flow, Sink
val source: Source[Int, NotUsed] = Source(1 to 1000)
val flow: Flow[Int, String, NotUsed] = 
  Flow[Int].filter(_ % 2 == 0).map(_ * 2).map(_.toString)
val sink: Sink[String, Future[Done]] = Sink.foreach(println)

// Connecting components
val graph: RunnableGraph[Future[Done]] = 
  source.via(flow).toMat(sink)(Keep.right)

val result: Future[Done] = graph.run()
```

### Stream Sources

```scala
// From collections
val fromList: Source[Int, NotUsed] = Source(List(1, 2, 3))
val fromRange: Source[Int, NotUsed] = Source(1 to 100)

// Infinite sources
val ticks: Source[Long, Cancellable] = 
  Source.tick(1.second, 1.second, 0L).map(_ => System.currentTimeMillis())

val repeated: Source[String, NotUsed] = Source.repeat("hello").take(10)

// From Future
val fromFuture: Source[User, NotUsed] = 
  Source.future(fetchUser(1L))

// From actor
val fromActor: Source[Message, ActorRef[Message]] =
  ActorSource.actorRef[Message](
    completionMatcher = { case Complete => },
    failureMatcher = { case Failure(ex) => ex },
    bufferSize = 100,
    overflowStrategy = OverflowStrategy.dropHead
  )
```

### Backpressure and Buffering

```scala
// Throttling
val throttled: Source[Int, NotUsed] = source
  .throttle(100, 1.second)

// Buffering
val buffered: Source[Int, NotUsed] = source
  .buffer(1000, OverflowStrategy.backpressure)

// Conflate (combine on slow downstream)
val conflated: Source[Int, NotUsed] = source
  .conflate(_ + _)

// Expand (repeat on fast downstream)
val expanded: Source[Int, NotUsed] = source
  .expand(Iterator.continually(_))

// Batch
val batched: Source[Seq[Int], NotUsed] = source
  .groupedWithin(100, 5.seconds)
```

### Parallel Processing

```scala
// Parallel map
val parallel: Source[Result, NotUsed] = source
  .mapAsync(4)(processAsync)

// Unordered parallel
val unordered: Source[Result, NotUsed] = source
  .mapAsyncUnordered(4)(processAsync)

// Fan-out/Fan-in
val fanOut: Source[Int, NotUsed] = source
  .via(
    Flow[Int].flatMapMerge(4, n => Source.single(n).via(processFlow))
  )

// Partition and merge
val graph = GraphDSL.create() { implicit builder =>
  import GraphDSL.Implicits.*
  
  val partition = builder.add(Partition[Int](2, n => if n % 2 == 0 then 0 else 1))
  val merge = builder.add(Merge[String](2))
  
  partition.out(0) ~> evenFlow ~> merge.in(0)
  partition.out(1) ~> oddFlow ~> merge.in(1)
  
  FlowShape(partition.in, merge.out)
}
```

---

## Event Sourcing with Akka Persistence

### Event Sourced Behavior

```scala
import akka.persistence.typed.*
import akka.persistence.typed.scaladsl.*

object UserAggregate:
  // Commands
  sealed trait Command
  case class CreateUser(name: String, email: String, replyTo: ActorRef[Response]) extends Command
  case class UpdateEmail(email: String, replyTo: ActorRef[Response]) extends Command
  case class GetState(replyTo: ActorRef[Option[User]]) extends Command

  // Events
  sealed trait Event
  case class UserCreated(id: String, name: String, email: String, at: Instant) extends Event
  case class EmailUpdated(email: String, at: Instant) extends Event

  // Response
  sealed trait Response
  case class Success(user: User) extends Response
  case class Failure(reason: String) extends Response

  // State
  case class User(id: String, name: String, email: String, createdAt: Instant)

  def apply(id: String): Behavior[Command] =
    EventSourcedBehavior[Command, Event, Option[User]](
      persistenceId = PersistenceId("User", id),
      emptyState = None,
      commandHandler = commandHandler(id),
      eventHandler = eventHandler
    ).withRetention(RetentionCriteria.snapshotEvery(100, 2))

  private def commandHandler(id: String)(state: Option[User], cmd: Command): Effect[Event, Option[User]] =
    state match
      case None => handleNew(id, cmd)
      case Some(user) => handleExisting(user, cmd)

  private def handleNew(id: String, cmd: Command): Effect[Event, Option[User]] =
    cmd match
      case CreateUser(name, email, replyTo) =>
        val event = UserCreated(id, name, email, Instant.now)
        Effect.persist(event).thenRun(state => replyTo ! Success(state.get))
      case GetState(replyTo) =>
        replyTo ! None
        Effect.none
      case _ =>
        Effect.none

  private def handleExisting(user: User, cmd: Command): Effect[Event, Option[User]] =
    cmd match
      case UpdateEmail(email, replyTo) =>
        Effect.persist(EmailUpdated(email, Instant.now))
          .thenRun(state => replyTo ! Success(state.get))
      case GetState(replyTo) =>
        replyTo ! Some(user)
        Effect.none
      case CreateUser(_, _, replyTo) =>
        replyTo ! Failure("User already exists")
        Effect.none

  private val eventHandler: (Option[User], Event) => Option[User] = (state, event) =>
    event match
      case UserCreated(id, name, email, at) =>
        Some(User(id, name, email, at))
      case EmailUpdated(email, _) =>
        state.map(_.copy(email = email))
```

---

## Cluster Patterns

### Cluster Sharding

```scala
import akka.cluster.sharding.typed.scaladsl.*

val TypeKey = EntityTypeKey[UserAggregate.Command]("User")

val sharding = ClusterSharding(system)

val shardRegion: ActorRef[ShardingEnvelope[UserAggregate.Command]] =
  sharding.init(Entity(TypeKey)(ctx => UserAggregate(ctx.entityId)))

// Send message to entity
shardRegion ! ShardingEnvelope("user-123", UserAggregate.GetState(replyTo))
```

### Cluster Singleton

```scala
import akka.cluster.singleton.typed.scaladsl.*

val singletonManager = ClusterSingleton(system)

val proxy: ActorRef[LeaderActor.Command] = singletonManager.init(
  SingletonActor(
    Behaviors.supervise(LeaderActor()).onFailure(SupervisorStrategy.restart),
    "leader"
  )
)
```

---

## Testing

### Actor TestKit

```scala
import akka.actor.testkit.typed.scaladsl.*

class UserActorSpec extends ScalaTestWithActorTestKit with AnyWordSpecLike:
  "UserActor" should {
    "return user when found" in {
      val probe = createTestProbe[Option[User]]()
      val actor = spawn(UserActor(mockRepository))
      
      actor ! UserActor.GetUser(1L, probe.ref)
      
      probe.expectMessage(Some(User(1L, "John", "john@example.com")))
    }
    
    "handle concurrent requests" in {
      val actor = spawn(UserActor(mockRepository))
      val probes = (1 to 100).map(_ => createTestProbe[Option[User]]())
      
      probes.zipWithIndex.foreach { case (probe, i) =>
        actor ! UserActor.GetUser(i.toLong, probe.ref)
      }
      
      probes.foreach(_.expectMessageType[Option[User]])
    }
  }
```

---

## Best Practices

Actor Design:
- Keep behaviors pure and stateless where possible
- Use typed protocols for compile-time safety
- Prefer ask pattern over storing replyTo in state

Supervision:
- Define explicit supervision strategies
- Use restartWithBackoff for transient failures
- Log failures before restarting

Streams:
- Use async boundaries for parallel stages
- Prefer mapAsync over blocking operations
- Monitor stream backpressure with metrics

Persistence:
- Keep events small and immutable
- Use snapshots for large aggregates
- Version events for schema evolution

---

Last Updated: 2026-01-06
Version: 2.0.0
