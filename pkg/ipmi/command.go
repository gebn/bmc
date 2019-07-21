package ipmi

import (
	"fmt"
)

// Command represents a particular operation that can be requested. Command
// identifiers are only unique within a given network function. See Appendix G
// in the v1.5 and v2.0 specs for assignments. This is a 1 byte uint on the
// wire.
type Command uint8

func (c Command) String() string {
	// cannot do much better than this without the context of the NetFn; we use
	// hex as this is what can be found in the spec
	return fmt.Sprintf("%#x", uint8(c))
}
