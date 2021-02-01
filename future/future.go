package future

// NewFuture is
func NewFuture(body func() interface{}) <-chan interface{} {
	r := make(chan interface{}, 1)
	go func() {
		r <- body()
	}()
	return r
}

// JoinFutures is
func JoinFutures(chans ...<-chan interface{}) <-chan []interface{} {
	r := make(chan []interface{})
	ra := make([]interface{}, len(chans))
	go func() {
		for i, c := range chans {
			ra[i] = <-c
		}
		r <- ra
	}()
	return r
}

// FirstFuture is
func FirstFuture(a <-chan interface{}, b <-chan interface{}) <-chan interface{} {
	r := make(chan interface{})
	go func() {
		select {
		case a1 := <-a:
			r <- a1
		case b1 := <-b:
			r <- b1
		}
	}()
	return r
}

// FirstFutures is
func FirstFutures(chans ...<-chan interface{}) <-chan interface{} {
	r := chans[0]
	for _, c := range chans {
		r = FirstFuture(r, c)
	}
	return r
}
