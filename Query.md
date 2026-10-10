In Golang, I have written a function to parse a function name, formatted just like in Spring-Data JPA,
to generate queries for Gorm. So it is mostly ORM related (ie, a struct mapping query stuff).

For example a function like "FindAllOrdersByUsernameOrderByOrderId(username string) []Order" will be translated to a query like
"SELECT * FROM user WHERE username IS ?"

The function parsed is actually a field in a struct, and so it is passed as a `reflect.StructField` to the function builder.

The function constructs the following object, which is then used to create the query.

```go
// <T> : the struct mapped by Gorm, it must have a field named "Id" as primary of type <ID>
type GormQueryBuilder[T OrmEntity[ID], ID comparable] struct {
	function   reflect.StructField
	conditions []whereQuery
	orders     []orderByQuery
}

// one where clause
// 
// operator will be translated from the text parsed and put in uppercase into SQL.
//
//  - "IS", "EQUALS" : translates to SQL operator "=" // The default operator if none is specified
//  - "NOT : translates to SQL operator "<>"
//  - "GREATERTHAN" or "AFTER" : translates to ">"
//  - "LESSTHAN", "BEFORE" : translates to "<"
//  - "GREATERTHANOREQUAL" : translates to ">="
//  - "LESSTHANOREQUAL" : translates to "<="
//  - "ISNULL", "NULL" : translates to "IS NULL"
//  - "ISNOTNULL", "NOTNULL" : translates to "NOT NULL"
//  - "LIKE", "IN", "BETWEEN": translates to themself
//  - "NOTLIKE" : "NOT LIKE"
//  - "NOTIN": "NOT IN"
//
// args is always 1, except for "ISNULL", "NULL", "ISNOTNULL", "NOTNULL" where it's 0 and "BETWEEN" where its 2.
type whereQuery struct {
	field    string // One of T member
	mode     string // AND or OR -- Defaults to AND
	operator string // Operator to use, See above.
	args     int // number of arguments this operator needs. See above.
}

// one of the OrderBy clause.
type orderByQuery struct {
	field string // One of T member
	order string // ASC or DESC
}

```

To help me I have developed a few types, whose implementation is not required for this to understand.

```go
// NewError return a new "Error".
//
// "code" should be the error type like "IllegalArgumentError"
// "message" is the error message
func NewError(code, message string) Error

// Optional is an implementation of an Optional type.
type Optional[T] struct{}

// OrmQueryFunc represent a query filter operation.
type OrmQueryFunc = func(*gorm.DB) *gorm.DB
```

Right now the function name is parsed using complex regexps. It works fine but the regexps complexity makes me fear it will fail one day.

I want to know if a better solution is possible. Right now the GormQueryBuilder is filled and later verified for the following.

- each "conditions"."field" and "orders"."field" are checked for existance in `T`
- foreach "conditions", "args" inputs of the "condition"."field" type in `T` type must be present in "function" except the following :
  - If the operator is "IN" or "NOTIN" it should be a slice the "condition"."field" type in `T`.
  - If the operator is "LIKE" on "NOTLIKE", the type in `T`  and the argument in the function should be a `string`.
- the function may include a variadic argument of `OrmQueryFunc` type.
- "function" returned type must be of `*T` or `[]T` or of `Optional[*T]`.

Assuming the following :

- a "word" is start by an uppercase letter and is followed only by lowercase letter
- a struct field mapped by ORM consist of at least one "word".

The function format is the following :

-It should always starts with "Find" or "Query" (or for private function "find" or query"), followed by zero or more "word", and by the word "By".
- It is followed by a list of where clauses, consisting of :
  - a `T` field (which may consist of multiple "words"), respecting the same letter case.
  - optionally an operator, if not present "Is" or "Equals" is assumed. valid operators are
    "Is", "Equals", "Not", "GreaterThan", "After", "LessThan", "Before", "GreaterThanOrEqual", "LessThanOrEqual",
    "IsNull", "Null", "IsNotNull", "NotNull", "In", "NotIn", Between", "Like ", "NotLike".
- Optionally it should be followed by "OrderBy", and a list of order clauses, consisting of :
  - a `T` field (which may consist of multiple "words"), respecting the same letter case.
  - optionally a sort order, which is either "Asc" (default) or "Desc"

One thing to be careful during parsing is to correctly parse the operators.
There is an "Is" and "Not" operator and many other operators starts with that. That the same for "GreaterThan" and "LessThan".

To modernize this, you could use the following Enum[E] type I have developed.

```go
// Where Mode
type whereModeEnum struct {
  AND,
  OR whereMode
}

type whereMode = Enum[whereModeEnum]

var whereModes whereModeEnum = WhereMode{}.New()

// where Operator
type operatorEnum struct {
  IS, EQUALS, NOT,
  GREATERTHAN, AFTER,
  LESSTHAN, BEFORE,
  GREATERTHANOREQUAL, LESSTHANOREQUAL,
  ISNULL, NULL,
  ISNOTNULL, NOTNULL,
  BETWEEN,
  LIKE, IN,
  NOTLIKE, NOTIN operator
}

type operator = Enum[operatorEnum]

var operators operatorEnum = operator{}.New() // Initialize the struct instance

// order sort
type orderSortingEnum struct {
  ASC,
  DESC orderSorting
}

type orderSorting = Enum[orderSortingEnum]

var orderSortings orderSortingEnum = orderSorting{}.New()

```

So for the types `whereMode`, `operator` and `orderSorting`, Enum[T] gives the following functions :

```go
// Compare compares this Enum to another Enum of the same type.
// It returns a negative integer, zero, or a positive integer as this Enum's
// ordinal is less than, equal to, or greater than the compared Enum's ordinal.
func (this Enum[T]) Compare(that Enum[T]) int

// Ordinal returns the zero-based index of the enum value within the container T,
// excluding any fields that are not of type Enum[T].
func (this Enum[T]) Ordinal() int

// Name returns the raw field name of this enum value as defined in the container struct T.
func (this Enum[T]) Name() string

// From parses the given string and sets the enum to the matching value,
// returning an error if no match is found.
func (this *Enum[T]) From(text string) error
```

So you could upgrade `whereQuery` and `orderByQuery` to something like :

```go
type whereQuery struct {
	field    reflect.StructField // One of T member
	mode     whereMode
	operator operator // Operator to use, See above.
  // args could be removed as the rules are simple for counting them
}

type orderByQuery struct {
	field reflect.StructField // One of T member
	order orderSorting // ASC or DESC
}
```

What I am asking is to recreate a new function with this signature :

```go
NewGormQueryBuilder(function reflect.StructField) (*GormQueryBuilder[T, ID], Error)

For "Error" code, you can use the "IllegalArgumentError".
```
