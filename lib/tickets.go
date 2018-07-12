package lib

import "fmt"

type tickets struct {
	active bool
	total  uint32
	pool   chan struct{}
}

func NewTickets(total uint32) Tickets {
	if total < 0 {
		panic("tickets must more than 0")
	}

	return &tickets{
		active: true,
		total:  total,
	}
}

func (t *tickets) Init() {
	t.pool = make(chan struct{}, t.total)
	var i uint32
	for i = 0; i < t.total; i++ {
		t.pool <- struct{}{}
	}
}

func (t *tickets) Put() {
	fmt.Println("put back !!!")
	t.pool <- struct{}{}
}

func (t *tickets) Get() bool {
	_, ok := <-t.pool
	return ok
}

/*func (t *tickets) Active() bool {
	return t.active
}
*/
func (t *tickets) Total() uint32 {
	return t.total
}

func (t *tickets) Remainder() uint32 {
	return uint32(len(t.pool))
}

func (t *tickets) Close() {
	if t.active == false {
		return
	}
	close(t.pool)
	t.active = false
}
