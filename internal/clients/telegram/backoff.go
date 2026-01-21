package telegram

import "time"

// ExponentialBackoff implements exponential backoff with jitter.
type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
}

// NewExponentialBackoff creates a new ExponentialBackoff with default settings.
func NewExponentialBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		InitialInterval: DefaultInitialBackoff,
		MaxInterval:     DefaultMaxBackoff,
		Multiplier:      DefaultBackoffFactor,
	}
}

// NextBackoff calculates the next backoff duration using exponential backoff.
func (b *ExponentialBackoff) NextBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	backoff := float64(b.InitialInterval)
	for i := 0; i < attempt; i++ {
		backoff *= b.Multiplier
		if backoff > float64(b.MaxInterval) {
			backoff = float64(b.MaxInterval)
			break
		}
	}

	return time.Duration(backoff)
}

// Reset is a no-op for stateless exponential backoff.
func (b *ExponentialBackoff) Reset() {}
