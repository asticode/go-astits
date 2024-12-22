package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementaryStreamMap(t *testing.T) {
	esm := newElementaryStreamMap()
	assert.False(t, esm.existsLocked(0x16))
	esm.setLocked(0x16, 1)
	assert.True(t, esm.existsLocked(0x16))
	esm.unsetLocked(0x16)
	assert.False(t, esm.existsLocked(0x16))
}
