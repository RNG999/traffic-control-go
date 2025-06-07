# Coding Agent Custom Instructions

As a coding agent, your primary directive is to produce high-quality, maintainable, performant, and secure code artifacts. Adhere strictly to the following principles and guidelines, which are derived from best practices and patterns identified in the analyzed documentation.

## 1. Naming: Choose Descriptive Names

Names must clearly indicate the purpose, function, or state of the entity they represent. Avoid cryptic abbreviations or single-letter names unless the scope is extremely limited.

**Instruction:** Use full, descriptive words that convey meaning.
**Sample Code:**

```go
// Bad:
var i int // What does 'i' represent?

// Good:
var fileLineCount int // Clearly indicates the count of lines in a file

// Bad:
func proc(d []byte) error { ... }

// Good:
func processInputData(data []byte) error { ... }
```

## 2. Naming: Match Abstraction Level

Names should reflect the level of abstraction of the code they reside in. Higher-level modules or functions should use higher-level names.

**Instruction:** Ensure names correspond to the conceptual level of the surrounding code.
**Sample Code:**

```java
// Within a high-level service class:
public void processOrder(Order order) {
    // Calls lower-level repository methods
    orderRepository.save(order);
    emailService.sendOrderConfirmation(order);
}

// Within the OrderRepository implementation:
private DatabaseConnection getConnection() { ... } // Lower-level detail
private PreparedStatement buildInsertStatement(Order order) { ... } // Lower-level detail
```

## 3. Naming: Use Standard Nomenclature

When standard terms exist within the problem domain or established programming patterns, use them consistently.

**Instruction:** Adopt well-known names where applicable (e.g., standard patterns, domain terms).
**Sample Code:**

```java
// Using standard DDD terms for classes
public class OrderService { ... }
public class OrderRepository { ... }
public class Order { ... } // Entity
public class Address { ... } // Value Object

// Using standard pattern names for methods
public Employee makeEmployee(EmployeeRecord r) { ... } // Factory Method
public void processRequestsForever() { ... } // Process Loop
```

## 4. Naming: Ensure Unambiguous Names

Names should have a single, clear meaning. Avoid names that could be interpreted in multiple ways.

**Instruction:** Choose names that prevent confusion about the entity's role or behavior.
**Sample Code:**

```java
// Bad:
public class Customer {
    private Address address; // Which address? Shipping, billing, residential?
}

// Good:
public class Customer {
    private ShippingAddress shippingAddress;
    private BillingAddress billingAddress;
}
```

## 5. Naming: Relate Name Length to Scope

The length of a name should be proportional to the size of its scope. Shorter names are acceptable in short methods with local variables, but longer names are required for variables or functions with larger scopes.

**Instruction:** Use longer, more precise names for wider scopes (fields, public methods) and shorter names for narrow scopes (local variables in short methods).
**Sample Code:**

```java
// Short scope, short name OK:
public void processList(List<String> items) {
    for (int i = 0; i < items.size(); i++) { // 'i' is acceptable here
        // ... use i ...
    }
}

// Wider scope, longer name needed:
private List<CustomerOrder> pendingOrders; // Field name should be descriptive
public void addOrderToPendingList(CustomerOrder order) { ... } // Method name should be descriptive
```

## 6. Naming: Avoid Encodings

Do not include type or scope information (like `m_`, `f_`, Hungarian notation, or prefixes indicating interfaces) in variable or function names, as modern IDEs provide this information.

**Instruction:** Refrain from encoding type or scope prefixes in names.
**Sample Code:**

```java
// Bad:
private int m_count; // Field prefix 'm_'
private String fName; // Field prefix 'f'
int iSize; // Type encoding 'i' for int

// Good:
private int count;
private String name;
int size;
```

## 7. Naming: Reflect Side-Effects in Names

If a function or variable has a side-effect (e.g., modifying state), this should be clear from its name.

**Instruction:** Use verbs that imply action or modification for methods with side-effects; avoid property-style names.
**Sample Code:**

```java
// Bad:
public boolean isPayday() {
    calculatePay(); // Side-effect not indicated by name
    return someCondition;
}

// Good:
public boolean isPayday() {
    // No side-effects
    return calculatePaydayStatus();
}

public void calculateAndDeliverPay() { // Name indicates action and side-effects
    Money pay = calculatePay();
    deliverPay(pay);
}
```

## 8. Functions: Keep Them Short

Functions should be very small, ideally just a few lines long. This enhances readability and makes functions easier to understand and test.

**Instruction:** Aim for functions that are typically under 20 lines, preferably much shorter (2-4 lines).
**Sample Code:**

```javascript
// Bad:
function renderHtml(data) {
    let html = '';
    // Lots of complex logic, mixed levels of abstraction
    if (data.type === 'header') {
        html += '<h1>' + data.title + '</h1>';
        // More header specific logic
    } else if (data.type === 'paragraph') {
        html += '<p>' + data.content + '</p>';
        // More paragraph specific logic
    }
    // ... much more logic ...
    return html;
}

// Good (using extracted methods):
function renderHtml(data) {
    let html = '';
    html += renderHeader(data.headerData);
    html += renderBody(data.bodyData);
    html += renderFooter(data.footerData);
    return html;
}

function renderHeader(headerData) { // Short, focused function
    // Specific header rendering logic
    return '<h1>' + headerData.title + '</h1>';
}
// ... other short functions ...
```

## 9. Functions: Ensure They Do One Thing

Functions should perform a single, well-defined task, and do it well. If a function has multiple sections performing different operations, extract those sections into separate functions.

**Instruction:** Refactor functions with multiple responsibilities into smaller, single-responsibility functions.
**Sample Code:**

```java
// Bad:
public void processOrder(Order order) {
    // Validate order
    validateOrder(order);
    // Save order to database
    orderRepository.save(order);
    // Send confirmation email
    emailService.sendOrderConfirmation(order);
}

// Good (if validation, saving, and sending are distinct conceptual steps):
public void processOrder(Order order) {
    validateOrder(order);
    saveOrder(order);
    sendConfirmationEmail(order);
}

private void validateOrder(Order order) { /* ... validation logic ... */ }
private void saveOrder(Order order) { /* ... saving logic ... */ }
private void sendConfirmationEmail(Order order) { /* ... email logic ... */ }
```

## 10. Functions: Avoid Flag Arguments

Flag arguments make a function do more than one thing. Instead of using a boolean or enum parameter to select behavior, create separate functions for each behavior.

**Instruction:** Replace functions with flag arguments by creating distinct functions for each path.
**Sample Code:**

```java
// Bad:
public void calculatePay(Employee e, boolean includeOvertime) {
    // ... calculate base pay ...
    if (includeOvertime) {
        // ... calculate and add overtime pay ...
    }
    // ...
}

// Good:
public void calculateBasePay(Employee e) { /* ... */ }
public void calculatePayWithOvertime(Employee e) { /* ... */ }
```

## 11. Functions: Limit Arguments

The ideal number of arguments for a function is zero (niladic). One (monadic) or two (dyadic) arguments are acceptable. Three (triadic) should be avoided where possible. More than three arguments is strongly discouraged and suggests that some arguments could be grouped into an object.

**Instruction:** Minimize the number of function arguments. Consider introducing parameter objects for multiple related arguments.
**Sample Code:**

```java
// Bad:
public void createBooking(Customer customer, Room room, Date startDate, Date endDate, int numberOfGuests, String specialRequests, double discountRate) { ... } // Too many arguments

// Good:
public void createBooking(BookingDetails details) { ... }

public class BookingDetails { // Parameter Object
    private Customer customer;
    private Room room;
    private Date startDate;
    private Date endDate;
    private int numberOfGuests;
    private String specialRequests;
    private double discountRate;
    // Constructor, getters...
}
```

## 12. Code Readability: Express Intent Clearly

Code should be as expressive as possible. Avoid constructs that obscure the author's intent, such as run-on expressions or cryptic calculations.

**Instruction:** Use intermediate variables with meaningful names to break down complex expressions and clarify intent.
**Sample Code:**

```java
// Bad:
public int m_otCalc() { // Obscured intent, bad naming, run-on expression
    return iThsWkd * iThsRte + (int) Math.round(0.5 * iThsRte * Math.max(0, iThsWkd - 400));
}

// Good:
public int calculateOvertimePay(int tenthsWorked, int tenthRate) {
    int straightTimeTenths = Math.min(tenthsWorked, 400); // Magic number replaced later
    int overTimeTenths = Math.max(0, tenthsWorked - straightTimeTenths);
    double overTimeRate = 1.5; // Magic number replaced later
    double bonusPay = overTimeTenths * tenthRate * (overTimeRate - 1.0); // Calculation explained
    return (int) Math.round(bonusPay);
}
```

## 13. Code Readability: Replace Magic Numbers

Raw numeric literals (magic numbers) should be replaced with well-named constants that explain their meaning, except for a few very obvious exceptions like 0 or 1 in simple formulas.

**Instruction:** Define constants for numbers that have a specific meaning beyond their literal value.
**Sample Code:**

```java
// Bad:
if (employee.getTenthsWorked() > 400) { // What does 400 mean? Overtime threshold?
    // ... calculate overtime ...
}
double overtimeRate = 1.5; // What does 1.5 represent? 1.5x pay?

// Good:
private static final int STANDARD_WORK_TENHTS_PER_WEEK = 400;
private static final double OVERTIME_PAY_MULTIPLIER = 1.5;

if (employee.getTenthsWorked() > STANDARD_WORK_TENHTS_PER_WEEK) {
    // ... calculate overtime using OVERTIME_PAY_MULTIPLIER ...
}
```

## 14. Code Readability: Avoid Mental Mapping

Do not require the reader to mentally translate variables into other concepts or rely on implicit knowledge to understand the code.

**Instruction:** Ensure variable names and code structures are self-explanatory, reducing the need for mental translation.
**Sample Code:**

```java
// Bad:
assertEquals("hBChl", hw.getState()); // What does "hBChl" mean?

// Good (with explanatory constant or breakdown):
private static final String HEATER_ON_BLOWER_ON_COOLER_ON_HITEMP_ALARM_ON_LOTEMP_ALARM_OFF = "hBChl";
assertEquals(HEATER_ON_BLOWER_ON_COOLER_ON_HITEMP_ALARM_ON_LOTEMP_ALARM_OFF, hw.getState());

// Or, ideally, the state is represented by an object or clear structure.
```

## 15. Comments: Explain Intent

Use comments to explain the intent behind a piece of code, especially if the code's purpose is not immediately obvious from its structure and naming.

**Instruction:** Add comments that clarify the reasoning or purpose of non-obvious code segments.
**Sample Code:**

```java
// Bad (no explanation):
// This calculation uses the square root of the maximum value.
// return (int) Math.sqrt(maxValue) * 2;

// Good (explaining the 'why'):
// We use the square root of the maximum value as the iteration limit
// for prime generation, as any composite number n must have a prime
// factor less than or equal to the square root of n.
// return (int) Math.sqrt(maxValue) * 2;
```

## 16. Comments: Avoid Redundant Comments

Comments should provide information not already present in the code. Avoid comments that merely restate what the code is doing or provide irrelevant historical details.

**Instruction:** Remove comments that are duplicative, outdated, or contain unnecessary information.
**Sample Code:**

```java
// Bad (redundant):
// increment the counter
counter++;

// Bad (irrelevant historical detail):
/*
This code implements the Base64 encoding described in RFC 4648,
which superseded RFCs 3548 and 2045. The prior RFCs had slightly
different padding rules in some corner cases...
*/
String encoded = base64Encode(data); // Keep the relevant code
```

## 17. Comments: Keep Comments Updated

Outdated comments are worse than no comments, as they can mislead the reader. Ensure comments accurately reflect the current state of the code.

**Instruction:** When changing code, review and update any associated comments to maintain accuracy.
**Sample Code:**

```java
// Code changes from returning -1 on error to throwing an exception
// Old comment (now wrong):
/**
 * Converts string to weekday code.
 * Returns -1 if the string is not convertable, the day of the week otherwise.
 */
// public static int stringToWeekdayCode(String s) { ... return -1; }

// New comment (reflecting exception):
/**
 * Converts string to weekday code.
 * @throws IllegalArgumentException if the string is not a valid weekday name.
 */
// public static Day stringToWeekdayCode(String s) { ... throw new IllegalArgumentException(...); }
```

## 18. Formatting: Maintain Vertical Density

Related code should be kept vertically close. Keep instance variables at the top of the class, followed by public functions, and then private utilities, ideally topologically sorted by dependency.

**Instruction:** Organize code vertically to group related elements and improve flow.
**Sample Code:**

```java
// Class structure example:
public class ReportGenerator {
    // Instance variables
    private ReportData data;
    private ReportFormatter formatter;

    // Constructor
    public ReportGenerator(ReportData data, ReportFormatter formatter) {
        this.data = data;
        this.formatter = formatter;
    }

    // Public methods (API)
    public String generateReport() {
        // ... logic ...
        return formatter.format(data);
    }

    // Private utility methods (called by public methods)
    private void processData() { /* ... */ }
    private void validateData() { /* ... */ }
}
```

## 19. Formatting: Limit Line Length

Keep lines of code short to enhance readability and avoid horizontal scrolling. While exact limits can vary (e.g., 80, 100, or 120 characters), consistency is key.

**Instruction:** Adhere to a defined maximum line length (e.g., 120 characters). Break long lines logically.
**Sample Code:**

```java
// Bad:
SomeVeryLongClassName descriptiveVariableName = new SomeVeryLongClassName(argument1, argument2, argument3, argument4, argument5); // Exceeds typical line length

// Good (wrapped):
SomeVeryLongClassName descriptiveVariableName =
    new SomeVeryLongClassName(
        argument1,
        argument2,
        argument3,
        argument4,
        argument5);
```

## 20. Formatting: Use Consistent Indentation

Use consistent indentation to clearly delineate code blocks and scope. Avoid collapsing scopes onto a single line.

**Instruction:** Always use proper indentation for code blocks (if, for, while, method bodies, etc.).
**Sample Code:**

```java
// Bad (collapsed scope):
if (condition) return result;

// Bad (incorrect indentation):
if (condition) {
result = 0;
}

// Good:
if (condition) {
    result = 0;
}
```

## 21. Formatting: Bracket Dummy Bodies

If a loop (like `while` or `for`) has an empty body, place the terminating semicolon on its own line, indented, and preferably enclose it in braces to make it visible and prevent errors.

**Instruction:** Make empty loop bodies explicit with indentation and braces.
**Sample Code:**

```java
// Bad (easy to miss semicolon):
while ((b = in.read()) != -1);

// Good:
while ((b = in.read()) != -1) {
    // Dummy body
}
```

## 22. Encapsulation: Avoid Data Clumps

Avoid passing the same set of variables together in multiple function calls. These "data clumps" suggest that the variables belong together in their own object.

**Instruction:** Identify groups of related data items that are frequently passed together and encapsulate them in a dedicated Value Object or data structure.
**Sample Code:**

```java
// Bad:
public void processAddress(String street, String city, String state, String zip) { ... }
public void saveAddress(String street, String city, String state, String zip) { ... } // Duplicated parameters

// Good:
public void processAddress(Address address) { ... }
public void saveAddress(Address address) { ... }

public class Address { // Encapsulates the data clump
    private String street;
    private String city;
    private String state;
    private String zip;
    // Constructor, getters...
}
```

## 23. Encapsulation: Hide Implementation Details

Classes and modules should hide their internal implementation details and expose only necessary interfaces. This principle applies to both data and methods.

**Instruction:** Use access modifiers (private, protected) to conceal internal state and helper methods.
**Sample Code:**

```java
// Bad:
public class Circle {
    public double radius; // Exposing internal state directly
}

// Good:
public class Circle {
    private double radius; // State is private

    public Circle(double radius) {
        this.radius = radius;
    }

    public double getRadius() { // Access via getter
        return radius;
    }

    private double calculateArea() { // Internal helper is private
        return Math.PI * radius * radius;
    }
}
```

## 24. Inheritance: Don't Inherit Constants

Using inheritance solely to share constants is a poor practice and abuses the language's scoping rules.

**Instruction:** Avoid inheriting interfaces or classes that define only constants. Use static imports or place constants in utility classes instead.
**Sample Code:**

```java
// Bad:
public interface PayrollConstants {
    public static final int TENTHS_PER_WEEK = 400;
    public static final double OVERTIME_RATE = 1.5;
}

public abstract class Employee implements PayrollConstants { ... } // Inheriting constants

// Good:
import static com.yourcompany.payroll.PayrollConstants.*; // Use static import

public class HourlyEmployee extends Employee {
    public Money calculatePay() {
        int straightTime = Math.min(tenthsWorked, TENTHS_PER_WEEK); // Access directly
        // ...
    }
}
```

## 25. Constants/Enums: Use Enums Instead of Ints

For sets of related constants, especially those representing categories or states, use enums rather than simple integer constants. Enums provide type safety and clarity.

**Instruction:** Define enums for symbolic constants that belong to a specific category or have a limited set of discrete values.
**Sample Code:**

```java
// Bad:
public static final int STATUS_PLANNED = 1;
public static final int STATUS_SCHEDULED = 2;
public static final int STATUS_COMMITTED = 3;

public void setStatus(int status) { // Accepts any int
    if (status == STATUS_PLANNED) { ... } // Prone to errors with wrong int values
}

// Good:
public enum BacklogItemStatusType {
    PLANNED, SCHEDULED, COMMITTED;
}

public void setStatus(BacklogItemStatusType status) { // Type-safe
    if (status == BacklogItemStatusType.PLANNED) { ... }
}
```

## 26. Value Objects: Design with Value Objects

Model concepts that represent descriptive aspects of the domain, lack a distinct identity, and are defined by their attributes as Value Objects.

**Instruction:** Identify domain concepts that fit the definition of a Value Object and model them as such.
**Sample Code:**

```java
// Modeling a currency amount
public final class Money { // Value Object
    private final BigDecimal amount;
    private final Currency currency;

    public Money(BigDecimal amount, Currency currency) {
        // Validation...
        this.amount = amount;
        this.currency = currency;
    }

    // Value equality based on attributes
    @Override
    public boolean equals(Object o) { /* ... */ }
    @Override
    public int hashCode() { /* ... */ }

    // Behavior related to the value
    public Money add(Money other) { /* ... */ }
    // No identity-based operations
}
```

## 27. Value Objects: Ensure Immutability

Value Objects should be immutable. Their state is set upon creation and cannot be changed thereafter.

**Instruction:** Make all fields in Value Objects `final` and ensure no methods modify the object's internal state. Return new instances for operations that would conceptually change the value.
**Sample Code:**

```java
public final class Point {
    private final int x;
    private final int y;

    public Point(int x, int y) {
        this.x = x;
        this.y = y;
    }

    // No setters

    public Point translate(int dx, int dy) { // Returns a new instance
        return new Point(this.x + dx, this.y + dy);
    }
}
```

## 28. Value Objects/Domain Primitives: Enforce Invariants in Constructor

For Value Objects and Domain Primitives, enforce all business rules, constraints, and invariants related to their value within their constructor. This ensures that once an instance is created, its value is always valid.

**Instruction:** Include validation logic in the constructor of Value Objects/Domain Primitives to guarantee valid state upon creation. Use libraries like Apache Commons Validate or similar built-in features.
**Sample Code:**

```java
import static org.apache.commons.lang3.Validate.*;

public final class Quantity {
    private static final int MIN_VALUE = 1;
    private static final int MAX_VALUE = 500;

    private final int value;

    public Quantity(final int value) {
        // Enforce invariant: quantity must be within range [MIN_VALUE, MAX_VALUE]
        inclusiveBetween(MIN_VALUE, MAX_VALUE, value, "Quantity must be between %d and %d", MIN_VALUE, MAX_VALUE);
        this.value = value;
    }

    public int value() {
        return value;
    }
}
```

## 29. Entities: Manage Identity and Mutability

Model concepts that have a distinct thread of continuity and identity over time as Entities. Entities are typically mutable, and their identity is key, not their attribute values.

**Instruction:** Model domain concepts with unique identities as Entities. Manage their mutable state through well-defined behaviors.
**Sample Code:**

```java
public class BacklogItem extends Entity { // Entity with identity
    private BacklogItemStatusType status; // Mutable state
    private String summary; // Mutable state
    // ... other attributes ...

    // Identity is managed internally (e.g., via unique ID)
    // Equality is based on identity, not value

    // Behaviors that modify state
    public void commitTo(Sprint aSprint) {
        // Validation and state change
        this.status = BacklogItemStatusType.COMMITTED;
    }

    public void updateSummary(String newSummary) {
        this.summary = newSummary;
    }
}
```

## 30. Aggregates: Define Consistency Boundaries

Group Entities and Value Objects into Aggregates, defining a consistent boundary around a set of related objects. Each Aggregate has a single Aggregate Root, which is an Entity. All external access to the Aggregate must go through the Root.

**Instruction:** Identify Aggregate boundaries in the domain model and ensure all modifications within the boundary are coordinated through the Aggregate Root.
**Sample Code:**

```java
// Example: Order Aggregate with Order as the root
public class Order extends Entity { // Aggregate Root
    private List<LineItem> lineItems; // Entities/Value Objects within the Aggregate
    private ShippingAddress shippingAddress; // Value Object

    // ... other attributes ...

    // Methods on the Aggregate Root manage consistency
    public void addLineItem(Product product, int quantity) {
        // Logic to add item, ensuring consistency rules within the Order aggregate
        LineItem newItem = new LineItem(product, quantity);
        lineItems.add(newItem);
    }

    // External access is only through Order methods
}

public class LineItem extends Entity { /* ... */ } // Part of the Order Aggregate
```

## 31. Domain Events: Model Past Occurrences

Model significant domain occurrences that represent something that happened in the past as Domain Events. Name Domain Events using past-tense verbs (e.g., `OrderPlaced`, `BacklogItemCommitted`).

**Instruction:** Define immutable classes or records to represent significant past events in the domain, using clear, past-tense names.
**Sample Code:**

```java
// Define a Domain Event representing a backlog item being committed
public final class BacklogItemCommitted { // Domain Event
    private final Date occurredOn;
    private final TenantId tenantId;
    private final BacklogItemId backlogItemId;
    private final SprintId sprintId;

    public BacklogItemCommitted(TenantId tenantId, BacklogItemId backlogItemId, SprintId sprintId) {
        this.occurredOn = new Date(); // Timestamp the event
        this.tenantId = tenantId;
        this.backlogItemId = backlogItemId;
        this.sprintId = sprintId;
    }

    // Getters for accessing event data (event is immutable)
    public Date occurredOn() { return occurredOn; }
    public TenantId tenantId() { return tenantId; }
    public BacklogItemId backlogItemId() { return backlogItemId; }
    public SprintId sprintId() { return sprintId; }
}
```

## 32. Domain Events: Publish from Aggregates

Domain Events are typically published from within the Aggregate Root when a state change occurs that is significant to other parts of the system.

**Instruction:** Design Aggregate Roots to publish relevant Domain Events after a state-changing operation successfully completes and the Aggregate is in a consistent state.
**Sample Code:**

```java
public class BacklogItem extends Entity {
    private BacklogItemStatusType status;
    private SprintId sprintId;
    // ... constructor, other methods ...

    public void commitTo(Sprint aSprint) {
        // ... validation and state change logic ...
        this.status = BacklogItemStatusType.COMMITTED;
        this.sprintId = aSprint.sprintId();

        // Publish the domain event
        DomainEventPublisher.instance().publish(
            new BacklogItemCommitted(
                this.tenantId(),
                this.backlogItemId(),
                this.sprintId()
            ));
    }
}
```

## 33. Factories: Encapsulate Complex Object Creation

Use Factories (either static factory methods or dedicated Factory classes) to encapsulate the complex logic required to create Aggregate roots or other complex objects, especially when the constructor is insufficient or involves external dependencies.

**Instruction:** Create Factories to abstract away the details of creating complex objects or families of related objects.
**Sample Code:**

```java
// Using a static factory method for complex creation logic
public class Order {
    // ... private constructor ...

    public static Order createNewOrder(Customer customer, List<Product> products) { // Static Factory Method
        // Complex logic to build the order, apply rules, etc.
        Order newOrder = new Order();
        // ... add line items, calculate initial price, etc. ...
        return newOrder;
    }
}

// Using a dedicated Factory class
public class EmployeeFactoryImpl implements EmployeeFactory { // Dedicated Factory class
    public Employee makeEmployee(EmployeeRecord r) throws InvalidEmployeeType {
        switch (r.type) {
            case COMMISSIONED:
                return new CommissionedEmployee(r);
            // ... other types ...
            default:
                throw new InvalidEmployeeType(r.type);
        }
    }
}
```

## 34. Repositories: Abstract Persistence

Use Repositories to provide a clear interface for storing and retrieving Aggregates. Repositories mediate between the domain model and the data mapping layers.

**Instruction:** Define Repository interfaces for each Aggregate Root, abstracting the details of data persistence. Implementations should handle the actual database interactions.
**Sample Code:**

```java
// Repository interface for the Order Aggregate
public interface OrderRepository {
    Order findById(OrderId orderId); // Find an Order by its ID
    void save(Order order); // Save or update an Order
    void remove(Order order); // Remove an Order
    // ... other query methods ...
}

// Implementation (in Infrastructure layer)
public class JpaOrderRepository implements OrderRepository {
    // ... JPA specific implementation ...
    @Override
    public Order findById(OrderId orderId) { /* ... */ }
    @Override
    public void save(Order order) { /* ... */ }
    // ...
}
```

## 35. Specifications: Encapsulate Validation/Selection Logic

Use Specifications to encapsulate business rules or criteria that are used for validation or selecting objects from a collection or repository.

**Instruction:** Define Specification classes for reusable validation or filtering logic that doesn't naturally belong to a single object.
**Sample Code:**

```java
// Specification for determining if an invoice is delinquent
public class DelinquentInvoiceSpecification implements Specification<Invoice> {
    private final Date currentDate;

    public DelinquentInvoiceSpecification(Date currentDate) {
        this.currentDate = currentDate;
    }

    @Override
    public boolean isSatisfiedBy(Invoice candidate) {
        int gracePeriod = candidate.customer().getPaymentGracePeriod();
        Date firmDeadline = DateUtility.addDaysToDate(candidate.dueDate(), gracePeriod);
        return currentDate.after(firmDeadline);
    }
}

// Usage:
Invoice invoice = ...;
DelinquentInvoiceSpecification spec = new DelinquentInvoiceSpecification(new Date());
if (spec.isSatisfiedBy(invoice)) {
    // ... handle delinquent invoice ...
}
```

## 36. Modules: Organize Code within Bounded Contexts

Use Modules (packages, namespaces) to organize code within a Bounded Context. Modules help manage complexity and express logical groupings of domain concepts and supporting code.

**Instruction:** Structure code into logical modules/packages based on domain areas or functional groupings within a Bounded Context.
**Sample Code:**

```java
// Package structure example for an Agile Project Management Bounded Context
com.saasovation.agilepm.domain.model.product // Domain model for Products
com.saasovation.agilepm.domain.model.backlogitem // Domain model for Backlog Items
com.saasovation.agilepm.domain.model.sprint // Domain model for Sprints
com.saasovation.agilepm.application // Application Services
com.saasovation.agilepm.port.adapter.persistence // Infrastructure/Persistence
```

## 37. Bounded Contexts: Define Explicit Boundaries

Identify and define explicit boundaries for Bounded Contexts, where a specific ubiquitous language and domain model are applied. Communication between Bounded Contexts requires explicit translation.

**Instruction:** Design systems as a collection of Bounded Contexts with clearly defined responsibilities and boundaries.
**Sample Code:**

```
// Conceptual diagram snippet (or documentation entry)
// illustrating Bounded Contexts and their relationships:

+------------------------+       +-----------------------------+
| Identity & Access CTX  |-----> | Collaboration CTX           |
| (Generic Subdomain)    |       | (Supporting Subdomain)      |
+------------------------+       +-----------------------------+
           |                                 |
           V                                 V
+-----------------------------+
| Agile Project Management CTX|
| (Core Domain)               |
+-----------------------------+
```

## 38. Bounded Contexts: Favor Asynchronous Integration

For integration between Bounded Contexts, particularly when aiming for higher autonomy, favor asynchronous communication patterns like event publishing/subscribing or messaging over synchronous calls (like RPC).

**Instruction:** Design inter-Bounded Context communication to be asynchronous where autonomy and decoupling are prioritized.
**Sample Code:**

```java
// Publishing a domain event to be consumed by another Bounded Context
public class Product {
    // ... state and methods ...

    public void scheduleRelease(ReleaseDetails details) {
        // ... update state ...
        DomainEventPublisher.instance().publish(
            new ProductReleaseScheduled(
                this.tenantId(),
                this.productId(),
                details.getScheduleDate()
            )); // Published to messaging infrastructure
    }
}
```

## 39. Functional Programming: Prefer Pure Functions

Aim for pure functions that, given the same input, always produce the same output and have no side effects (they don't modify mutable state or interact with I/O outside their return value).

**Instruction:** Write functions as pure as possible. Isolate side effects to distinct parts of the application.
**Sample Code:**

```javascript
// Pure function:
function add(a, b) {
    return a + b; // Always returns the sum, no side effects
}

// Impure function:
let total = 0;
function addToTotal(value) {
    total += value; // Modifies external state (side effect)
    return total;
}
```

## 40. Functional Programming: Use Immutable Data

Prefer using immutable data structures and objects. If a modification is needed, create a new instance with the updated value instead of changing the existing one.

**Instruction:** Design data structures and objects to be immutable where feasible, particularly for Value Objects and within functional cores.
**Sample Code:**

```clojure
;; Immutable variable (atom allows managed mutation)
(def counter (atom 0)) ; Initialize counter

;; Safely increment counter using swap! (disciplined mutation)
(swap! counter inc)
```

## 41. Functional Programming: Leverage Type System

Utilize the type system, including Algebraic Data Types (ADTs) and Pattern Matching, to represent domain concepts, enforce constraints, and make invalid states unrepresentable.

**Instruction:** Model domain concepts using types (e.g., enums, records, union types/ADTs) to encode rules and constraints. Use pattern matching for handling different cases.
**Sample Code:**

```fsharp
// Using a discriminated union (ADT) to represent Tweet types
type Tweet =
    | TextTweet of string // A tweet containing only text
    | ImageTweet of string * string // A tweet containing text and an image URL

// Using pattern matching to handle different Tweet types
let processTweet tweet =
    match tweet with
    | TextTweet text -> printfn "Processing text tweet: %s" text
    | ImageTweet (text, imageUrl) -> printfn "Processing image tweet: %s (Image: %s)" text imageUrl

// Using record types with validation for constrained values (Domain Primitives)
type UnitQuantity = UnitQuantity of int // Constraint: between 1 and 1000
let createUnitQuantity value =
    if value >= 1 && value <= 1000 then
        Ok (UnitQuantity value) // Success case
    else
        Error "Quantity must be between 1 and 1000" // Failure case
```

## 42. Input Validation: Validate Inputs for Purpose

Validate input data based on its intended purpose and context, not just its format or type. Ensure data is "valid for a purpose."

**Instruction:** Implement validation routines that check if input data is meaningful and usable within the specific process step.
**Sample Code:**

```java
public void processOrder(Order order) {
    // Generic format validation might happen earlier.
    // Here, validate if the order details are valid for processing (e.g., has items, valid shipping address).
    if (!order.hasItems() || !order.getShippingAddress().isValidForShipment()) {
        throw new InvalidOrderException("Order cannot be processed.");
    }
    // ... proceed with processing ...
}
```

## 43. Input Validation: Check Length and Range

Verify that input values, especially strings and numbers, fall within acceptable length and range limits defined by business rules or system constraints.

**Instruction:** Add explicit checks for length boundaries for strings and numeric ranges for numbers.
**Sample Code:**

```java
import static org.apache.commons.lang3.Validate.*;

public class Username {
    private static final int MIN_LENGTH = 4;
    private static final int MAX_LENGTH = 40;
    private final String value;

    public Username(final String value) {
        notBlank(value);
        final String trimmed = value.trim();
        // Explicitly check length range
        inclusiveBetween(MIN_LENGTH, MAX_LENGTH, trimmed.length(),
                         "Username length must be between %d and %d", MIN_LENGTH, MAX_LENGTH);
        // ... other validations ...
        this.value = trimmed;
    }
}
```

## 44. Input Validation: Sanitize or Reject Invalid Characters/Patterns

Sanitize or reject input that contains characters or patterns that are not expected or could be used for injection attacks (e.g., SQL injection, XML external entities).

**Instruction:** Implement strict character set or pattern validation and reject input that doesn't conform. Avoid relying solely on sanitization if input format is critical.
**Sample Code:**

```java
import static org.apache.commons.lang3.Validate.*;

public final class EmailAddress {
    public final String value;

    public EmailAddress(final String value) {
        // Validate against a strict pattern to prevent injection/malformed addresses
        matchesPattern(value.toLowerCase(), "^[a-z0-9.!#$%&'*+/=?^_`{|}~-]+@\\bhospital\\.com$", "Illegal email address format");
        // ... other validations like length ...
        this.value = value.toLowerCase();
    }
}

// Or, for a name that should only contain letters and spaces:
public final class Name {
    public Name(final String value) {
        // Validate against a pattern allowing only specified characters
        matchesPattern(value,"^[a-zA-Z ]+$", "Invalid name. Contains illegal characters.");
        // ... other validations like length ...
        this.value = value;
    }
}
```

## 45. Input Validation: Implement Fail Fast

Validate inputs and preconditions at the earliest possible point (e.g., method/constructor entry) and fail immediately if validation fails. This prevents invalid data from propagating through the system.

**Instruction:** Add validation checks at the beginning of methods or constructors and throw exceptions or return error indicators upon failure.
**Sample Code:**

```java
import static org.apache.commons.lang3.Validate.*;

public void processData(String dataId, int quantity) {
    // Fail fast validation at method entry
    notBlank(dataId, "Data ID cannot be blank");
    isTrue(quantity > 0, "Quantity must be positive");

    // ... core logic proceeds only with valid inputs ...
}
```

## 46. Security: Handle Integer Overflow/Underflow

Be aware of the range limits of integer types. Use larger integer types (e.g., 64-bit for intermediate calculations) and perform checks before casting back to smaller types to prevent overflow or underflow vulnerabilities.

**Instruction:** Use integer types sufficient for the maximum possible value in calculations. Add explicit checks when results might exceed type limits or before downcasting.
**Sample Code:**

```c
// Calculating pay: millihours * hourlycents
uint32_t millihours;
uint32_t hourlycents;
uint32_t basepay; // Result type

// Bad (prone to overflow if product exceeds UINT32_MAX):
// basepay = (millihours * hourlycents + 500) / 1000;

// Good (using 64-bit for multiplication):
uint64_t product64 = (uint64_t)millihours * hourlycents;
if (product64 > UINT32_MAX_FOR_PAY_CALCULATION) { // Check against logical maximum
    // Handle error, e.g., return 0 or throw exception
    return 0;
}
basepay = (uint32_t)(product64 + 500) / 1000; // Safe downcast after check
```

## 47. Security: Avoid Logging Sensitive Information

Never log sensitive data such as passwords, credit card numbers, social security numbers, or personal identifiable information (PII), even in development or debug logs.

**Instruction:** Implement logging practices that explicitly exclude sensitive data. Use placeholder values or read-once domain primitives if necessary.
**Sample Code:**

```java
// Bad:
logger.info("User login attempt: username={}, password={}", username, password);

// Good:
logger.info("User login attempt: username={}", username); // Exclude password

// Using a SensitiveValue domain primitive (read-once)
public final class Password {
    private char[] value;
    private boolean consumed = false;

    public synchronized char[] value() {
        validState(!consumed, "Password value has already been consumed");
        final char[] returnValue = value.clone();
        Arrays.fill(value, '0'); // Clear the internal value after read
        consumed = true;
        return returnValue;
    }

    @Override
    public String toString() { // Redact value in default string representation
        return "Password{value=*****}";
    }
}
```

## 48. Error Handling: Handle Failures Securely

Design code to handle failures gracefully and securely. Avoid exposing internal details in error messages. Use custom exceptions or result objects instead of standard exceptions for domain-specific failures.

**Instruction:** Define clear error handling strategies. Use specific domain exceptions or union types/result objects to communicate failure reasons explicitly without leaking internal information.
**Sample Code:**

```java
// Using a custom domain exception
public class AccountNotFound extends RuntimeException {
    public AccountNotFound(AccountNumber accountNumber, Customer customer) {
        super(format("Account %s not found for customer %s", accountNumber, customer.getId())); // Log context internally, don't expose raw IDs widely
    }
}

// Usage:
public Account fetchAccountFor(Customer customer, AccountNumber accountNumber) {
    // ... find account ...
    return accountDatabase.selectAccountsFor(customer).stream()
        .filter(account -> account.number().equals(accountNumber))
        .findFirst()
        .orElseThrow(() -> new AccountNotFound(accountNumber, customer)); // Throw specific domain exception
}

// Using a Result object (functional style)
public final class Result<T, E> { // Generic Result type
    // ... success/failure states ...
}

public Result<MoneyTransferReceipt, MoneyTransferFailureReason> transferFunds(...) {
    // ... check funds ...
    if (!sourceAccount.hasSufficientFunds(amount)) {
        return Result.failure(MoneyTransferFailureReason.INSUFFICIENT_FUNDS); // Return failure object
    }
    // ... execute transfer ...
    return Result.success(receipt); // Return success object
}
```

## 49. Error Handling: Specify Timeouts for External Calls

Always set explicit timeouts for network calls and other interactions with external systems or resources. Infinite timeouts can lead to resource exhaustion (e.g., memory hogs, blocked threads).

**Instruction:** Configure timeouts for all operations that involve waiting for an external response.
**Sample Code:**

```java
// Example using a library that supports timeouts
// HttpClient httpClient = HttpClient.newBuilder()
//    .connectTimeout(Duration.ofSeconds(5)) // Set connection timeout
//    .build();
//
// HttpRequest request = HttpRequest.newBuilder()
//    .uri(URI.create("http://external.service/api/resource"))
//    .timeout(Duration.ofSeconds(10)) // Set request timeout
//    .build();
//
// try {
//    HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
//    // ... process response ...
// } catch (IOException | InterruptedException | TimeoutException e) {
//    // Handle timeout or other network errors
// }
```

## 50. Logging: Use Appropriate Levels

Use different log levels (TRACE, DEBUG, INFO, NOTICE, WARN, ERROR, FATAL) to categorize the severity and importance of log messages. This allows for filtering logs based on the operational context (development, production).

**Instruction:** Select the log level that best matches the nature and severity of the event being logged.
**Sample Code:**

```java
// Using different log levels
logger.debug("Processing request for user: {}", userId); // Detailed info for debugging
logger.info("User {} successfully logged in.", username); // Standard event
logger.warn("Database query took longer than expected: {}ms", duration); // Potential issue
logger.error("Failed to process order {}: {}", orderId, e.getMessage(), e); // Error with exception details
logger.fatal("Application failed to bind to port {}. Shutting down.", port); // Critical, unrecoverable error
```

## 51. Logging: Include Sufficient Context

Log messages should contain enough context to be understandable and useful for troubleshooting without requiring access to the source code or external systems. Include relevant IDs, parameters, and outcomes.

**Instruction:** Ensure log messages provide the necessary information (e.g., user ID, transaction ID, input parameters, error details) to diagnose issues.
**Sample Code:**

```java
// Bad (lacks context):
logger.error("Transaction failed.");

// Good (with context):
logger.error("Transaction failed for user {} (ID: {}) with reason: {}",
             transactionId, userId, reason);

// Example using structured logging (key=value pairs)
logger.info("finished operation",
            "result", "success",
            "elapsed", elapsedTime.toMillis(),
            "operation_name", "process_widget",
            "widget_id", widget.getId());
```

## 52. Logging: Avoid Vendor Lock-in

Design the logging strategy to be independent of a specific logging framework or vendor. Use logging facades or interfaces to decouple application code from the underlying logging implementation.

**Instruction:** Use a logging facade (like SLF4J in Java) or define a custom logger interface to allow changing the logging backend easily.
**Sample Code:**

```java
// Using SLF4J facade (Java)
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class MyService {
    private static final Logger logger = LoggerFactory.getLogger(MyService.class);

    public void doSomething() {
        logger.info("Service is doing something.");
    }
}
// The actual logging implementation (e.g., Logback, Log4j) can be configured externally.
```

## 53. Unit Testing: Write Unit Tests

Write tests that verify the behavior of small, isolated units of code (e.g., a class or a method). Follow the Red-Green-Refactor cycle of Test-Driven Development (TDD): write a failing test (Red), write the minimum code to make it pass (Green), and then refactor the code while keeping the test green.

**Instruction:** Implement unit tests for code components. Use a testing framework and follow the TDD cycle.
**Sample Code:**

```java
// TDD Cycle Example (using JUnit)

// 1. Red: Write a failing test
@Test
public void testMultiplication() {
    Dollar five = new Dollar(5);
    Dollar product = five.times(2);
    assertEquals(10, product.amount); // Fails because times() doesn't multiply or return correct value
}

// 2. Green: Write minimum code to pass the test
public class Dollar {
    int amount;
    public Dollar(int amount) {
        this.amount = amount;
    }
    public Dollar times(int multiplier) {
        // Hardcode to pass this specific test
        return new Dollar(10);
    }
}

// 3. Refactor: Add another test ($5 * 3 = $15) and generalize the code
@Test
public void testMultiplication() {
    Dollar five = new Dollar(5);
    assertEquals(new Dollar(10), five.times(2)); // Using value equality
    assertEquals(new Dollar(15), five.times(3)); // New test case
}

public class Dollar {
    int amount;
    public Dollar(int amount) {
        this.amount = amount;
    }
    public Dollar times(int multiplier) {
        // Generalized implementation
        return new Dollar(this.amount * multiplier);
    }
    // Need equals() for assertion to work correctly (test for value equality)
    @Override
    public boolean equals(Object object) {
        Dollar dollar = (Dollar) object;
        return amount == dollar.amount;
    }
}
```

## 54. Unit Testing: Use Descriptive Test Names

Test names should clearly describe the specific behavior being tested, including the conditions and expected outcome.

**Instruction:** Write test method names that function as documentation, explaining what the code does under specific circumstances. Use a consistent naming convention (e.g., `Unit_Scenario_ExpectedOutcome`).
**Sample Code:**

```java
// Bad:
public void testProcess() { ... }

// Good:
public void Delivery_with_a_past_date_is_invalid() { ... }

// Using a common convention: UnitUnderTest_StateUnderTest_ExpectedBehavior
public void EmailAddress_withInvalidFormat_throwsIllegalArgumentException() { ... }
public void Account_withSufficientFunds_allowsTransfer() { ... }
```

## 55. Unit Testing: Test Edge Cases and Boundaries

Include test cases that specifically target the boundaries of valid input ranges, edge cases, and error conditions to ensure robust behavior.

**Instruction:** Design tests that check minimum, maximum, and off-by-one values for ranges, as well as invalid or extreme inputs.
**Sample Code:**

```java
import static org.apache.commons.lang3.StringUtils.repeat;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.DynamicTest.dynamicTest;
import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import java.util.stream.Stream;

class EmailAddressTest {
    @TestFactory
    Stream<DynamicTest> should_be_rejected() {
        return Stream.of(
            "a@hospital.com", // Below minimum length
            repeat("X", 64) + "@something.com", // Domain not hospital.com
            repeat("X", 65) + "@hospital.com", // Local part too long
            repeat("X", 65) + "@hospital.com.", // Invalid TLD
            ".jane@hospital.com", // Starts with invalid character
            "jane.@hospital.com", // Contains invalid sequence
            repeat("X", 65) + "@hospital.com.", // Total length exceeds maximum
            repeat("X", 10000) // Extreme length input
        )
        .map(input -> dynamicTest("Rejected: " + input,
            () -> assertThrows(IllegalArgumentException.class, // Or specific domain exception
                () -> new EmailAddress(input))
        ));
    }

    @TestFactory
    Stream<DynamicTest> should_be_accepted() {
        return Stream.of(
            "aa@hospital.com", // Minimum valid length
            repeat("X", 64) + "@hospital.com", // Maximum local part length
            repeat("X", 10) + "@hospital.com" // Within valid length range
        )
        .map(input -> dynamicTest("Accepted: " + input,
            () -> assertDoesNotThrow(() -> new EmailAddress(input))
        ));
    }
}
```

## 56. Unit Testing: Use Test Doubles

Employ Test Doubles (Stubs, Mocks, Spies, Fakes, Dummy Objects) to isolate the code under test from its dependencies. This allows controlling the behavior of dependencies and verifying interactions, making tests faster, more reliable, and independent.

**Instruction:** Use Test Doubles for external dependencies (databases, services, system clock, etc.) to control the test environment and verify SUT interactions.
**Sample Code:**

```java
// Using a Test Stub for TimeProvider
public class TimeDisplayTest {
    @Test
    public void testDisplayCurrentTime_AtMidnight() {
        // Fixture setup: Create a Test Stub for TimeProvider
        TimeProvider testStub = new TimeProvider() {
            @Override
            public Calendar getTime() {
                // Return a fixed time (midnight)
                Calendar cal = new GregorianCalendar();
                cal.set(Calendar.HOUR_OF_DAY, 0);
                cal.set(Calendar.MINUTE, 0);
                return cal;
            }
        };

        // Instantiate the SUT and inject the stub
        TimeDisplay sut = new TimeDisplay();
        sut.setTimeProvider(testStub); // Assuming Dependency Injection

        // Exercise SUT
        String result = sut.getCurrentTimeAsHtmlFragment();

        // Verify outcome
        assertEquals("<span class=\"tinyBoldText\">Midnight</span>", result);
    }
}
```

## 57. Unit Testing: Manage Test Data

Use Creation Methods (Test Data Builders, Object Mother, Anonymous Factory) to create complex test data needed for fixture setup. This hides the "necessary but irrelevant" details of data creation, keeping tests clean and focused on the behavior under test.

**Instruction:** Encapsulate complex test data creation logic in dedicated methods or classes to simplify test fixtures.
**Sample Code:**

```java
public class FlightTest {
    private int uniqueFlightNumber = 2000; // Helper for unique numbers

    // Creation Method for a standard flight
    public Flight createAnonymousFlight() {
        Airport departure = new Airport("Calgary", "YYC");
        Airport destination = new Airport("Toronto", "YYZ");
        return new Flight(new BigDecimal(uniqueFlightNumber++), departure, destination);
    }

    // Creation Method for a cancelled flight
    public Flight createAnonymousCancelledFlight() {
        Flight flight = createAnonymousFlight();
        flight.cancel();
        return flight;
    }

    @Test
    public void testStatus_cancelled() {
        // Setup using Creation Method
        Flight flight = createAnonymousCancelledFlight();

        // Exercise and Verify
        assertEquals(FlightState.CANCELLED, flight.getStatus());
    }
}
```

## 58. Microbenchmarking: Measure Performance of Small Code Units

Use microbenchmarks to measure the performance (latency, allocations) of small, isolated pieces of code, typically individual functions or methods.

**Instruction:** Write microbenchmarks for performance-critical functions using the language's native benchmarking tools.
**Sample Code:**

```go
// Go Microbenchmark Example (in a file ending in _test.go)

package main

import (
	"testing"
)

// Assume a function Sum(filename string) (int64, error) exists
// and a file "numbers.txt" with many integers is available.

func BenchmarkSum(b *testing.B) {
	// Reset timer to exclude setup time
	b.ResetTimer()

	// The loop runs the function b.N times
	for i := 0; i < b.N; i++ {
		// Call the function being benchmarked
		_, err := Sum("numbers.txt")
		if err != nil {
			b.Fatal(err) // Fail benchmark on error
		}
	}
}

/*
// To run this benchmark:
go test -bench=. -benchmem

// Example Output:
// BenchmarkSum-8   	      10	 197999371 ns/op	 30400872 B/op	   1600001 allocs/op
*/
```

## 59. Macrobenchmarking: Measure Performance at System Level

Use macrobenchmarks (integration tests, end-to-end tests, load tests) to measure the performance of the application or system as a whole, under realistic load conditions.

**Instruction:** Design and run macrobenchmarks to evaluate system performance, throughput, latency, and resource usage under simulated production load. Use tools like k6 or custom e2e frameworks.
**Sample Code:**

```go
// Go Macrobenchmark (e2e) Example using efficientgo/e2e
// This example requires a running instance of the application under test (e.g., 'labeler').
// Assumes the application is configured to expose Prometheus metrics.

package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/efficientgo/e2e"
	e2einteractive "github.com/efficientgo/e2e/interactive" // For interactive mode
	e2emonitoring "github.com/efficientgo/e2e/monitoring" // For monitoring setup
	"github.com/efficientgo/core/testutil"
)

func TestLabeler_Load(t *testing.T) {
	// Create an e2e environment (e.g., using Docker Compose)
	e, err := e2e.NewDockerEnvironment("labeler_load_test")
	testutil.Ok(t, err)
	t.Cleanup(e.Close) // Ensure environment is cleaned up

	// Define the application service to test (assuming it's defined in the e2e environment)
	labeler := e.Runnable("labeler_app") // Replace with actual service name

	// Define a load testing tool (k6) runnable
	k6 := e.Runnable("k6_load").Init(e2e.StartOptions{
		Image: "grafana/k6:0.39.0", // Use a k6 Docker image
	})

	// Set up monitoring (e.g., Prometheus, Grafana) for the environment
	mon := e2emonitoring.New().WithPrometheus().WithGrafana()
	e.Add(mon)

	// Start services and wait for them to be ready
	testutil.Ok(t, e2e.StartAndWaitReady(labeler, k6, mon.Prometheus(), mon.Grafana()))

	// Run k6 load script against the application
	// This is a simplified example. A real script would be more complex.
	loadScript := `
		import http from 'k6/http';
		import { check, sleep } from 'k6';

		export default function () {
			const res = http.get('http://labeler_app:8080/status'); // Replace with actual endpoint
			check(res, { 'status was 200': (r) => r.status == 200 });
			sleep(1); // Control request rate
		}
	`
	// Execute k6 script for a defined duration and number of users
	testutil.Ok(t, k6.Exec(e2e.NewCommand(
		"/bin/sh", "-c", fmt.Sprintf(`cat <<EOF | k6 run -u 5 -d 1m -
%s
EOF`, loadScript),
	)))

	// Optionally, enter interactive mode to inspect metrics in Grafana
	// e2einteractive.Run(e)

	// After the test, analyze metrics from Prometheus/Grafana to assess performance.
}

/*
// To run this macrobenchmark:
// Requires Docker and docker-compose setup for the e2e environment definition.
go test -tags=e2e -run TestLabeler_Load .
*/
```

## 60. Performance: Formalize Efficiency Requirements

Define clear, measurable, and resource-aware efficiency requirements (RAER) for operations or workflows. These requirements should specify limits on latency, CPU usage, memory, disk I/O, etc., often related to the size of the input dataset (e.g., using complexity functions).

**Instruction:** Work with stakeholders to define formal efficiency requirements before optimization efforts begin.
**Sample Code:**

```markdown
## Efficiency Requirements (RAER)

**Program:** "Data Processing Service"
**Operation:** "Process large input file"
**Dataset:** "File containing N records"

*   **Maximum Latency:** O(N log N) or specify percentile (e.g., 95th percentile < 5 seconds)
*   **CPU Cores Limit:** <= 4 cores for single file processing
*   **Memory Limit:** <= 1 GB + 100 bytes/record (approximate O(N) space complexity)
*   **Disk Read Throughput:** >= 100 MB/s
*   **Network Egress:** <= 50 MB/s
```

## 61. Performance: Use Profiling to Identify Bottlenecks

When performance goals are not met, use profiling tools to identify where the application spends most of its time (CPU), allocates most memory (Heap), or where goroutines/threads are blocked.

**Instruction:** Capture and analyze profiles (CPU, Heap, Goroutine, Blocking) for the running application under load to pinpoint performance bottlenecks. Use tools like Go's pprof.
**Sample Code:**

```bash
 Commands for capturing Go profiles

 Start application with pprof agent enabled (e.g., import _ "net/http/pprof")
 application_command &

 Capture CPU profile for 30 seconds
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.pprof

 Capture Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap > heap.pprof

 Capture Goroutine profile
go tool pprof http://localhost:6060/debug/pprof/goroutine > goroutine.pprof

 Analyze profile (e.g., generate a flame graph)
go tool pprof -http=:8080 cpu.pprof # Opens web UI

 Analyze heap profile showing top allocations by size
go tool pprof -http=:8080 -alloc_space heap.pprof # Opens web UI showing memory allocations
```

## 62. Performance: Monitor Key Metrics

Instrument the application to expose key performance metrics (latency, throughput, CPU usage, memory usage, error rates) and monitor them continuously, especially in production and during load tests.

**Instruction:** Integrate metrics collection (e.g., using Prometheus client libraries) into the application for essential performance indicators.
**Sample Code:**

```go
// Go Metrics Example (using prometheus/client_golang)

package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Define a histogram to record request latency, categorized by endpoint and result
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",
			Help: "Duration of HTTP requests.",
			Buckets: []float64{0.001, 0.01, 0.1, 1, 10}, // Latency buckets
		},
		[]string{"endpoint", "result"}, // Labels for categorization
	)
)

func main() {
	// Register the HTTP handler for metrics
	http.Handle("/metrics", promhttp.Handler())

	// Example of using the metric in an HTTP handler
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		result := "success" // Assume success by default
		defer func() {
			// Observe the request duration, adding labels
			httpRequestDuration.WithLabelValues("/process", result).Observe(time.Since(start).Seconds())
		}()

		// ... handle request ...
		err := processRequest(r)
		if err != nil {
			result = "error" // Update result label on error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// ... write response ...
	})

	// ... start HTTP server ...
}

// Example of runtime metrics (Go)
import (
	"fmt"
	"runtime"
	"runtime/metrics"
)

var memMetrics = []metrics.Sample{
	{Name: "/gc/heap/allocs:bytes"},
	{Name: "/memory/classes/heap/objects:bytes"},
}

func printMemRuntimeMetric() {
	runtime.GC() // Trigger GC to get up-to-date heap stats
	metrics.Read(memMetrics)

	fmt.Println("Total bytes allocated:", memMetrics.Value.Uint64())
	fmt.Println("In-use bytes:", memMetrics.Value.Uint64())
}
```

## 63. Performance: Address Resource Leaks

Identify and fix resource leaks, such as forgotten goroutines, unclosed file handles, or database connections. Resource leaks can lead to gradual performance degradation and eventual system instability or crashes.

**Instruction:** Ensure all resources (files, network connections, goroutines started internally, database handles) are properly closed or released when no longer needed, especially in the face of errors or cancellations. Use mechanisms like `defer` or context cancellation.
**Sample Code:**

```go
// Go Resource Leak Example (Goroutine leak)

// Bad (goroutine might leak if context is cancelled before respCh receives):
func HandleVeryWrong(w http.ResponseWriter, r *http.Request) {
	respCh := make(chan int) // Unbuffered channel

	go func() {
		// This computation might take a long time
		respCh <- ComplexComputationWithCtx(r.Context()) // Will block here if HandleVeryWrong returns early
	}()

	select {
	case <-r.Context().Done(): // Request cancelled
		// Handle cancellation, but the goroutine above is still running (leaked)
		return
	case resp := <-respCh: // Computation finished
		// Process response
		_, _ = w.Write([]byte(strconv.Itoa(resp)))
		return
	}
}

// Good (use buffered channel or ensure read):
func HandleLessWrong(w http.ResponseWriter, r *http.Request) {
	respCh := make(chan int, 1) // Buffered channel (size 1)

	go func() {
		respCh <- ComplexComputationWithCtx(r.Context())
	}()

	select {
	case <-r.Context().Done():
		return // Goroutine can complete and write to buffered channel without blocking
	case resp := <-respCh:
		_, _ = w.Write([]byte(strconv.Itoa(resp)))
		return
	}
}

// Alternative Good (always read from channel after select):
func HandleBetter(w http.ResponseWriter, r *http.Request) {
	respCh := make(chan int)

	go func() {
		respCh <- ComplexComputationWithCtx(r.Context())
	}()

	select {
	case <-r.Context().Done():
		// Ensure the goroutine completes and its result is consumed
		// A timeout on the context might be needed for ComplexComputationWithCtx
		// Await goroutine completion here if possible, or rely on context cancellation within it.
		return
	case resp := <-respCh:
		_, _ = w.Write([]byte(strconv.Itoa(resp)))
		return
	}
    // Ensure respCh is read from even if context is done first
    go func() { <-respCh }() // Consume the result asynchronously
}

// Go File Handle Leak Example
func processFile(filename string) error {
	// Bad: File might not be closed on error
	// f, err := os.Open(filename)
	// if err != nil {
	// 	return err
	// }
	// // ... process file ...
	// return f.Close() // Close only on successful return

	// Good: Use defer to ensure close
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close() // Ensures f.Close() is called when the function returns

	// ... process file ...
	return nil
}
```

## 64. Performance: Optimize Based on Data

Base optimization efforts on empirical data from profiling and benchmarking, rather than assumptions or guesses. Identify the most impactful bottlenecks before attempting optimizations.

**Instruction:** Use profiling and benchmarking results to guide where performance optimization efforts should be focused.
**Sample Code:**

```
 Workflow steps based on data:
 1. Identify efficiency problem (e.g., using metrics, user reports).
 2. Define or confirm efficiency goals (RAER).
 3. Write benchmarks/tests that reproduce the problematic behavior under load.
 4. Use profiling tools (pprof, etc.) to identify specific functions/code paths consuming most resources.
 5. Focus optimization efforts on the identified bottlenecks.
 6. Re-run benchmarks/profiles to measure the impact of optimizations.
 7. Stop optimizing when goals are met or further effort cost outweighs benefits.
```

## 65. Data Modeling: Choose Appropriate Data Models

Select data models (relational, document, graph, etc.) that are suitable for the specific needs of the application's domain and query patterns.

**Instruction:** When designing data storage and access, select a data model that aligns with the structure of the domain and the required query capabilities.
**Sample Code:**

```sql
-- Relational model example (simplified Twitter timeline)
CREATE TABLE users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    -- ...
);

CREATE TABLE tweets (
    tweet_id BIGINT PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(user_id),
    text TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    -- ...
);

CREATE TABLE follows (
    follower_id BIGINT NOT NULL REFERENCES users(user_id),
    followee_id BIGINT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (follower_id, followee_id)
);

-- Query example for a user's home timeline (simplified fan-out approach)
CREATE TABLE home_timelines (
    user_id BIGINT NOT NULL REFERENCES users(user_id),
    tweet_id BIGINT NOT NULL REFERENCES tweets(tweet_id),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY (user_id, timestamp DESC, tweet_id) -- Composite index for efficient reads
);
```

## 66. Data Modeling: Represent Data with Type Safety

Use the type system to represent data structures and ensure type safety throughout the application, especially when interacting with data stores or external systems.

**Instruction:** Define explicit types for data structures, particularly those coming from or going to external systems, to ensure data integrity and prevent errors.
**Sample Code:**

```go
// Define structs for data received from/sent to an API or database
type UserProfile struct {
	ID        string    `json:"id"`         // Explicit ID type
	Username  string    `json:"username"`   // Explicit string type
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"` // Explicit time type
	IsActive  bool      `json:"is_active"`  // Explicit boolean type
	Age       int       `json:"age,omitempty"` // Optional integer
}

// Using Domain Primitives for enhanced type safety and validation
type UserID string // UserID is a string, but its type conveys meaning
type Email string  // Email is a string with validation constraints

func NewEmail(value string) (Email, error) {
	// Validate email format here
	if !isValidEmail(value) {
		return "", fmt.Errorf("invalid email format")
	}
	return Email(value), nil
}

// Function signature using domain primitives for clarity and safety
func GetUserByEmail(email Email) (*UserProfile, error) { /* ... */ }
```

## 67. Concurrency: Use Context for Cancellation/Timeouts

When using concurrency, use a `Context` object (or equivalent) to manage cancellation signals and deadlines/timeouts across goroutines or threads. This helps prevent resource leaks and ensures coordinated shutdown.

**Instruction:** Pass a `Context` object to concurrent operations that might need to be cancelled or time out. Check the context's `Done()` channel or `Err()` method within the concurrent operation.
**Sample Code:**

```go
// Go Concurrency with Context Example
func processRequestWithTimeout(ctx context.Context, requestData string) (string, error) {
	// Create a context with a timeout for this specific operation
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel() // Cancel the context when the function returns

	resultCh := make(chan string)
	errorCh := make(chan error)

	go func() {
		// Simulate a potentially long-running operation
		result, err := performHeavyProcessing(ctxWithTimeout, requestData)
		if err != nil {
			errorCh <- err
			return
		}
		resultCh <- result
	}()

	select {
	case <-ctxWithTimeout.Done(): // Operation timed out or parent context cancelled
		return "", ctxWithTimeout.Err()
	case result := <-resultCh: // Operation completed successfully
		return result, nil
	case err := <-errorCh: // Operation failed
		return "", err
	}
}

func performHeavyProcessing(ctx context.Context, data string) (string, error) {
	// Check context periodically or before blocking operations
	select {
	case <-ctx.Done():
		return "", ctx.Err() // Return early if context is cancelled
	default:
		// Continue processing
	}

	// Simulate work
	time.Sleep(3 * time.Second)

	// Check context again after work
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
		// Final result
		return "Processed: " + data, nil
	}
}
```

## 68. Distributed Systems: Understand Quorum and Consistency

When designing distributed systems, understand quorum requirements (w + r > n) for read and write operations to ensure consistency under node failures or network partitions.

**Instruction:** Design distributed data operations considering quorum size (w=writes, r=reads, n=replicas) to balance consistency, availability, and performance based on application requirements.
**Sample Code:**

```
// Configuration example for a distributed key-value store with 5 replicas (n=5)
// Balancing strong consistency and availability:
write_quorum: 3  # w=3
read_quorum: 3   # r=3
 w + r = 3 + 3 = 6 > n=5. Can tolerate (n - w) = 5 - 3 = 2 unavailable nodes for writes.
 Can tolerate (n - r) = 5 - 3 = 2 unavailable nodes for reads.

// Prioritizing availability over strong consistency (potential for stale reads):
write_quorum: 2  # w=2
read_quorum: 2   # r=2
 w + r = 2 + 2 = 4 <= n=5. Can tolerate 3 unavailable nodes for writes, 3 for reads.
 Reads might return stale data as the write might not have reached the nodes read from.
```

## 69. Distributed Systems: Implement Fencing Tokens for Write Ordering

Use fencing tokens (monotonically increasing numbers issued by a lock service) when coordinating writes from multiple clients to prevent incorrect overwrites due to delayed or retried operations.

**Instruction:** Implement a mechanism using fencing tokens whenever multiple clients can write to the same resource in a way that could lead to race conditions or lost updates.
**Sample Code:**

```
// Simplified logic example

// Client 1:
// 1. Acquires lease/lock for resource X from lock service.
//    Lock service returns lease with fencing token T=33.
// 2. Client 1 reads value V_old for resource X.
// 3. Client 1 prepares new value V_new.
// 4. Client 1 pauses/gets delayed.

// Client 2:
// 1. Acquires lease/lock for resource X.
//    Lock service returns lease with fencing token T=34 (token is higher).
// 2. Client 2 reads value V_old'.
// 3. Client 2 prepares new value V_new'.
// 4. Client 2 sends write request for V_new' to storage service, including token T=34.
//    Storage service accepts write for X with token 34.

// Client 1 (resumes):
// 5. Client 1 sends write request for V_new to storage service, including its old token T=33.
//    Storage service receives write for X with token 33.
//    Storage service compares received token (33) with the latest token seen for X (34).
//    Since 33 < 34, the storage service rejects Client 1's write request.
// This prevents Client 1's stale write from overwriting Client 2's more recent write.
```

## 70. Distributed Systems: Aim for Exactly-Once Processing Semantics

Design message processing or event handling in distributed systems to achieve "effectively-once" semantics where possible, ensuring that each message or event is processed exactly once despite potential retries or failures. This often involves idempotent operations.

**Instruction:** Implement message consumers or event handlers to be idempotent, meaning that applying the same operation multiple times produces the same result as applying it once. This enables safe retries.
**Sample Code:**

```
// Example: Processing a "ProcessOrder" command in a system that might redeliver messages.

// Command payload includes a unique Order ID.
// Command: { "type": "ProcessOrder", "orderId": "UUID-12345", "items": [...] }

// Processing logic (Idempotent):
func HandleProcessOrderCommand(command ProcessOrderCommand) {
    // Use the unique order ID as an idempotency key.
    orderId := command.OrderID

    // Check if this order has already been processed.
    if OrderProcessingStatus.IsProcessed(orderId) {
        log.Printf("Order %s already processed. Skipping.", orderId)
        return // Operation is idempotent; just return if already done.
    }

    // Perform the actual order processing logic.
    // ... process items, update database, etc. ...

    // Mark the order as processed *only after* the processing is complete.
    // This step itself needs to be atomic with the state change or be part of a transaction.
    OrderProcessingStatus.MarkAsProcessed(orderId)

    log.Printf("Order %s processed successfully.", orderId)
}
```

## 71. Infrastructure: Use Standard Logging Libraries

Do not implement custom logging mechanisms (e.g., writing directly to files, managing rotation). Use standard, well-established logging libraries that integrate with system APIs and allow configuration for various environments.

**Instruction:** Always use a standard logging library (e.g., Logback/Log4j/SLF4J in Java, standard `log` or third-party libraries in Go, etc.) for all application logging.
**Sample Code:**

```java
// Example using a standard library (Log4j 2 with SLF4J)

// In pom.xml or build.gradle, include SLF4j API and Log4j 2 implementation dependencies.

// In src/main/resources/log4j2.xml, configure the logging output (console, file, network, etc.),
// levels, and formats.

// In Java code:
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class MyApp {
    // Get a logger instance for this class
    private static final Logger logger = LoggerFactory.getLogger(MyApp.class);

    public static void main(String[] args) {
        // Use the logger to log messages at different levels
        logger.info("Application started.");
        try {
            // ... application logic ...
        } catch (Exception e) {
            logger.error("An unexpected error occurred.", e);
        }
        logger.info("Application finished.");
    }
}
```

## 72. Infrastructure: Process Changes with Streams

Leverage change data capture (CDC) mechanisms like database streams (e.g., DynamoDB Streams) or message queues to react to data changes in real-time or near-real-time, enabling asynchronous processing and propagation of events.

**Instruction:** Design systems to consume data change streams or messages from a message bus to trigger subsequent actions or propagate changes to other services.
**Sample Code:**

```
# Example using AWS DynamoDB Streams
#
# 1. Enable DynamoDB Streams on the Event Table.
#    Configuration: Stream view type (e.g., NEW_AND_OLD_IMAGES).
#
# 2. Configure an AWS Lambda function to be triggered by the DynamoDB Stream.
#    The Lambda function receives batches of records representing item modifications (INSERT, MODIFY, REMOVE).
#
# 3. Lambda Function Logic:
#    - Read records from the stream event.
#    - For each record (representing an event item saved in the event table):
#      - Identify the type of change (e.g., INSERT).
#      - Extract the new item data (the event payload).
#      - Process the event (e.g., update a read model, send a notification, apply TTL).
#      - Example: If a snapshot item is written (INSERT), the Lambda can find corresponding old event items
#        and set their TTL attribute based on configuration (e.g., 48 hours) to manage storage costs.
#      - Example: Propagate the event message to other subscribers via a message queue.
#
# This pattern allows decoupling the write to the event store from subsequent actions.
```