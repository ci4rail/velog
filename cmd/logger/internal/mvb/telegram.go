package mvb

import (
	"strconv"

	mvbpb "github.com/ci4rail/io4edge_api/mvbSniffer/go/mvbSniffer/v1"
)

// TelegramObject is a wrapper for the mvbpb.Telegram
type TelegramObject struct {
	telegram *mvbpb.Telegram
}

func newTelegramObject(telegram *mvbpb.Telegram) *TelegramObject {
	return &TelegramObject{
		telegram: telegram,
	}
}

// Timestamp returns the timestamp of the telegram
func (t *TelegramObject) Timestamp() int64 {
	return int64(t.telegram.Timestamp)
}

// Address returns the address of the telegram
func (t *TelegramObject) Address() uint32 {
	return uint32(t.telegram.Address)
}

// Data returns the data of the telegram
func (t *TelegramObject) Data() []byte {
	return t.telegram.Data
}

// AdditionalInfo returns additional information about the telegram
func (t *TelegramObject) AdditionalInfo() []string {
	strArr := []string{
		strconv.Itoa(int(t.telegram.Type)),
	}
	return strArr
}
