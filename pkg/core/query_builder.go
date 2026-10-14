package core

import (
	"iter"
	"reflect"
	"slices"
	"strings"

	. "github.com/BeamFoundry/rays/pkg/core"
	. "github.com/BeamFoundry/rays/pkg/lang"
	. "github.com/BeamFoundry/rays/pkg/util"
	"gorm.io/gorm"
)

type gormQueryBuilder[T any, ID comparable] struct {
	logger    Logging
	fields    iter.Seq2[string, *reflect.StructField]
	operators []string
}

func (this gormQueryBuilder[T, ID]) New(logger Logging) *gormQueryBuilder[T, ID] {
	this.logger = logger
	entity := reflect.TypeFor[T]()

	fields := slices.SortedFunc(entity.Fields(), func(a, b reflect.StructField) int {
		return strings.Compare(b.Name, a.Name)
	})

	this.fields = func(yield func(string, *reflect.StructField) bool) {
		for _, field := range fields {
			if !yield(field.Name, &field) {
				return
			}
		}
	}

	sortFunc := func(a, b string) int { return strings.Compare(b, a) }
	ops := []string{
		"GreaterThanOrEqual", "LessThanOrEqual",
		"IsNotNull", "NotNull",
		"GreaterThan", "LessThan",
		"NotLike", "Between",
		"Equals", "IsNull", "NotIn",
		"After", "Before", "Null", "Like",
		"Not", "In", "Is",
	}
	this.operators = slices.SortedFunc(slices.Values(ops), sortFunc)
	return &this
}

func (this *gormQueryBuilder[T, ID]) parseOrderByQuery(s string) (*orderByQuery, string, Error) {
	o := &orderByQuery{
		order: orderSortings.ASC,
	}

	for name, field := range this.fields {
		if !String(s).StartsWith(name) {
			continue
		}
		rest := s[len(name):]
		o.field = field
		switch {
		case rest == "":
		case String(rest).StartsWith("Asc"):
			rest = rest[3:]
		case String(rest).StartsWith("Desc"):
			rest = rest[4:]
			o.order = orderSortings.DESC
		default:
		}
		return o, rest, nil
	}
	return nil, "", NewError("IllegalArgumentError", "Could not match a valid T field for condition")
}

func (this *gormQueryBuilder[T, ID]) parseWhereQuery(s string) (*whereQuery, string, Error) {
	w := &whereQuery{
		mode:     whereModes.AND,
		operator: operators.IS,
	}
	switch {
	case String(s).StartsWith("By"):
		s = s[2:]
	case String(s).StartsWith("And"):
		s = s[3:]
	case String(s).StartsWith("Or"):
		s = s[2:]
		w.mode = whereModes.OR
	default:
		return nil, "", NewError("IllegalArgumentError", "Expected 'And' or 'Or' between conditions")
	}

	for name, field := range this.fields {
		if !String(s).StartsWith(name) {
			continue
		}
		rest := s[len(name):]
		w.field = field
		if rest == "" || String(rest).StartsWith("And") || String(rest).StartsWith("Or") {
			return w, rest, nil
		}
		for _, opStr := range this.operators {
			if String(rest).StartsWith(opStr) {
				rest = rest[len(opStr):]
				var op operator
				if err := (&op).From(String(opStr).ToUpperCase().String()); err == nil {
					w.operator = op
				}
				return w, rest, nil
			}
		}
		return nil, "", NewError("IllegalArgumentError", "Invalid operator")
	}
	return nil, "", NewError("IllegalArgumentError", "Could not match a valid T field for condition")
}

func (this *gormQueryBuilder[T, ID]) Build(function *reflect.StructField) (*gormQuery[T, ID], Error) {
	name := CaseUtils(function.Name).Capitalize().String()

	// 1. Verify Prefix
	if !String(name).StartsWith("Find") && !String(name).StartsWith("Query") && !String(name).StartsWith("List") {
		return nil, NewError("IllegalArgumentError", "Function name must start with Find, Query, find, or query")
	}

	// 2. Find the "By" separator
	byIndex := Strings(name).Index("By")
	if byIndex == -1 {
		return nil, NewError("IllegalArgumentError", "Function name must contain 'By'")
	}

	// 3. Extract the Condition and OrderBy segments
	predicateStr := name[byIndex:]
	conditionStr := predicateStr
	orderStr := ""

	orderByIndex := Strings(predicateStr).LastIndex("OrderBy")
	if orderByIndex != -1 {
		conditionStr = predicateStr[:orderByIndex]
		orderStr = predicateStr[orderByIndex+7:] // Skip "OrderBy"
	}

	query := &gormQuery[T, ID]{
		logger:   this.logger,
		function: function,
	}

	this.logger.Debug("Analyzing function", "name", name, "predicate", predicateStr, "condition", conditionStr, "order", orderStr)

	// 4. Parse Conditions
	for conditionStr != "" {
		condition, rest, err := this.parseWhereQuery(conditionStr)
		if err != nil {
			return nil, err
		}

		conditionStr = rest

		this.logger.Debug("Added Condition", "field", condition.field.Name, "operator", condition.Operator())

		query.conditions = append(query.conditions, condition)
	}

	// 5. Parse Orders
	for orderStr != "" {
		order, rest, err := this.parseOrderByQuery(orderStr)
		if err != nil {
			return nil, err
		}

		orderStr = rest

		this.logger.Debug("Added ordering", "field", order.field.Name, "order", order.order.String())

		query.orders = append(query.orders, order)
	}

	if err := query.validate(); err != nil {
		return nil, err
	}

	return query, nil
}

type gormQuery[T any, ID comparable] struct {
	logger     Logging
	function   *reflect.StructField
	conditions []*whereQuery
	orders     []*orderByQuery
}

func (this *gormQuery[T, ID]) validate() Error {
	entity := reflect.TypeFor[T]()

	// Build a list of types we expect from the function arguments based on parsed clauses
	var expectedTypes []reflect.Type
	for _, cond := range this.conditions {
		fieldType := cond.field.Type
		switch cond.operator.Name() {
		case "ISNULL", "NULL", "ISNOTNULL", "NOTNULL":
			// 0 arguments
		case "BETWEEN":
			expectedTypes = append(expectedTypes, fieldType, fieldType) // 2 arguments
		case "IN", "NOTIN":
			expectedTypes = append(expectedTypes, reflect.SliceOf(fieldType))
		case "LIKE", "NOTLIKE":
			expectedTypes = append(expectedTypes, reflect.TypeFor[string]())
		default:
			expectedTypes = append(expectedTypes, fieldType) // 1 argument
		}
	}

	numIn := this.function.Type.NumIn()

	// Check if trailing argument is variadic/optional OrmQueryFunc
	if numIn == len(expectedTypes)+1 &&
		this.function.Type.IsVariadic() &&
		this.function.Type.In(len(expectedTypes)).Elem() == reflect.TypeFor[OrmQueryFunc]() {
		numIn-- // Ignore the trailing OrmQueryFunc for parameter mapping counts
	}

	if len(expectedTypes) != numIn {
		return NewError("IllegalArgumentError", "Argument count mismatch between function definition and parsed queries")
	}

	for i, expected := range expectedTypes {
		if this.function.Type.In(i) != expected {
			return NewError("IllegalArgumentError", "Argument type mismatch for condition parameters")
		}
	}

	retType := this.function.Type.Out(0)
	ptrTType := reflect.PointerTo(entity)
	sliceTType := reflect.SliceOf(entity)
	optionalTType := reflect.TypeFor[Optional[*T]]()
	gormType := reflect.TypeFor[*gorm.DB]()

	if retType != ptrTType && retType != sliceTType && retType != optionalTType && retType != gormType {
		return NewError("IllegalArgumentError", "Invalid return type; must be *T, []T, Optional[*T] or *gorm.DB")
	}

	this.logger.Info("Created repository function", "entity", entity, "function", this.function.Name)
	return nil
}

func (this *gormQuery[T, ID]) Query(tx GormTransactional) func(in []reflect.Value) []reflect.Value {
	output := this.function.Type.Out(0)
	one := output == reflect.TypeFor[*T]()
	multi := output.Kind() == reflect.Slice
	optional := output == reflect.TypeFor[Optional[*T]]()

	var hasError bool
	if this.function.Type.NumOut() > 1 && this.function.Type.Out(1) == reflect.TypeFor[Error]() {
		hasError = true
	}

	return func(in []reflect.Value) []reflect.Value {
		var model T
		db := tx.Get().Model(&model)
		if one || optional {
			db = db.Limit(1)
		}
		queries := []OrmQueryFunc{}
		index := 0
		for _, condition := range this.conditions {
			args := []any{}
			for i := index; i < index+condition.NumArgs(); i++ {
				args = append(args, in[i].Interface())
			}
			this.logger.Trace("Adding condition", "type", condition, "args", args)
			queries = append(queries, condition.Query(args...))
			index += condition.NumArgs()
		}
		for _, order := range this.orders {
			queries = append(queries, order.Query())
		}

		if this.function.Type.NumIn() == index+1 && this.function.Type.IsVariadic() {
			if this.function.Type.In(index).Elem() == reflect.TypeFor[OrmQueryFunc]() {
				if q, ok := in[index].Interface().([]OrmQueryFunc); ok {
					this.logger.Trace("Function is Variadic", "function", this.function.Type)
					queries = append(queries, q...)
				}
			}
		}

		db = db.Scopes(queries...)

		if output == reflect.TypeFor[*gorm.DB]() {
			return []reflect.Value{reflect.ValueOf(db)}
		}

		ret := []reflect.Value{}
		var res *gorm.DB

		if multi {
			var t []T
			res = db.Find(&t)
			ret = append(ret, reflect.ValueOf(t))
		} else {
			var t T
			res = db.First(&t)

			switch {
			case optional:
				var opt Optional[*T]
				if res.Error == nil {
					//nolint:errcheck // Won't check
					(&opt).SetValue(&t)
				}
				ret = append(ret, reflect.ValueOf(opt))
			case res.Error == nil:
				ret = append(ret, reflect.ValueOf(&t))
			default:
				ret = append(ret, reflect.Zero(reflect.TypeFor[*T]()))
			}
		}

		if hasError {
			if res.Error != nil {
				ret = append(ret, reflect.ValueOf(NewError("GormError", res.Error.Error())))
			} else {
				ret = append(ret, reflect.Zero(reflect.TypeFor[Error]()))
			}
		}
		return ret
	}
}
