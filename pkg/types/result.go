package types

// Result represents the outcome of an operation that can fail.
type Result[T any] struct {
	value T
	err   error
}

// Success creates a successful Result.
func Success[T any](value T) Result[T] {
	return Result[T]{value: value, err: nil}
}

// Failure creates a failed Result.
func Failure[T any](err error) Result[T] {
	var zero T
	return Result[T]{value: zero, err: err}
}

// IsSuccess checks if the Result is successful.
func (r Result[T]) IsSuccess() bool {
	return r.err == nil
}

// IsFailure checks if the Result is a failure.
func (r Result[T]) IsFailure() bool {
	return r.err != nil
}

// Value returns the value if successful, panics if not.
func (r Result[T]) Value() T {
	if r.err != nil {
		panic("attempted to get value from failed Result")
	}

	return r.value
}

// Error returns the error if failed, nil if successful.
func (r Result[T]) Error() error {
	return r.err
}

// Map applies a function to the value if successful.
func (r Result[T]) Map(f func(T) T) Result[T] {
	if r.err != nil {
		return r
	}

	return Success(f(r.value))
}

// FlatMap applies a function that returns a Result.
func (r Result[T]) FlatMap(f func(T) Result[T]) Result[T] {
	if r.err != nil {
		return r
	}

	return f(r.value)
}

// Bind is an alias for FlatMap for monadic composition.
func (r Result[T]) Bind(f func(T) Result[T]) Result[T] {
	return r.FlatMap(f)
}

// Match executes one of two functions based on success/failure.
func (r Result[T]) Match(onSuccess func(T), onError func(error)) {
	if r.err != nil {
		onError(r.err)
	} else {
		onSuccess(r.value)
	}
}

// OrElse returns the value if successful, otherwise the provided default.
func (r Result[T]) OrElse(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}

	return r.value
}

// Option represents a value that may or may not be present.
type Option[T any] struct {
	value   *T
	present bool
}

// Some creates an Option with a value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: &value, present: true}
}

// None creates an empty Option.
func None[T any]() Option[T] {
	return Option[T]{value: nil, present: false}
}

// OptionFromPtr creates an Option from a pointer.
func OptionFromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// IsSome checks if the Option has a value.
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone checks if the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.present
}

// Value returns the value if present, panics if not.
func (o Option[T]) Value() T {
	if !o.present {
		panic("attempted to get value from None Option")
	}
	return *o.value
}

// GetOrElse returns the value if present, otherwise the provided default.
func (o Option[T]) GetOrElse(defaultValue T) T {
	if !o.present {
		return defaultValue
	}
	return *o.value
}

// Map applies a function to the value if present.
func (o Option[T]) Map(f func(T) T) Option[T] {
	if !o.present {
		return None[T]()
	}
	return Some(f(*o.value))
}

// FlatMap applies a function that returns an Option.
func (o Option[T]) FlatMap(f func(T) Option[T]) Option[T] {
	if !o.present {
		return None[T]()
	}
	return f(*o.value)
}

// Filter returns the Option if the predicate is true, otherwise None.
func (o Option[T]) Filter(predicate func(T) bool) Option[T] {
	if !o.present || !predicate(*o.value) {
		return None[T]()
	}
	return o
}

// Match executes one of two functions based on presence.
func (o Option[T]) Match(onSome func(T), onNone func()) {
	if o.present {
		onSome(*o.value)
	} else {
		onNone()
	}
}

// ToPtr returns a pointer to the value if present, otherwise nil.
func (o Option[T]) ToPtr() *T {
	if !o.present {
		return nil
	}
	// Create a copy to avoid sharing the internal pointer
	copy := *o.value
	return &copy
}

// Functional helpers for collections

// Map applies a function to each element in a slice.
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter returns a new slice containing only elements that satisfy the predicate.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce applies a function to reduce a slice to a single value.
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// Find returns the first element that satisfies the predicate.
func Find[T any](slice []T, predicate func(T) bool) Option[T] {
	for _, v := range slice {
		if predicate(v) {
			return Some(v)
		}
	}
	return None[T]()
}

// UpdateOrAppend updates the first element matching the predicate or appends if none found.
func UpdateOrAppend[T any](slice []T, item T, predicate func(T) bool) []T {
	for i, v := range slice {
		if predicate(v) {
			result := make([]T, len(slice))
			copy(result, slice)
			result[i] = item
			return result
		}
	}
	// If no match found, append
	result := make([]T, len(slice)+1)
	copy(result, slice)
	result[len(slice)] = item
	return result
}

// RemoveIf returns a new slice with elements matching the predicate removed.
func RemoveIf[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if !predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
