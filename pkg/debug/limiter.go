package debug

type Limiter struct {
}

func NewLimiter() *Limiter {
	return &Limiter{}
}

func (l *Limiter) Limit() bool {
	return false
}
