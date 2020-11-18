package astits

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgramMap(t *testing.T) {
	pm := NewProgramMap()
	assert.False(t, pm.Exists(1))
	pm.Set(1, 1)
	assert.True(t, pm.Exists(1))
}
