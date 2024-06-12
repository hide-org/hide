package result

type Result[T any] struct {
	value T
	Error error
}

type None struct{}

type Empty = Result[None]

func Success[T any](value T) Result[T] {
	return Result[T]{value: value, Error: nil}
}

func Failure[T any](err error) Result[T] {
	var zeroValue T
	return Result[T]{value: zeroValue, Error: err}
}

func EmptySuccess() Empty {
	return Empty{value: None{}, Error: nil}
}

func EmptyFailure(err error) Empty {
	return Empty{value: None{}, Error: err}
}

func (r Result[T]) IsSuccess() bool {
	return r.Error == nil
}

func (r Result[T]) IsFailure() bool {
	return r.Error != nil
}

func (r Result[T]) Get() T {
	if r.IsSuccess() {
		return r.value
	}
	panic("Result is a failure")
}

func (r Result[T]) OrElse(f func() Result[T]) Result[T] {
	if r.IsSuccess() {
		return r
	}
	return f()
}
