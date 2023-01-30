package rlpx

import (
	"fmt"

	"github.com/indexsupply/x/rlp"
)

type Capability struct {
	Name          string                                // Name of the protocol, eg "eth"
	Version       string                                // Version of the protocol, eg "67"
	HighestMsgID  uint                                  // The highest message ID for the protocol. Since RLPX messages are multiplexed in a compact manner, this determines the "message space" for this capability
	HandleMessage func(msgID uint, item rlp.Item) error // A handler for an incoming message ID and RLP item
}

func (c Capability) CanonicalName() string {
	return fmt.Sprintf("%s/%s", c.Name, c.Version)
}

type Capabilities []Capability

func (c Capabilities) Len() int {
	return len(c)
}

func (c Capabilities) Less(i, j int) bool {
	return c[i].CanonicalName() < c[j].CanonicalName()
}

func (c Capabilities) Swap(i, j int) {
	t := c[i]
	c[j] = c[i]
	c[j] = t
}
