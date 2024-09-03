package sync

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(n int) *Semaphore {
	return &Semaphore{ch: make(chan struct{}, n)}
}

func (s *Semaphore) Acquire(n int) {
	if cap(s.ch) == 0 {
		return
	}

	e := struct{}{}
	for i := 0; i < n; i++ {
		s.ch <- e
	}
}

func (s *Semaphore) Release(n int) {
	if cap(s.ch) == 0 {
		return
	}

	for i := 0; i < n; i++ {
		<-s.ch
	}
}
