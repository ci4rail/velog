package processdatastore_test

import (
	"testing"

	"github.com/ci4rail/mvb-can-logger/pkg/processdatastore"

	"github.com/stretchr/testify/assert"
)

type MyObject struct {
	timestamp int64
	address   uint32
	data      []byte
}

func newMyObject(timestamp int64, address uint32, data []byte) *MyObject {
	return &MyObject{
		timestamp: timestamp,
		address:   address,
		data:      data,
	}
}

func (o *MyObject) Timestamp() int64 {
	return o.timestamp
}

func (o *MyObject) Address() uint32 {
	return o.address
}

func (o *MyObject) Data() []byte {
	return o.data
}

func (o *MyObject) AdditionalInfo() []string {
	return nil
}

func TestStore(t *testing.T) {
	s := processdatastore.NewStore()
	o := newMyObject(123, 456, []byte{1, 2, 3})
	s.Write(o)
	e, err := s.ReadAndClearEntry(456)
	assert.NoError(t, err)
	assert.Equal(t, 1, e.NumUpdates)
	assert.Equal(t, o, e.RecentObject)
	assert.Equal(t, uint32(456), e.RecentObject.Address())
	assert.Equal(t, []byte{1, 2, 3}, e.RecentObject.Data())

	// now try to read the entry again, which should fail
	e, err = s.ReadAndClearEntry(456)
	assert.Error(t, err)
	assert.Nil(t, e)

	// multiple writes
	s.Write(newMyObject(124, 456, []byte{2, 3, 4}))
	s.Write(newMyObject(125, 456, []byte{2, 3, 5}))
	s.Write(newMyObject(126, 456, []byte{2, 3, 6}))
	e, err = s.ReadAndClearEntry(456)
	assert.NoError(t, err)
	assert.Equal(t, 3, e.NumUpdates)
	assert.Equal(t, []byte{2, 3, 6}, e.RecentObject.Data())

	// never written address
	e, err = s.ReadAndClearEntry(111)
	assert.Error(t, err)
	assert.Nil(t, e)
}
