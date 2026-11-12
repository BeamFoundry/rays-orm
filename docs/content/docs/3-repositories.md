---
title: "3. Auto-Generated Repositories"
weight: 3
---

# Auto-Generated Repositories

Rays ORM eliminates CRUD boilerplate by automatically implementing interface methods based on their function signatures.


## 1. The Base Repository

Rays provides `OrmRepository[T, ID]`, which automatically provides implementations for the standard operations
defined by the `OrmRepositoryInterface`:

* **`Query(queries ...OrmQueryFunc) *gorm.DB`**: Lets you execute queries on `<T>`, calling gorm `Model() *gorm.DB` first.
* **`FindById(id ID) Optional[*T]`**: Lets you find an optional model pointer by its Id of type `<ID>`.
* **`Delete(object *T) Error`**: Lets you delete a model `<T>` pointer, previously returned.
* **`DeleteById(id ID) Error`**: Lets you delete a model using its id of type `<ID>`.
* **`Save(object *T) (*T, Error)`**: Saves a model `<T>` pointer, returning the pointer and/or an `Error`.
* **`SaveAll(object []*T) ([]*T, Error)`**: Saves a slice of `<T>` pointers, returning it and/or an `Error`.

## 2. The Query Builder Syntax

To add custom queries, you define an unexported `func` field on your struct. During container wiring, Rays ORM parses
the name of the function field and **automatically generates the GORM query code** to fulfill it.

### Anatomy of a Method Name

A query method name must follow a strict grammatical structure:
`[Prefix][Subject]By[Conditions]OrderBy[SortRules]`

1.  **Prefix:** Must be `find`, `Find`, `query`, `Query`, `list`, `List`, `stream`, or `Stream`.
    * Use `find`/`Find` or `query`/`Query` for single entities (`*T`, `Optional[*T]`) or collections (`[]T`).
    * Use `list`/`List` for retrieving multiple entities into a slice (`[]T`).
    * Use `stream`/`Stream` for returning lazy, Java 8-style streams (`Stream[T]`).
2.  **Subject:** (Optional) Capitalized words describing what is being found (e.g., `ActiveUsers`).
3.  **Separator:** Must include `By`.
4.  **Conditions:** Field names followed by an optional Operator, separated by `And` or `Or`.
5.  **OrderBy:** (Optional) The keyword `OrderBy` followed by one or more sorting rules. Each rule consists of a field name
and an optional direction (`Asc` or `Desc`, defaulting to `Asc`). Multiple sorting rules are chained directly without separators.

*Examples:*
* `findUsersByAgeGreaterThanAndNameLikeOrderByNameAscCreatedAtDesc`
* `listUsersByStatusAndAgeGreaterThan(string, uint) ([]User, Error)`
* `streamAllByStatusOrderByCreatedAtDesc(string) Stream[User]`

### Supported Operators

| Operator Keyword | Equivalent SQL | Arguments Required | Example |
| :--- | :--- | :--- | :--- |
| *(None)*, `Is`, `Equals` | `=` | 1 (Field Type) | `findByName(string)` |
| `Not` | `!=` | 1 (Field Type) | `findByStatusNot(string)` |
| `GreaterThan`, `After` | `>` | 1 (Field Type) | `findByAgeGreaterThan(int)` |
| `LessThan`, `Before` | `<` | 1 (Field Type) | `findByCreatedAtBefore(time.Time)` |
| `GreaterThanOrEqual` | `>=` | 1 (Field Type) | `findByAgeGreaterThanOrEqual(int)` |
| `LessThanOrEqual` | `<=` | 1 (Field Type) | `findByAgeLessThanOrEqual(int)` |
| `IsNull`, `Null` | `IS NULL` | 0 | `findByDeletedAtIsNull()` |
| `IsNotNull`, `NotNull` | `IS NOT NULL` | 0 | `findByEmailNotNull()` |
| `Between` | `BETWEEN x AND y` | 2 (Field Type) | `findByAgeBetween(int, int)` |
| `Like` | `LIKE` | 1 (String) | `findByNameLike(string)` |
| `NotLike` | `NOT LIKE` | 1 (String) | `findByNameNotLike(string)` |
| `In` | `IN (...)` | 1 (Slice of Field Type) | `findByRoleIn([]string)` |
| `NotIn` | `NOT IN (...)` | 1 (Slice of Field Type) | `findByStatusNotIn([]int)` |

### Security: SQL Injection Prevention

It is crucial to note that Rays ORM does **not** concatenate raw strings to build queries. Under the framework hood, your method
names are translated into **parameterized queries** using `?` placeholders (e.g., `.Where("name = ?", name)`).

The arguments provided to your function are passed directly to GORM's execution engine. This ensures that your application is
entirely protected against SQL injection attacks by default.

### Return Types

Your generated function can define the following return signatures:
* **Single Entity:** `*T` or `Optional[*T]`
* **Multiple Entities:** `[]T` or `Stream[T]` (Java 8-style lazy Stream from `github.com/BeamFoundry/rays/pkg/util/stream`)
* **Query Builder:** `*gorm.DB` (Returns the query before mapping to an object)
* **Error Handling:** You may append `Error` as a second return parameter to any of the above to catch database faults.

When returning a `Stream[T]`, you can leverage functional transformations like `.Filter(...)`, `.Map(...)`, `.SortedBy(...)`, and `.ToSlice()`:

```go
stream, err := userRepo.StreamUsersByAgeGreaterThan(18)
if err != nil {
    return err
}

adultNames := stream.
    Filter(func(u User) bool { return u.Age >= 21 }).
    Map(func(u User) string { return u.Name }).
    ToSlice()
```

### The Escape Hatch: `OrmQueryFunc`

If you need dynamic query logic that cannot be expressed in the method name, you can add a variadic parameter of type `...OrmQueryFunc`
(which is an alias for `func(*gorm.DB) *gorm.DB`) as the final argument.

See **Chapter 7: Advanced Querying** for details on using this for preloads and custom SQL.

## 3. Bringing it Together

```go
package repositories

import (
    "time"
    . "github.com/BeamFoundry/rays-orm/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/core"
    . "github.com/BeamFoundry/rays/pkg/lang"
    . "github.com/BeamFoundry/rays/pkg/util/stream"
)

type Profile struct {
    ID        uint
    UserID    uint
    Bio       string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Comment struct {
    ID        uint
    PostID    uint
    Content   string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Post struct {
    ID        uint
    UserID    uint
    Title     string
    Status    string
    Comments  []Comment
    CreatedAt time.Time
    UpdatedAt time.Time
}

type User struct {
    ID        uint
    Name      string
    Status    string
    Role      string
    Age       uint
    Profile   Profile
    Posts     []Post
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 1. Define the public contract
type UserRepositoryInterface interface {
    OrmRepositoryInterface[User, uint]

    FindByName(name string) *User
    ListActiveUsers(status string, minAge uint) ([]User, Error)
    ListByStatus(status string, customFilters ...OrmQueryFunc) ([]User, Error)
    FindByRole(role string, customFilters ...OrmQueryFunc) []User
    StreamUsersByAgeGreaterThan(minAge uint) (Stream[User], Error)
    FindByAgeRangeDynamic(min, max uint, customFilters ...OrmQueryFunc) []User
}

// 2. Define the implementation
type UserRepository struct {
    OrmRepository[User, uint]

    findByName                              func(string) *User
    listByStatusAndAgeGreaterThan           func(string, uint) ([]User, Error)
    listByStatus                            func(string, ...OrmQueryFunc) ([]User, Error)
    findByRole                              func(string, ...OrmQueryFunc) []User
    streamByAgeGreaterThan                  func(uint) (Stream[User], Error)
    findByAgeBetweenOrderByNameAscIDDesc    func(uint, uint, ...OrmQueryFunc) []User
}

// 3. Expose the generated functions
func (this *UserRepository) FindByName(name string) *User {
    return this.findByName(name)
}

func (this *UserRepository) ListActiveUsers(status string, minAge uint) ([]User, Error) {
    return this.listByStatusAndAgeGreaterThan(status, minAge)
}

func (this *UserRepository) ListByStatus(status string, customFilters ...OrmQueryFunc) ([]User, Error) {
    return this.listByStatus(status, customFilters...)
}

func (this *UserRepository) FindByRole(role string, customFilters ...OrmQueryFunc) []User {
    return this.findByRole(role, customFilters...)
}

func (this *UserRepository) StreamUsersByAgeGreaterThan(minAge uint) (Stream[User], Error) {
    return this.streamByAgeGreaterThan(minAge)
}

func (this *UserRepository) FindByAgeRangeDynamic(min, max uint, customFilters ...OrmQueryFunc) []User {
    return this.findByAgeBetweenOrderByNameAscIDDesc(min, max, customFilters...)
}
```
