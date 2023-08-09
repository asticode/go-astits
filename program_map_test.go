package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgramMap(t *testing.T) {
	pm := newProgramMap()
	assert.False(t, pm.existsUnlocked(1))
	pm.setUnlocked(1, 1)
	assert.True(t, pm.existsUnlocked(1))
	pm.unsetUnlocked(1)
	assert.False(t, pm.existsUnlocked(1))
}
