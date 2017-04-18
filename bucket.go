package lazybucket

import (
	"errors"
	"reflect"
	"sync"
	"time"
)

// This is implementation of kinda leaky bucket.
// Every function that needs to be controlled must
// be first registered through RegisterFunction, once
// the function is registered a go routine is dispatched.
// This go routine is called feeder and its porpose is to
// reset the counter for every registered function when
// its time window is exhausted. We do not actually call
// any external function here, we only control if they
// can or can not be called using the IsAbleToCall
// function.
//
// The steps to use this package is: 1) get a new bucket
// through GetBucket. 2) register a function within the
// bucket suplying the max number of calls and also the
// time window in seconds. 3) before every call to a
// registered function you must check if you are able or
// not by calling IsAbleToCall(). 4) at the end, unregister
// the function to free resources.

type Bucket struct {
	Functions map[uintptr]uint64
	Locker    sync.Mutex
}

// GetBucket initiates a new bucket. Buckets allows
// throttling control of an arbitrary number of
// functions.
func GetBucket() Bucket {
	return Bucket{
		Functions: make(map[uintptr]uint64),
		Locker:    sync.Mutex{},
	}
}

// RegisterFunction registers a function into the
// bucket and dispatches the feeder. fn is the function,
// rate the the numbers of calls allowed in twindow time
// in seconds of the time window
func (bucket *Bucket) RegisterFunction(fn interface{}, rate, twindow uint64) error {

	var fnaddr uintptr
	var exists bool

	if fn == nil {
		return errors.New("invalid function")
	}

	if rate == 0 {
		return errors.New("invalid rate")
	}

	if twindow == 0 {
		return errors.New("invalid time window")
	}

	fnaddr = reflect.ValueOf(fn).Pointer()

	bucket.Locker.Lock()
	_, exists = bucket.Functions[fnaddr]
	if exists {
		bucket.Locker.Unlock()
		return errors.New("function already registered")
	}

	bucket.Functions[fnaddr] = rate
	bucket.Locker.Unlock()
	go bucket.doFeed(fnaddr, rate, twindow)
	return nil
}

// doFeed is a go routine that resets the rate
// at every twindow seconds. exits only when
// the function has been unregistered. Function
// for internal use only
func (bucket *Bucket) doFeed(fnaddr uintptr, rate, twindow uint64) {
	var exists bool

	for {
		time.Sleep(time.Duration(twindow) * time.Second)

		bucket.Locker.Lock()
		_, exists = bucket.Functions[fnaddr]
		if exists == false {
			bucket.Locker.Unlock()
			return
		}

		bucket.Functions[fnaddr] = rate
		bucket.Locker.Unlock()
	}
}

func (bucket *Bucket) UnregisterFunction(fn interface{}) {
	var fnaddr uintptr
	var exists bool

	fnaddr = reflect.ValueOf(fn).Pointer()

	bucket.Locker.Lock()
	_, exists = bucket.Functions[fnaddr]
	if exists == false {
		bucket.Locker.Unlock()
		return
	}

	delete(bucket.Functions, fnaddr)
	bucket.Locker.Unlock()
}

// IsAbleToCall returns if fn function may or not be
// called. This function decreases the function counter
// for the fn function. An error may be returned if fn
// is not registered yet.
func (bucket *Bucket) IsAbleToCall(fn interface{}) (bool, error) {

	var fnaddr uintptr
	var exists bool

	fnaddr = reflect.ValueOf(fn).Pointer()

	bucket.Locker.Lock()
	_, exists = bucket.Functions[fnaddr]
	if exists == false {
		bucket.Locker.Unlock()
		return false, errors.New("function not registered")
	}

	if bucket.Functions[fnaddr] == 0 {
		bucket.Locker.Unlock()
		return false, nil
	}

	bucket.Functions[fnaddr]--
	bucket.Locker.Unlock()
	return true, nil
}
