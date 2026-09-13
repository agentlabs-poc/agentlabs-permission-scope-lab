package codec

import (
	"strings"
	"sync"
	"time"
)

// A Snowflake is a 64-bit, time-sortable identifier: a millisecond timestamp
// since a fixed epoch, a node id, and a per-millisecond sequence. agentlabs-auth
// generates them through go-common/id over bwmarrin/snowflake, and Auth-AL uses
// the same layout so the two services share one id scheme rather than two.
//
// The bit layout below is that library's: 41 bits of time, 10 of node, 12 of
// sequence.
const (
	snowflakeEpoch    = int64(1288834974657) // 2010-11-04T01:42:54.657Z
	snowflakeNodeBits = 10
	snowflakeStepBits = 12

	SnowflakeNodeMax = int64(-1) ^ (int64(-1) << snowflakeNodeBits)
	snowflakeStepMax = int64(-1) ^ (int64(-1) << snowflakeStepBits)
)

// Snowflakes generates ids for one node. It is safe for concurrent use: the
// sequence is what keeps two ids issued in the same millisecond distinct, so it
// has to be guarded.
type Snowflakes struct {
	mu   sync.Mutex
	now  func() time.Time
	node int64
	time int64
	step int64
}

// NewSnowflakes returns a generator for one node id. A node id outside the
// library's range is refused rather than silently masked, because two nodes
// believing they hold different ids while sharing one produces collisions.
func NewSnowflakes(node int64, now func() time.Time) (*Snowflakes, error) {
	if node < 0 || node > SnowflakeNodeMax {
		return nil, errNodeRange
	}
	if now == nil {
		now = time.Now
	}
	return &Snowflakes{now: now, node: node}, nil
}

// Next returns the next id, rendered base 36 — 12 characters today, 13 at the
// int64 ceiling. Rendering here rather than at the caller is what keeps one
// spelling of an id in the system.
func (s *Snowflakes) Next() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UnixMilli() - snowflakeEpoch
	if now == s.time {
		s.step = (s.step + 1) & snowflakeStepMax
		if s.step == 0 {
			// The sequence wrapped inside one millisecond: wait for the next
			// rather than reissue an id already handed out.
			for now <= s.time {
				now = s.now().UnixMilli() - snowflakeEpoch
			}
		}
	} else {
		s.step = 0
	}
	s.time = now
	id := now<<(snowflakeNodeBits+snowflakeStepBits) | s.node<<snowflakeStepBits | s.step
	return renderBase36(id)
}

func renderBase36(value int64) string {
	if value == 0 {
		return "0"
	}
	var out strings.Builder
	var digits [RoleIDMaxLen]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = RoleIDAlphabet[value%36]
		value /= 36
	}
	out.Write(digits[i:])
	return out.String()
}
