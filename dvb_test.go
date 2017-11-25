package astits

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	dvbDuration      = time.Hour + 45*time.Minute + 30*time.Second
	dvbDurationBytes = []byte{0x1, 0x45, 0x30} // 014530
	dvbTime, _       = time.Parse("2006-01-02 15:04:05", "1993-10-13 12:45:00")
	dvbTimeBytes     = []byte{0xc0, 0x79, 0x12, 0x45, 0x0} // C079124500
)

func TestParseDVBTime(t *testing.T) {
	var offset int
	d := parseDVBTime(dvbTimeBytes, &offset)
	assert.Equal(t, dvbTime, d)
	assert.Equal(t, 5, offset)
}

func TestParseDVBDuration(t *testing.T) {
	var offset int
	d := parseDVBDuration(dvbDurationBytes, &offset)
	assert.Equal(t, dvbDuration, d)
	assert.Equal(t, 3, offset)
}
