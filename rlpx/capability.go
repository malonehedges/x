package rlpx

import (
	"fmt"

	"github.com/indexsupply/x/rlp"
)

type Capability struct {
	Name          string
	Version       string
	TotalMessages uint
	HandleMessage func(msgID uint, item rlp.Item) error
}

func (c Capability) CanonicalName() string {
	fmt.Sprintf("%s/%s", c.Name, c.Version)
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

// ForMsgID takes in the multiplexed msgID and returns the capability and capability msgID
// that matches.
// From the docs:
// Message IDs are assumed to be compact from ID 0x10 onwards (0x00-0x0f is reserved for the "p2p" capability) and given to each shared (equal-version, equal-name) capability in alphabetic order.
func (c Capabilities) ForMsgID(msgID uint) (*Capability, uint) {
	var idx uint = 0
	for _, cap := range c {
		if (idx + cap.TotalMessages) >= msgID {
			return &cap, msgID - idx
		}
		idx += cap.TotalMessages
	}
	return nil, msgID
}
