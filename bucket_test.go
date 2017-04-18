package lazybucket

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func registered() {}

func TestRegisterInvalidFunction(t *testing.T) {

	var err error
	var b Bucket

	b = GetBucket()
	err = b.RegisterFunction(nil, 10, 20)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid function")
}

func TestRegisterWithInvalidRate(t *testing.T) {

	var fnaddr uintptr
	var exists bool
	var err error
	var b Bucket

	b = GetBucket()
	err = b.RegisterFunction(registered, 0, 20)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid rate")

	fnaddr = reflect.ValueOf(registered).Pointer()
	_, exists = b.Functions[fnaddr]
	assert.False(t, exists)
}

func TestRegisterWithInvalidTimeWindow(t *testing.T) {

	var fnaddr uintptr
	var exists bool
	var err error
	var b Bucket

	b = GetBucket()
	err = b.RegisterFunction(registered, 10, 0)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "invalid time window")

	fnaddr = reflect.ValueOf(registered).Pointer()
	_, exists = b.Functions[fnaddr]
	assert.False(t, exists)
}

func TestRegisterFunction(t *testing.T) {
	var fnaddr uintptr
	var err error
	var b Bucket
	var fn func()

	fn = func() {}

	b = GetBucket()
	err = b.RegisterFunction(fn, 10, 1)
	assert.Nil(t, err)
	fnaddr = reflect.ValueOf(fn).Pointer()
	assert.Equal(t, b.Functions[fnaddr], uint64(10))
}

func TestUnRegisterFunction(t *testing.T) {
	var fnaddr uintptr
	var exists bool
	var err error
	var b Bucket

	b = GetBucket()
	err = b.RegisterFunction(registered, 10, 1)
	assert.Nil(t, err)
	fnaddr = reflect.ValueOf(registered).Pointer()
	assert.Equal(t, b.Functions[fnaddr], uint64(10))

	b.UnregisterFunction(registered)
	_, exists = b.Functions[fnaddr]
	assert.False(t, exists)
}

func TestRegisterTheSameFunctionTwice(t *testing.T) {
	var fnaddr uintptr
	var err error
	var b Bucket

	b = GetBucket()
	err = b.RegisterFunction(registered, 10, 1)
	assert.Nil(t, err)
	fnaddr = reflect.ValueOf(registered).Pointer()
	assert.Equal(t, b.Functions[fnaddr], uint64(10))

	err = b.RegisterFunction(registered, 11, 10)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "function already registered")
	assert.Equal(t, b.Functions[fnaddr], uint64(10))
}

func TestIsAbleToCallOnInvalidFunction(t *testing.T) {
	var able bool
	var err error
	var b Bucket

	b = GetBucket()

	able, err = b.IsAbleToCall(registered)
	assert.False(t, able)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "function not registered")
}

func TestIsAbleToCallFunction(t *testing.T) {
	var able bool
	var err error
	var b Bucket
	var i int

	b = GetBucket()

	// we can call `registered' 10 times in one second
	err = b.RegisterFunction(registered, 10, 1)
	assert.Nil(t, err)

	// let's see...
	for i = 0; i < 10; i++ {
		able, err = b.IsAbleToCall(registered)
		assert.True(t, able)
		assert.Nil(t, err)
	}

	// now we have exhausted the calls
	able, err = b.IsAbleToCall(registered)
	assert.False(t, able)
	assert.Nil(t, err)

	// sleeps
	time.Sleep(time.Second * 2)

	// now we are supposed to be allowed again
	// lets see...
	for i = 0; i < 10; i++ {
		able, err = b.IsAbleToCall(registered)
		assert.True(t, able)
		assert.Nil(t, err)
	}

	// exhausted number of calls again
	able, err = b.IsAbleToCall(registered)
	assert.False(t, able)
	assert.Nil(t, err)
}
