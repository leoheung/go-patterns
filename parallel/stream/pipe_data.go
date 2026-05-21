package stream

import (
	"sync"

	"github.com/leoheung/go-patterns/cryptography"
)

type Data[I, O any] struct {
	id        string
	input     I
	output    O
	ready     chan struct{}
	readyOnce sync.Once
}

type DataCh[I, O any] struct {
	ch chan *Data[I, O]
}

func (d *Data[I, O]) GetInput() I {
	return d.input
}

func (d *Data[I, O]) GetOutput() O {
	return d.output
}

func (d *Data[I, O]) GetID() string {
	return d.id
}

func (d *Data[I, O]) SetOutput(output O) {
	d.output = output
	d.readyOnce.Do(func() { close(d.ready) })
}

func (d *Data[I, O]) Wait() {
	<-d.ready
}

func NewDataCh[I, O any]() *DataCh[I, O] {
	return &DataCh[I, O]{
		ch: make(chan *Data[I, O]),
	}
}

func (dc *DataCh[I, O]) Enq(input I) *Data[I, O] {
	id := cryptography.RandUUID()
	ret := Data[I, O]{
		id:        id,
		input:     input,
		ready:     make(chan struct{}),
		readyOnce: sync.Once{},
	}

	dc.ch <- &ret
	return &ret
}

func (dc *DataCh[I, O]) GetCh() chan any {
	ret := make(chan any)

	go func() {
		defer close(ret)

		for {
			x, ok := <-dc.ch
			if !ok {
				return
			}

			ret <- x
		}
	}()

	return ret
}

func (dc *DataCh[I, O]) Close() {
	close(dc.ch)
}
