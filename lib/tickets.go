/*
 * Revision History:
 *     Initial: 2018/7/04        ShiChao
 */

package lib

type tickets struct {
	total uint32
	pool  chan struct{}
}

func NewTickets(total uint32) Tickets {
	if total < 0 {
		panic("tickets must more than 0")
	}

	t := &tickets{
		total: total,
	}
	t.init()

	return t
}

func (t *tickets) init() {
	t.pool = make(chan struct{}, t.total)
	var i uint32
	for i = 0; i < t.total; i++ {
		t.pool <- struct{}{}
	}
}

func (t *tickets) Put() {
	t.pool <- struct{}{}
}

func (t *tickets) Get() bool {
	_, ok := <-t.pool
	return ok
}

func (t *tickets) Total() uint32 {
	return t.total
}

func (t *tickets) Remainder() uint32 {
	return uint32(len(t.pool))
}
