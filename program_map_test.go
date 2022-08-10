package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgramMap(t *testing.T) {
	pm := newProgramMap()
	assert.False(t, pm.exists(1))
	pm.set(1, 1)
	assert.True(t, pm.exists(1))
}
