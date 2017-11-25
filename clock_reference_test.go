package astits

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var clockReference = newClockReference(3271034319, 58)

func TestClockReference(t *testing.T) {
	assert.Equal(t, 36344825768814*time.Nanosecond, clockReference.Duration())
	assert.Equal(t, int64(36344), clockReference.Time().Unix())
}
