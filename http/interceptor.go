package http

type Interceptor[T any] interface {
	Intercept(T) error
}

type ChainInterceptor[T any] []Interceptor[T]

func (c ChainInterceptor[T]) Intercept(v T) error {
	for _, i := range c {
		if err := i.Intercept(v); err != nil {
			return err
		}
	}

	return nil
}
