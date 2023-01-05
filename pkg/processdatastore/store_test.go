package processdatastore_test

import (
	"testing"

	"github.com/ci4rail/velog/pkg/processdatastore"

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
	o1, updates, err := s.Read(456)
	assert.NoError(t, err)
	assert.Equal(t, 1, updates)
	assert.Equal(t, o, o1)
	assert.Equal(t, uint32(456), o1.Address())
	assert.Equal(t, []byte{1, 2, 3}, o1.Data())

	// now try to read the entry again, numUpdates should be 0
	o1, updates, err = s.Read(456)
	assert.NoError(t, err)
	assert.Equal(t, 0, updates)
	assert.Equal(t, o, o1)

	// multiple writes
	s.Write(newMyObject(124, 456, []byte{2, 3, 4}))
	s.Write(newMyObject(125, 456, []byte{2, 3, 5}))
	s.Write(newMyObject(126, 456, []byte{2, 3, 6}))
	o1, updates, err = s.Read(456)
	assert.NoError(t, err)
	assert.Equal(t, 3, updates)
	assert.Equal(t, []byte{2, 3, 6}, o1.Data())

	// never written address
	o1, updates, err = s.Read(111)
	assert.Error(t, err)
	assert.Nil(t, o1)
	assert.Equal(t, 0, updates)

	// list
	s.Write(newMyObject(127, 458, []byte{2, 3, 7}))
	s.Write(newMyObject(128, 457, []byte{2, 3, 8}))

	list := s.List()
	assert.Equal(t, 3, len(list))
	assert.Equal(t, 456, list[0])
	assert.Equal(t, 457, list[1])
	assert.Equal(t, 458, list[2])
}
