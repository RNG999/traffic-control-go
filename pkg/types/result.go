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
