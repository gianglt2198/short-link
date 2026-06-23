package helpers

import (
	"github.com/bwmarrin/snowflake"
)

// CodeGenerator produces short codes for URL shortening.
type CodeGenerator interface {
	Generate() (string, error)
}

const (
	alphabet  = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	codeSpace = uint64(62 * 62 * 62 * 62 * 62 * 62) // 62^6 =~ 56.8 B
	codeLen   = 6
)

// Base62Encode encodes n into a fixed-length 6-character base62 string.
func Base62Encode(n uint64) string {
	buf := make([]byte, codeLen)
	for i := codeLen - 1; i >= 0; i-- {
		buf[i] = alphabet[n%62]
		n /= 62
	}
	return string(buf)
}

// SnowflakeCodeGenerator implements CodeGenerator using Snowflake IDs
// projected into the 62^6 space via modulo.
type SnowflakeCodeGenerator struct {
	node *snowflake.Node
}

var _ CodeGenerator = (*SnowflakeCodeGenerator)(nil)

// NewSnowflakeCodeGenerator creates a generator using snowflake node 1.
// For distributed deployments, configure a unique node ID (0–1023) per host.
func NewSnowflakeCodeGenerator() (*SnowflakeCodeGenerator, error) {
	node, err := snowflake.NewNode(1) // TODO: in distributed system, change it into the number of node
	if err != nil {
		return nil, err
	}
	return &SnowflakeCodeGenerator{node: node}, nil
}

func (g *SnowflakeCodeGenerator) Generate() (string, error) {
	id := uint64(g.node.Generate().Int64())
	return Base62Encode(id % codeSpace), nil
}
