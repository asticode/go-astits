package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElementaryStreamMap(t *testing.T) {
	esm := newElementaryStreamMap()
	assert.False(t, esm.existsUnlocked(0x16))
	esm.setUnlocked(0x16, 1)
	assert.True(t, esm.existsUnlocked(0x16))
	esm.unsetUnlocked(0x16)
	assert.False(t, esm.existsUnlocked(0x16))
}
