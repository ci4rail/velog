package processdatastore

// Object is the interface for all objects that can be stored in the process data store
type Object interface {
	// Timestamp returns the reception timestamp of the object
	Timestamp() int64

	// Address returns the address of the object
	Address() uint32

	// Data returns the data of the object
	Data() []byte

	// Additional Info returns additional info of the object
	AdditionalInfo() []string
}
