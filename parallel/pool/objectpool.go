package pool

import "fmt"

type ObjPool[T any] struct {
	obj_ch chan *T
	new_fn func() *T
}

// NewObjPool creates a new generic object pool with a fixed size.
//
// # Parameters
//
//   - poolSize: maximum number of objects in the pool. Must be > 0.
//   - newFn: factory function to create new objects when pool is empty.
//
// # Behavior
//
//   - Initializes the pool by pre-creating poolSize objects.
//   - Get() retrieves an object from pool, or creates a new one if empty.
//   - Put() returns an object to pool. If pool is full, the object is dropped.
//   - Unlike sync.Pool, objects in this pool are NOT cleared by GC.
//   - One typical usecase is to avoid allocate & GC clear frequently and reuse the objects in the pool
//
// # Example
//
//	pool, err := NewObjPool(10, func() *Connection {
//	    return OpenConnection()
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get object from pool
//	conn := pool.Get()
//	defer pool.Put(conn)
//
//	// Use conn...
//
// # Thread Safety
//
// This function is thread-safe. All operations (Get/Put) can be called
// concurrently from multiple goroutines.
func NewObjPool[T any](pool_size int, new_fn func() *T) (*ObjPool[T], error) {
	if pool_size <= 0 {
		return nil, fmt.Errorf("pool size <= 0")
	}

	ret := &ObjPool[T]{
		obj_ch: make(chan *T, pool_size),
		new_fn: new_fn,
	}

	for range pool_size {
		ret.obj_ch <- new_fn()
	}

	return ret, nil
}

func (op *ObjPool[T]) Get() *T {
	select {
	case obj := <-op.obj_ch:
		return obj
	default:
		return op.new_fn()
	}
}

func (op *ObjPool[T]) Put(obj *T) {
	select {
	case op.obj_ch <- obj:
		return

	default:
		return
	}
}
