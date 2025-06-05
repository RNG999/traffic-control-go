# Coding Agent Custom Instructions

Based on thorough analysis of the provided documentation, the following comprehensive guidelines and instructions are generated for coding tasks across various domains including CQRS, Event Sourcing, Domain-Driven Design (DDD), Functional Programming, Dependency Injection (DI), Testing, and general software design principles. Adhering to these instructions is crucial for developing high-quality, maintainable, and robust software.

## 1. CQRS Principles & Module Segregation

Implement the Command Query Responsibility Segregation (CQRS) pattern where applicable, strictly separating the command model (for updates) and the query model (for reads). Ensure that not only the models but also the top-level modules containing them are separated or isolated. This separation helps in optimizing each side independently and clarifying responsibilities. Partial application of CQRS is permissible where full separation is not necessary, focusing on read-optimized models for specific needs.

**Instructions:**
- Define distinct models for command handling (write-side) and query handling (read-side).
- Structure code to physically or logically separate modules related to command processing from those related to query processing.

**Sample Code (Conceptual Module Structure):**

```
/src
  /Commands
    /Models
      UpdateUserModel.cs
    /Handlers
      UpdateUserCommandHandler.cs
    /Domain
      UserAggregate.cs // Command-side domain logic
  /Queries
    /Models
      UserDto.cs
    /Handlers
      GetUserQueryHandler.cs
    /ReadModel
      UserReadModel.cs // Read-side optimized structure
  /Common
    UserEvents.cs // Shared event definitions
```

## 2. Event Sourcing Fundamentals

When implementing systems that require a history of changes, robust state reconstruction, or complex stream processing, utilize Event Sourcing. The core principle is that the sequence of immutable domain events is the single source of truth for the application state.

**Instructions:**
- Represent all state changes as immutable domain events.
- Persist domain events in an append-only event store.
- Reconstruct the current state of an aggregate or system by replaying the sequence of events.
- Consider CDC (Change Data Capture) + Outbox pattern as an alternative approach where simply propagating events is insufficient, understanding the transaction log as a source of truth akin to Event Sourcing.

**Sample Code (Event Structure - C#):**

```csharp
public abstract record DomainEvent;

public record UserCreatedEvent(Guid UserId, string Username, string Email) : DomainEvent;
public record UsernameChangedEvent(Guid UserId, string NewUsername) : DomainEvent;

// Aggregate state reconstruction (Conceptual)
public class UserAggregate
{
    private Guid _userId;
    private string _username;
    private string _email;
    private int _version;
    private readonly List<DomainEvent> _changes = new();

    public static UserAggregate FromEvents(IEnumerable<DomainEvent> events)
    {
        var aggregate = new UserAggregate();
        foreach (var @event in events)
        {
            aggregate.ApplyEvent(@event);
            aggregate._version++;
        }
        return aggregate;
    }

    private void ApplyEvent(DomainEvent @event)
    {
        switch (@event)
        {
            case UserCreatedEvent created:
                _userId = created.UserId;
                _username = created.Username;
                _email = created.Email;
                break;
            case UsernameChangedEvent usernameChanged:
                _username = usernameChanged.NewUsername;
                break;
                // Handle other events
        }
    }

    public void ChangeUsername(string newUsername)
    {
        // Business logic/validation
        var @event = new UsernameChangedEvent(_userId, newUsername);
        ApplyEvent(@event);
        _changes.Add(@event);
    }

    public IEnumerable<DomainEvent> GetUncommittedChanges() => _changes;
}
```

## 3. Read Model Construction

Build read-optimized models (denormalized views, projections) by consuming domain events from the event store. These models are specifically designed for efficient querying and displaying data to the client, often deviating from the normalized structure of the write model.

**Instructions:**
- Create event consumers (projectors) that subscribe to the event stream.
- Update the read model database(s) based on incoming events.
- Design the read model schema to match client display requirements, potentially denormalizing data from multiple aggregates.
- Avoid using the command-side domain model directly for queries due to issues like N+1 problems, exposure of internal state via getters for DTO projection, and difficulty in query optimization.

**Sample Code (Event Projection - C#):**

```csharp
// Event Consumer/Projector
public class UserProjector
{
    private readonly ReadModelDatabaseContext _dbContext;

    public UserProjector(ReadModelDatabaseContext dbContext)
    {
        _dbContext = dbContext;
    }

    public void Handle(UserCreatedEvent @event)
    {
        var readModel = new UserReadModel
        {
            Id = @event.UserId,
            Username = @event.Username,
            Email = @event.Email,
            CreatedAt = DateTime.UtcNow // Add projection-specific data
        };
        _dbContext.Users.Add(readModel);
        _dbContext.SaveChanges();
    }

    public void Handle(UsernameChangedEvent @event)
    {
        var readModel = _dbContext.Users.Find(@event.UserId);
        if (readModel != null)
        {
            readModel.Username = @event.NewUsername;
            _dbContext.SaveChanges();
        }
    }

    // Handle other events...
}

// Read Model Entity
public class UserReadModel
{
    public Guid Id { get; set; }
    public string Username { get; set; }
    public string Email { get; set; }
    public DateTime CreatedAt { get; set; }
    // Additional denormalized fields for queries/display
}
```

## 4. DDD Value Object Design

Design objects that describe characteristics or attributes, and carry no concept of identity, as Value Objects. Implement them as immutable objects with value equality. Their behavior should be side-effect-free.

**Instructions:**
- Define Value Objects as classes or records distinguished by their attributes, not a unique identifier.
- Ensure Value Objects are immutable; their state cannot be changed after creation.
- Implement value equality (`Equals`, `GetHashCode`) based on all relevant attributes.
- Design methods on Value Objects to be side-effect-free, returning new instances if changes are needed.

**Sample Code (Value Object - C# Record):**

```csharp
public record Address(string Street, string City, string PostalCode, string Country);

// Usage demonstrating immutability and value equality
var address1 = new Address("123 Main St", "Anytown", "12345", "USA");
var address2 = new Address("123 Main St", "Anytown", "12345", "USA");
var address3 = address1 with { City = "Otherville" }; // Creates a new instance

Console.WriteLine(address1 == address2); // Output: True (Value equality)
Console.WriteLine(address1 == address3); // Output: False
```

## 5. DDD Entity Design

Model objects distinguished by their identity, rather than their attributes, as Entities. Keep their class definitions simple, focusing on their unique identity and lifecycle within the domain.

**Instructions:**
- Define Entities with a unique identifier that remains constant throughout their lifecycle.
- Focus the Entity's definition on its identity and the behaviors that change its state over time.
- Use Entities for complex concepts where identity is crucial for tracking and managing state changes.

**Sample Code (Entity - C#):**

```csharp
public class Order // Entity
{
    public Guid Id { get; } // Identity is primary

    private OrderStatus _status;
    private List<OrderItem> _items;

    // Constructor focuses on identity and initial state
    public Order(Guid id, OrderStatus status = OrderStatus.Created)
    {
        Id = id;
        _status = status;
        _items = new List<OrderItem>();
    }

    // Business logic (behavior) that changes state
    public void AddItem(OrderItem item)
    {
        // Validation and state change logic
        _items.Add(item);
    }

    public void Ship()
    {
        if (_status == OrderStatus.Created)
        {
            _status = OrderStatus.Shipped;
            // Record event (in Event Sourcing)
        }
    }

    // ... other behaviors ...
}

public enum OrderStatus { Created, Shipped, Delivered }
public record OrderItem(Guid ProductId, int Quantity); // Often a Value Object
```

## 6. Handling Optional/Failure with Types (Maybe/Either)

Represent the possibility of a value being absent or an operation failing using explicit types like `Maybe` (or `Option`) and `Either` (or `Result`) instead of null references or exceptions for expected outcomes. This makes the potential absence or failure explicit in the type signature.

**Instructions:**
- Use `Maybe`/`Option` types for values that might be null or absent.
- Use `Either`/`Result` types for operations that can either succeed with a result (typically on the "Right" side) or fail with an error (typically on the "Left" or "Error" side).
- Leverage functional patterns (`map`, `chain`/`bind`, pattern matching) to work with values inside these containers safely.

**Sample Code (Maybe/Result - F#):**

```fsharp
// Representing optional value
let findUserById userId =
    // Logic to find user
    if userId = 1 then Some { Id = 1; Name = "Alice" }
    else None

let userOption = findUserById 1
let userName = match userOption with
               | Some user -> user.Name
               | None -> "User not found" // Explicitly handle None

// Representing operation result (success or failure)
type PaymentError =
| CardTypeNotRecognized
| PaymentRejected

type PaymentResult = Result<PaidInvoice, PaymentError>

let processPayment invoice payment =
    // Logic to process payment
    if payment.Amount > invoice.Amount then Error PaymentRejected
    elif payment.CardType = "Unknown" then Error CardTypeNotRecognized
    else Ok { PaidInvoiceId = Guid.NewGuid(); AmountPaid = payment.Amount }
    |> printfn "Payment result: %A" // Explicitly handle Ok/Error
```

**Sample Code (Maybe/Either - JavaScript/Functional Style):**

```javascript
// Maybe implementation (simplified)
class Maybe {
    static of(x) { return x == null ? new Nothing() : new Just(x); }
    map(f) { throw new Error("Not implemented"); }
    chain(f) { throw new Error("Not implemented"); }
    // ... other methods (join, etc.)
}
class Just extends Maybe {
    constructor(x) { super(x); this.$value = x; }
    map(f) { return Maybe.of(f(this.$value)); }
    chain(f) { return f(this.$value); } // Flattens nested Maybe
}
class Nothing extends Maybe {
    map(f) { return this; }
    chain(f) { return this; }
}

// Usage
const safeHead = xs => Maybe.of(xs != null && xs.length > 0 ? xs : null);

safeHead()
    .map(x => x + 1) // Just(2)
    .chain(x => Maybe.of(x * 2)) // Just(4)
    .map(x => `Value is: ${x}`); // Just("Value is: 4")

safeHead([])
    .map(x => x + 1) // Nothing
    .chain(x => Maybe.of(x * 2)) // Nothing
    .map(x => `Value is: ${x}`); // Nothing

// Either implementation (simplified)
class Either {
    static of(x) { return new Right(x); }
    map(f) { throw new Error("Not implemented"); }
    chain(f) { throw new Error("Not implemented"); }
    // ...
}
class Left extends Either {
    constructor(x) { super(x); this.$value = x; }
    map(f) { return this; } // Ignores map
    chain(f) { return this; } // Ignores chain
}
class Right extends Either {
    constructor(x) { super(x); this.$value = x; }
    map(f) { return Either.of(f(this.$value)); }
    chain(f) { return f(this.$value); } // Flattens nested Either
}

// Usage
const validateName = name => name && name.length > 3
    ? Either.of(name)
    : new Left("Name must be longer than 3 characters");

validateName("Alice")
    .map(n => `Hello, ${n}`); // Right("Hello, Alice")

validateName("Bob")
    .map(n => `Hello, ${n}`); // Left("Name must be longer than 3 characters")
```

## 7. Functional Programming Concepts (Functors, Monads)

Apply functional programming concepts, particularly Functors and Monads, to structure computations involving containers, side effects, or potential failures. Understand `map` for applying a function inside a container, and `chain` (`>>=`) for sequencing computations that return nested containers, effectively flattening them. Use `of` (or `pure`) to place a value into a default minimal context (Pointed Functor).

**Instructions:**
- Use the `map` function (or method) to apply a function `a -> b` to a value inside a container `f a`, resulting in `f b`.
- Use the `chain` (or `flatMap`, `>>=`) function (or method) to sequence computations `a -> m b` when working with a monadic container `m a`, resulting in `m b`. This is used when the next step itself produces a container.
- Use the `of` (or `pure`) static method to lift a value into a container, `a -> m a`.
- Recognize common monadic types like `Maybe`/`Option`, `Either`/`Result`, `List`/`Array`, `IO`, `Task`, `State`.

**Sample Code (Monad `chain` - F#):**

```fsharp
// Example: Sequence Maybe computations using `>>=` (chain)
let maybeAdd (x : int option) (y : int option) : int option =
    x >>= (fun x_val ->
    y >>= (fun y_val ->
    Some (x_val + y_val))) // Lift final value back into Maybe

// Using do notation (syntactic sugar for `>>=`)
let maybeAddDo (x : int option) (y : int option) : int option =
    // The `do` keyword enables monadic sequencing
    // requires the Maybe type to have a Monad instance defined
    // (This is simplified syntax; actual Idris `do` or F# computation expressions are richer)
    // In F#, you'd use `option { ... }` computation expression
    // let! x_val = x
    // let! y_val = y
    // return x_val + y_val
    None // placeholder for F# option computation expression

// Example usage
maybeAdd (Some 5) (Some 3) |> printfn "%A" // Some 8
maybeAdd (Some 5) None |> printfn "%A" // None
```

**Sample Code (Functor `map`, Monad `chain` - JavaScript):**

```javascript
// map: Apply function inside the container
Maybe.of(5).map(x => x + 1); // Just(6)
Either.of(5).map(x => x + 1); // Right(6)
new Left("error").map(x => x + 1); // Left("error")
.map(x => x * 2); // (Array is a Functor)

// chain: Sequence computations returning containers
const safeDivideBy = divisor => num => divisor === 0 ? new Left("Divide by zero") : Either.of(num / divisor);

Either.of(10)
    .chain(safeDivideBy(2)) // Right(5)
    .chain(safeDivideBy(0)); // Left("Divide by zero")

Either.of(10)
    .chain(safeDivideBy(5)) // Right(2)
    .chain(safeDivideBy(2)); // Right(1)
```

## 8. Type-Driven Development (TFD) Workflow

Adopt a Type-Driven Development approach, starting with defining types to express abstract concepts and business logic. Use the type definitions and compiler feedback (intellisense, type checking errors) for rapid experimentation and refinement of the model before writing extensive implementation code. This enables quick iteration on design ideas and leverages the type system as a guide.

**Instructions:**
- Begin development by defining data types (Algebraic Data Types like records, discriminated unions) that model the problem domain and required behaviors.
- Use the compiler and tooling (REPL, IDE features like M-Enter in Idris/F#, waves under code in IDEs) to validate and iterate on type definitions quickly.
- Leverage compiler errors (red squiggly lines/waves) upon changing types as a form of fast feedback guiding necessary code adjustments.
- Postpone writing full implementation logic until the type structure is stable and expresses the domain effectively.

**Sample Workflow (Conceptual - F#/Idris style):**

1.  **Define Types:**
    ```fsharp
    type OrderId = OrderId of int
    type OrderLineId = OrderLineId of int
    type Quantity = Quantity of int // Ensure Quantity > 0 via smart constructor/validation
    type Price = Price of decimal // Ensure Price >= 0

    type OrderLine = { Id : OrderLineId; ProductId : Guid; Quantity : Quantity; Price : Price }

    type OrderStatus = Created | Placed | Shipped | Cancelled

    // Define potential errors for placing an order
    type PlaceOrderError =
    | OrderAlreadyPlaced
    | InsufficientStock of Guid * Quantity // ProductId and requested Quantity
    | InvalidOrderLines // e.g., empty list

    // Define the command/operation signature
    type PlaceOrderCommand = { OrderId : OrderId; CustomerId : Guid }
    type PlaceOrderResult = Result<OrderPlacedEvent, PlaceOrderError list>

    // Function signature expressing the intent
    val placeOrder : PlaceOrderCommand -> Order -> ProductStockService -> Result<OrderPlacedEvent, PlaceOrderError list>
    ```
2.  **Rapid Experimentation (REPL/M-Enter):** Use the REPL or similar interactive tools to test type construction and simple function signatures. Get immediate type feedback.
    ```fsharp
    // In F# Interactive or Idris REPL
    // Try creating instances, checking types
    let q = Quantity 10 // If smart constructor, test validation here
    let price = Price 5.99m
    let line = { Id = OrderLineId 1; ProductId = Guid.NewGuid(); Quantity = q; Price = price }

    // Check function type
    // > placeOrder;;
    // val placeOrder : PlaceOrderCommand -> Order -> ProductStockService -> Result<OrderPlacedEvent, PlaceOrderError list>
    ```
3.  **Refine based on feedback:** If types don't express constraints well (e.g., `Quantity` allows 0), refine the type definition (e.g., use a smart constructor or a refined type). Compiler errors guide where changes are needed.
4.  **Implement logic incrementally:** Once types are expressive, implement the function logic, leveraging pattern matching for ADTs. Compiler ensures all cases are handled.

## 9. Algebraic Data Types (ADTs)

Utilize Algebraic Data Types (ADTs), combining Product Types (records, structs) for representing compositions of data and Sum Types (discriminated unions, enums with associated data) for representing distinct possibilities or choices. ADTs are fundamental for modeling domain concepts precisely and leveraging pattern matching.

**Instructions:**
- Define Product Types for objects that have multiple attributes (e.g., a User has a Name and an Email).
- Define Sum Types for values that can be one of several distinct forms, potentially carrying different associated data (e.g., a Payment Method can be Credit Card, Debit Card, or PayPal, each with different details).
- Leverage the expressiveness of ADTs to encode business rules and constraints directly into the type system.

**Sample Code (ADTs - F#):**

```fsharp
// Product Type (Record)
type User = {
    Id : Guid
    Username : string
    Email : string option // Email is optional
}

// Sum Type (Discriminated Union)
type PaymentMethod =
| CreditCard of cardNumber:string * expiryDate:string * cvv:string
| DebitCard of cardNumber:string * expiryDate:string * cvv:string * issueNumber:string option
| PayPal of email:string
| BankTransfer of accountNumber:string * sortCode:string

// Using ADTs
let createUser id username email = { Id = id; Username = username; Email = email } // email can be None

let processPaymentUsing method =
    match method with
    | CreditCard(num, expiry, cvv) -> printfn "Processing Credit Card %s ending %s" num (expiry.[expiry.Length - 4..])
    | DebitCard(num, expiry, cvv, issue) -> printfn "Processing Debit Card..." // Handle issue option
    | PayPal(email) -> printfn "Processing PayPal for %s" email
    | BankTransfer(acc, sort) -> printfn "Processing Bank Transfer..."
```

## 10. Pattern Matching Implementation

Use pattern matching extensively, particularly with Sum Types (discriminated unions) and other algebraic structures (lists, options, results), to handle different cases or data shapes explicitly. This leads to more robust and readable code where the compiler can verify that all possibilities are covered.

**Instructions:**
- Use `match ... with` expressions (or similar constructs in other languages) to deconstruct ADTs and handle each possible case.
- Ensure all cases of a Sum Type are covered by pattern matching, allowing the compiler to check for completeness.
- Use pattern matching on data structures like lists (`[]`, `x :: xs`) or option types (`Some`, `None`) to handle their different forms.

**Sample Code (Pattern Matching - F#):**

```fsharp
type Shape =
| Circle of radius:double
| Rectangle of width:double * height:double
| Triangle of base:double * height:double

let calculateArea shape =
    match shape with
    | Circle r -> System.Math.PI * r * r
    | Rectangle(w, h) -> w * h
    | Triangle(b, h) -> 0.5 * b * h

let describeList (list: list<int>) =
    match list with
    | [] -> "Empty list"
    | [x] -> sprintf "List with one element: %d" x
    | x :: xs -> sprintf "List starts with %d, tail has %d elements" x (List.length xs)

calculateArea (Circle 5.0) |> printfn "%f"
describeList [1; 2; 3] |> printfn "%s"
```

**Sample Code (Pattern Matching with `case` - Idris):**

```idris
-- Data type for Natural numbers (defined in Prelude)
-- data Nat = Z | S Nat -- Z is zero, S is successor

isEven : Nat -> Bool
isEven Z = True
isEven (S Z) = False -- Equivalent to isEven 1 = False
isEven (S (S k)) = isEven k -- Equivalent to isEven (n+2) = isEven n

-- describeList example from sources
describeList : List Int -> String
describeList [] = "Empty"
describeList (x :: xs) = "Non-empty, tail = " ++ show xs

isEven (S (S (S (S Z)))) -- isEven 4 -> True
describeList -- "Non-empty, tail ="
```

## 11. Dependency Injection (DI) Principle

Implement dependency injection to manage dependencies between components. Prefer constructor injection. Avoid using the Ambient Context anti-pattern, which involves static properties or methods (`TimeProvider.Current`, `System.getSecurityManager()`) to access dependencies, as it hides dependencies, makes components harder to test and less flexible.

**Instructions:**
- Identify components' dependencies and provide them externally, typically through constructor parameters.
- Inject abstractions (interfaces) rather than concrete implementations where variation or testability is required.
- Strictly avoid static service locators or static properties to access dependencies (Ambient Context).

**Sample Code (Avoiding Ambient Context - C#):**

```csharp
// Ambient Context Anti-Pattern (AVOID THIS STYLE)
public interface ITimeProvider { DateTime Now { get; } }
public static class TimeProvider // Problematic static access
{
    public static ITimeProvider Current { get; set; } // Global mutable state/access point
}

public class Greeter // Hard to test/control time
{
    public string GetWelcomeMessage()
    {
        DateTime now = TimeProvider.Current.Now; // Accessing via static property
        string partOfDay = now.Hour < 6 ? "night" : "day";
        return string.Format("Good {0}.", partOfDay);
    }
}

// Preferred Approach (Dependency Injection)
public interface ITimeProvider { DateTime Now { get; } } // Same interface

public class Greeter // Dependencies injected
{
    private readonly ITimeProvider _timeProvider;

    // Constructor Injection
    public Greeter(ITimeProvider timeProvider)
    {
        _timeProvider = timeProvider;
    }

    public string GetWelcomeMessage()
    {
        DateTime now = _timeProvider.Now; // Accessing via injected dependency
        string partOfDay = now.Hour < 6 ? "night" : "day";
        return string.Format("Good {0}.", partOfDay);
    }
}

// Usage with DI Container (Conceptual)
// container.Register<ITimeProvider, SystemTimeProvider>(); // Real implementation
// container.Register<ITimeProvider, MockTimeProvider>(); // Test implementation
// var greeter = container.Resolve<Greeter>();
// var message = greeter.GetWelcomeMessage();
```

## 12. Decorator Pattern Implementation

Apply the Decorator pattern to dynamically add behavior to an object. A decorator wraps another component of the same abstraction (interface) and forwards calls, potentially adding logic before or after the forwarded call. Guard Clauses can be implemented as a form of Decorator for null input.

**Instructions:**
- Define an interface for the core component and decorators to implement.
- Create a concrete component implementing the interface.
- Create a Decorator class that takes an instance of the interface in its constructor (the "decoratee").
- Implement the interface methods in the Decorator by calling the corresponding methods on the decoratee and adding pre/post behavior.
- Use Decorators to add cross-cutting concerns like logging, validation (e.g., null checks), or security checks without modifying the original component.

**Sample Code (Decorator - C#):**

```csharp
public interface IGreeter
{
    string Greet(string name);
}

public class FormalGreeter : IGreeter // Concrete Component
{
    public string Greet(string name)
    {
        return $"Hello, Mr. {name}.";
    }
}

public class NiceToMeetYouGreeterDecorator : IGreeter // Decorator
{
    private readonly IGreeter _decoratee;

    public NiceToMeetYouGreeterDecorator(IGreeter decoratee)
    {
        _decoratee = decoratee;
    }

    public string Greet(string name)
    {
        string greeting = _decoratee.Greet(name);
        return $"{greeting} Nice to meet you."; // Adding behavior
    }
}

public class NullInputGreeterDecorator : IGreeter // Decorator as a Guard Clause
{
    private readonly IGreeter _decoratee;

    public NullInputGreeterDecorator(IGreeter decoratee)
    {
        _decoratee = decoratee;
    }

    public string Greet(string name)
    {
        if (string.IsNullOrEmpty(name))
        {
            return "Hello."; // Default behavior for null/empty input
        }
        return _decoratee.Greet(name); // Forward call if input is valid
    }
}

// Usage
// IGreeter greeter = new FormalGreeter();
// IGreeter decoratedGreeter = new NiceToMeetYouGreeterDecorator(greeter);
// IGreeter nullSafeDecoratedGreeter = new NullInputGreeterDecorator(decoratedGreeter);

// Console.WriteLine(nullSafeDecoratedGreeter.Greet("Samuel L. Jackson"));
// Output: Hello, Mr. Samuel L. Jackson. Nice to meet you.
// Console.WriteLine(nullSafeDecoratedGreeter.Greet(null));
// Output: Hello.
```

## 13. Test-Driven Development (TDD) Cycle

Practice Test-Driven Development (TDD) following the Red/Green/Refactor cycle. Write a failing test (Red), write just enough code to make it pass (Green), and then improve the code while keeping the test passing (Refactor). Use implementation strategies like Fake It, Triangulation, or Obvious Implementation as appropriate to drive development with small, confident steps.

**Instructions:**
- Write an automated test case for a small piece of desired functionality.
- Run the test and ensure it fails as expected (Red).
- Write the simplest possible code in the application logic to make the failing test pass (Green). Strategies:
    - **Fake It:** Return a constant value or hardcoded result that satisfies the current test.
    - **Triangulation:** Once two or more tests targeting slightly different inputs for the same logic exist, generalize the implementation to pass all tests.
    - **Obvious Implementation:** If the implementation is trivial and immediately clear, write the correct code directly. Use this cautiously if surprised by failures.
- Run the tests again to confirm they all pass (Green).
- Refactor the code (application and test code) to improve its design, readability, and structure without changing its behavior. Ensure tests remain green after refactoring.
- Repeat the cycle for the next small piece of functionality.
- Use assertions (`assertEquals`, `Assert.Equal`, etc.) to verify expected outcomes in tests.

**Sample Workflow (TDD - Conceptual, inspired by sources):**

1.  **Red:** Write a test for a `plus` function in a Money object that adds two amounts. Initially, the `plus` function might not exist or returns an incorrect value.
    ```csharp
    [Test]
    public void testSum()
    {
        // Assuming Money class with a Plus method
        Money five = Money.dollar(5);
        Money result = five.plus(five);
        Assert.AreEqual(Money.dollar(10), result); // This test will fail (Red)
    }
    ```
2.  **Green (Fake It):** Implement just enough code to make the test pass. A simple `plus` method might return a hardcoded `Money.dollar(10)`.
    ```csharp
    public class Money
    {
        protected int amount;
        protected string currency; // Add currency later?

        public Money(int amount, string currency) { this.amount = amount; this.currency = currency; }
        public static Money dollar(int amount) { return new Money(amount, "USD"); }

        // Simple implementation to pass the first test (Fake It)
        public Money plus(Money addend)
        {
            return Money.dollar(10); // Faked result
        }

        // Need equals for Assert.AreEqual
        public override bool Equals(object obj)
        {
            // Fake it for now based on the test inputs (5 + 5 = 10)
            return true; // This is a Fake It step based on expected 5+5==10
        }
        public override int GetHashCode() { return 0; } // Dummy
    }
    ```
3.  **Red (New Test/Triangulation):** Add a new test case with different inputs. The current faked implementation will likely fail.
    ```csharp
    [Test]
    public void testSum()
    {
        Money five = Money.dollar(5);
        Money result = five.plus(five);
        Assert.AreEqual(Money.dollar(10), result); // Passes

        // New test case (Red)
        Money three = Money.dollar(3);
        Money seven = Money.dollar(7);
        Money sum = three.plus(seven); // Will likely fail with the Faked plus
        Assert.AreEqual(Money.dollar(10), sum);
    }
    ```
4.  **Green (Generalize/Triangulation):** Modify the `plus` implementation to pass both tests. This often involves the actual logic.
    ```csharp
    public Money plus(Money addend)
    {
        // Generalize the implementation
        int sumAmount = this.amount + addend.amount;
        return new Money(sumAmount, this.currency); // Assuming same currency for now
    }

     public override bool Equals(object obj)
    {
        if (obj == null || GetType() != obj.GetType())
        {
            return false;
        }
        Money other = (Money)obj;
        // Now implement actual value equality
        return amount == other.amount && currency == other.currency;
    }
    public override int GetHashCode() => HashCode.Combine(amount, currency);
    ```
5.  **Refactor:** Improve the code's structure, naming, etc., while keeping all tests passing.

## 14. Immutability

Ensure objects, especially Value Objects, are immutable. Once created, their state cannot be changed. This contributes to thread safety, easier reasoning about data flow, and aligns well with functional programming principles and concepts like event sourcing.

**Instructions:**
- Make all fields in immutable objects `readonly` (or equivalent in the language).
- Ensure no methods modify the object's internal state; methods that conceptually "change" the object should return a new instance with the modified state.
- Avoid exposing mutable collections directly; return copies or immutable wrappers.

**Sample Code (Immutability - C# Record):**

```csharp
// Immutable object (Record)
public record ProductId(Guid Value); // Single-property Value Object

public record OrderLineItem // Immutable Value Object
{
    public ProductId ProductId { get; init; } // init-only setter for object initializer
    public int Quantity { get; init; }
    public decimal Price { get; init; }

    // Constructor ensures initial state is set
    public OrderLineItem(ProductId productId, int quantity, decimal price)
    {
        if (quantity <= 0) throw new ArgumentOutOfRangeException(nameof(quantity));
        if (price < 0) throw new ArgumentOutOfRangeException(nameof(price));

        ProductId = productId;
        Quantity = quantity;
        Price = price;
    }

    // Example of "modifying" - returns a new instance
    public OrderLineItem WithQuantity(int newQuantity)
    {
        if (newQuantity <= 0) throw new ArgumentOutOfRangeException(nameof(newQuantity));
        return this with { Quantity = newQuantity }; // C# `with` expression creates a new record
    }
}

// Usage
var productId = new ProductId(Guid.NewGuid());
var item1 = new OrderLineItem(productId, 5, 10.5m);

// item1.Quantity = 6; // Compile-time error - immutable

var item2 = item1.WithQuantity(7); // item1 remains unchanged, item2 is a new instance
Console.WriteLine($"Item1 Quantity: {item1.Quantity}"); // Output: 5
Console.WriteLine($"Item2 Quantity: {item2.Quantity}"); // Output: 7
```

## 15. Error Handling (Type-Based)

Handle errors explicitly using types (`Result`, `Either`) and pattern matching, rather than relying solely on exceptions for expected failure conditions. This makes failure a part of the function's return type and forces callers to handle potential errors, improving robustness and clarity.

**Instructions:**
- Define specific error types (often Sum Types/Discriminated Unions) for different failure reasons.
- Return `Result<SuccessType, ErrorType>` or `Either<ErrorType, SuccessType>` from functions that might fail.
- Use pattern matching (`match`, `case`) to handle both success and failure cases when calling functions that return `Result`/`Either`.
- Avoid using exceptions for control flow or anticipated errors.

**Sample Code (Result Pattern - C#):**

```csharp
// Using a Result type (common pattern in C# functional libraries)
public readonly struct Result<TValue, TError>
{
    private readonly TValue _value;
    private readonly TError _error;
    public bool IsSuccess { get; }
    public bool IsFailure => !IsSuccess;

    private Result(TValue value) { _value = value; IsSuccess = true; _error = default; }
    private Result(TError error) { _error = error; IsSuccess = false; _value = default; }

    public static Result<TValue, TError> Success(TValue value) => new Result<TValue, TError>(value);
    public static Result<TValue, TError> Failure(TError error) => new Result<TValue, TError>(error);

    public T Match<T>(Func<TValue, T> onSuccess, Func<TError, T> onFailure)
        => IsSuccess ? onSuccess(_value) : onFailure(_error);

    // Add Map, Bind/Chain, etc. methods for fluent/functional usage
}

public enum FileOperationError { FileNotFound, PermissionDenied, DiskFull }

public static class FileService
{
    public static Result<string, FileOperationError> ReadFile(string path)
    {
        if (!File.Exists(path)) return Result<string, FileOperationError>.Failure(FileOperationError.FileNotFound);
        try
        {
            string content = File.ReadAllText(path);
            return Result<string, FileOperationError>.Success(content);
        }
        catch (UnauthorizedAccessException)
        {
            return Result<string, FileOperationError>.Failure(FileOperationError.PermissionDenied);
        }
        catch (IOException ex) when (ex.Message.Contains("disk full"))
        {
            return Result<string, FileOperationError>.Failure(FileOperationError.DiskFull);
        }
        catch (Exception)
        {
            // Handle unexpected errors as exceptions or a generic error case
            throw;
        }
    }
}

// Usage
Result<string, FileOperationError> result = FileService.ReadFile("mydata.txt");

result.Match(
    onSuccess: content => { Console.WriteLine($"File content: {content}"); return 0; },
    onFailure: error => { Console.Error.WriteLine($"Failed to read file: {error}"); return -1; }
);
```

## 16. Concurrency & State Management (Functional Style)

Model side effects, state, and concurrency using explicit types and patterns (`IO`, `Task`, `State`, `RunIO`, `Process`) rather than relying on implicit side effects or shared mutable state. Use `do` notation (or computation expressions) as syntactic sugar for sequencing these operations (`>>=`).

**Instructions:**
- Represent actions with potential side effects (I/O, state changes) using types like `IO`, `Task`, `State`.
- Structure sequences of these actions using monadic functions (`chain`/`>>=`) or the `do` notation which translates to these functions.
- Keep functions dealing with core logic pure (no side effects), separating the effectful parts to the boundaries of the application or module.

**Sample Code (IO and `do` notation - Idris):**

```idris
-- Example: Reading a string from console and printing
greet : IO ()
greet = do
    putStr "Enter your name: " -- IO action
    name <- getLine         -- IO action, bind result to 'name'
    if name == ""
        then putStrLn "Bye bye!" -- IO action
        else do                  -- Sequence more IO actions
            putStrLn ("Hello " ++ name)
            greet                -- Recursive call (productive/potentially infinite process)

-- The `do` block sequences IO actions using the (>>=) operator for IO
-- putStr "Enter your name: " >>= \_ => -- underscore ignores result of putStr
-- getLine >>= \name =>
-- case name == "" of
--     True => putStrLn "Bye bye!" >>= \_ => pure () -- pure () is an IO action that does nothing but returns ()
--     False => putStrLn ("Hello " ++ name) >>= \_ => greet

-- To run an IO action (unsafe in pure languages, but necessary at entry point)
-- partial main : IO ()
-- main = greet -- In Idris, typically run via `:exec main`

-- Example: Sequencing Maybe computations with do notation
maybeAdd : Maybe Int -> Maybe Int -> Maybe Int
maybeAdd x y = do
    x_val <- x -- If x is Nothing, the entire computation is Nothing
    y_val <- y -- If y is Nothing, the entire computation is Nothing
    Just (x_val + y_val) -- Wrap the result back in Just
    -- This is syntactic sugar for x >>= \x_val => y >>= \y_val => Just (x_val + y_val)
```

## 17. Equality Checking with Types (Dependent Types)

Use type-level mechanisms where available (Dependent Types in Idris) to express and prove properties about data, such as equality between values being guaranteed by the type system itself. This provides strong compile-time guarantees about relationships between data, beyond simple value equality checks at runtime.

**Instructions:**
- Where supported by the language, use built-in equality types (`x = y`) to represent the proposition that two values are equal.
- Use proof terms (like `Refl` in Idris) to provide evidence that equality holds.
- Leverage functions (`cong`, `sym`, `plusZeroRightNeutral`, etc. in Idris's Prelude) to construct proofs of equality or rewrite terms based on equality.
- Use `Dec` type to represent decidable properties (where you can prove either A or not A), which is useful for checks that might succeed or fail (like checking if an element is in a list, or if two numbers are equal).

**Sample Code (Equality Types - Idris):**

```idris
-- Built-in equality type: (=) : (a : Type) -> a -> a -> Type
-- Refl : (x : a) -> x = x -- Proof term for identity equality

-- Function to check if two Nat numbers are equal, returning a proof if they are
checkEqNat : (num1 : Nat) -> (num2 : Nat) -> Maybe (num1 = num2)
checkEqNat Z Z = Just Refl -- Proof that 0 = 0
checkEqNat Z (S k) = Nothing -- 0 cannot equal a successor
checkEqNat (S k) Z = Nothing -- A successor cannot equal 0
checkEqNat (S k) (S j) = case checkEqNat k j of -- Recursively check the predecessors
    Nothing => Nothing
    Just prf => Just (cong prf) -- If k=j is proven, then S k = S j is also proven via cong

-- Using the Dec type for decidable equality on Nat
-- Dec : Type -> Type -> Type (simplified) -> Either (not A) A -> Dec A notA
-- Dec a b = Yes a | No b -- where b is a proof that a cannot exist (Void)
-- DecEq interface provides decEq : (x : a) -> (y : a) -> Dec (x = y)

-- checkEqNat using Dec
-- checkEqNat' : (num1 : Nat) -> (num2 : Nat) -> Dec (num1 = num2)
-- checkEqNat' num1 num2 = decEq num1 num2 -- Leverages the DecEq instance for Nat

-- Example usage
-- checkEqNat 3 3   -- Just (Refl {x = 3})
-- checkEqNat 3 4   -- Nothing
-- checkEqNat' 3 3  -- Yes (Refl {x = 3})
-- checkEqNat' 3 4  -- No proof_that_3_ne_4 (internally uses Void)
```

## 18. Pattern Matching with Views

Use Views as a mechanism to define custom ways to pattern match on data structures, allowing destructuring based on logical properties (e.g., the last element of a list, or splitting a list in half) rather than just the physical constructors. This can simplify code that processes data in non-standard ways. Use `with` blocks where available to apply views concisely.

**Instructions:**
- Define a View type that represents the desired structure or property for pattern matching.
- Define a "covering function" that maps a value of the original type to an instance of the View type. This function must be total.
- Use `with` blocks (in Idris) or similar constructs to apply the covering function and then pattern match on the resulting view.
- Use this technique for traversals or pattern matching that are difficult with standard constructors (e.g., accessing the last element of a list).

**Sample Code (Views - Idris):**

```idris
-- Define a View for a list that exposes its last element
data ListLast : List a -> Type where
    Empty    : ListLast [] -- View for an empty list
    NonEmpty : (xs : List a) -> (x : a) -> ListLast (xs ++ [x]) -- View for a non-empty list

-- Define the covering function
total listLast : (xs : List a) -> ListLast xs
listLast [] = Empty
listLast (x :: xs) = case listLast xs of -- Recursively find the last element's view
    Empty => NonEmpty [] x -- x is the last element of [x]
    NonEmpty ys y => NonEmpty (x :: ys) y -- y is the last element of xs, now it's y for (x::ys)

-- Define a function using the view with a 'with' block
describeListEnd : List Int -> String
describeListEnd input with (listLast input) -- Apply the listLast view to input
    describeListEnd [] | Empty = "Empty" -- Pattern match on the input AND the view
    describeListEnd (xs ++ [x]) | (NonEmpty xs x) = -- Match on the 'physical' list structure AND the view structure
        "Non-empty, initial portion = " ++ show xs

-- Example usage:
-- describeListEnd -- "Non-empty, initial portion ="
-- describeListEnd [] -- "Empty"
```

## 19. Defensive Programming

Implement defensive programming techniques such as Guard Clauses and preconditions to handle invalid inputs or states early. Use techniques like the If-Then-Throw pattern for preconditions or Guard Clauses at the beginning of functions to validate arguments and state, failing fast if conditions are not met.

**Instructions:**
- At the start of a method or function, check for invalid arguments or states using conditional statements.
- If a condition indicates an invalid state or input that the method cannot handle, throw an appropriate exception immediately (If-Then-Throw pattern) or return an explicit error type (`Result`, `Either`).
- Use Guard Clauses (conditions followed by early exit like return or throw) to make the main logic path clearer.

**Sample Code (Guard Clauses - C#):**

```csharp
public class OrderService
{
    public void PlaceOrder(Order order)
    {
        // Guard Clauses for preconditions
        if (order == null)
        {
            throw new ArgumentNullException(nameof(order)); // If-Then-Throw Pattern
        }
        if (order.Items == null || !order.Items.Any())
        {
            throw new InvalidOperationException("Order must contain items.");
        }
         if (order.Status != OrderStatus.Created)
        {
            throw new InvalidOperationException($"Cannot place order with status: {order.Status}");
        }

        // Main logic proceeds only if guards pass
        // ... place order logic ...
        order.Status = OrderStatus.Placed; // Assuming Order is mutable for this example
    }

    // Alternative using type-based error handling
     public Result<Success, Error> PlaceOrderFunctional(Order order)
    {
        // Guard Clauses returning Error types
        if (order == null)
        {
            return Result<Success, Error>.Failure(Errors.ArgumentNull(nameof(order)));
        }
        if (order.Items == null || !order.Items.Any())
        {
            return Result<Success, Error>.Failure(Errors.Validation("Order must contain items."));
        }
        // ... etc.

        // ... success logic ...
        return Result<Success, Error>.Success(...);
    }
}
```

## 20. Polyglot Persistence Consideration

When designing microservices or systems with diverse data needs, recognize that a single storage technology may not be optimal for all services. Consider using Polyglot Persistence, where different services or bounded contexts use different database technologies (e.g., relational database, document database, event store, graph database) best suited for their specific data access patterns and requirements. This requires learning and managing multiple technologies but can be a reasonable investment for optimized storage solutions per domain/service.

**Instructions:**
- Evaluate the data storage requirements for each distinct service, bounded context, or read model.
- Select the database technology (RDBMS, NoSQL variants, Event Store) that best fits the specific workload (transactional, analytical, key-value, graph, event stream) and data structure of that component.
- Accept the complexity of managing multiple database technologies as a potential trade-off for performance and suitability.

**Sample Scenario (Conceptual):**

- **Order Processing (Command Side):** Requires strong transactional consistency, might use a relational database or an event store.
- **Product Catalog (Read Side):** Requires fast, flexible querying and potentially denormalized data; could use a document database optimized for reads.
- **User Activity Feed (Read Side):** A stream of events optimized for fast append and retrieval in time order; could use a dedicated event store or a specialized time-series database.
- **Recommendation Engine:** Might require modeling relationships between users and products; could use a graph database.

**Configuration Example (Conceptual - assuming configuration file):**

```yaml
# application.yaml or similar configuration
services:
  order-service:
    database:
      type: postgresql
      connectionString: ${POSTGRES_CONNECTION}
  product-catalog-service:
    database:
      type: mongodb
      connectionString: ${MONGODB_CONNECTION}
      readModelCollection: products
  activity-feed-service:
    database:
      type: eventstore
      connectionString: ${EVENTSTORE_CONNECTION}
  recommendation-service:
    database:
      type: neo4j
      uri: ${NEO4J_URI}
```
