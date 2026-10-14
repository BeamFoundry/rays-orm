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

1.  **Prefix:** Must be `find`, `Find`, `query`, or `Query`.
2.  **Subject:** (Optional) Capitalized words describing what is being found (e.g., `ActiveUsers`).
3.  **Separator:** Must include `By`.
4.  **Conditions:** Field names followed by an optional Operator, separated by `And` or `Or`.
5.  **OrderBy:** (Optional) The keyword `OrderBy` followed by one or more sorting rules. Each rule consists of a field name
and an optional direction (`Asc` or `Desc`, defaulting to `Desc`). Multiple sorting rules must be separated by the `And` keyword.

*Example:* `findUsersByAgeGreaterThanAndNameLikeOrderByNameAscAndCreatedAtDesc`

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
* **Multiple Entities:** `[]*T`
* **Query Builder:** `*gorm.DB` (Returns the query before mapping to an object)
* **Error Handling:** You may append `Error` as a second return parameter to any of the above to catch database faults.

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
)

type User struct {
    ID        uint
    Name      string
    Age       uint
    CreatedAt time.Time
    UpdatedAt time.Time
}

// 1. Define the public contract
type UserRepositoryInterface interface {
    OrmRepositoryInterface[User, uint]

    FindByName(name string) *User
    FindActiveUsers(status string, minAge int) ([]*User, Error)
    FindByAgeRangeDynamic(min, max int, customFilters ...OrmQueryFunc) []*User
}

// 2. Define the implementation
type UserRepository struct {
    OrmRepository[User, uint]

    findByName                        func(string) *User
    findByStatusAndAgeGreaterThan     func(string, int) ([]*User, Error)
    findByAgeBetweenOrderByNameAscAndIdDesc func(int, int, ...OrmQueryFunc) []*User
}

// 3. Expose the generated functions
func (this *UserRepository) FindByName(name string) *User {
    return this.findByName(name)
}

func (this *UserRepository) FindActiveUsers(status string, minAge int) ([]*User, Error) {
    return this.findByStatusAndAgeGreaterThan(status, minAge)
}

func (this *UserRepository) FindByAgeRangeDynamic(min, max int, customFilters ...OrmQueryFunc) []*User {
    return this.findByAgeBetweenOrderByNameAscAndIdDesc(min, max, customFilters...)
}
```
