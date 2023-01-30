package capabilities

import (
	"fmt"

	"github.com/indexsupply/x/rlp"
	"github.com/indexsupply/x/rlpx"
)

var Eth67 = rlpx.Capability{
	Name:          "eth",
	Version:       "67",
	HandleMessage: handleEthMessage,
}

func handleEthMessage(msgID uint, item rlp.Item) error {
	switch msgID {
	case 0x01:
		fmt.Printf("<status version=%d network=%d difficulty=%d\n", item.At(0).Uint16(), item.At(1).Uint16(), item.At(2).Uint64())
	}
	return nil
}
