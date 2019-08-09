package astits

import (
	"testing"
	"time"

	astibyte "github.com/asticode/go-astitools/byte"
	"github.com/stretchr/testify/assert"
)

var (
	dvbDurationMinutes      = time.Hour + 45*time.Minute
	dvbDurationMinutesBytes = []byte{0x1, 0x45} // 0145
	dvbDurationSeconds      = time.Hour + 45*time.Minute + 30*time.Second
	dvbDurationSecondsBytes = []byte{0x1, 0x45, 0x30} // 014530
	dvbTime, _              = time.Parse("2006-01-02 15:04:05", "1993-10-13 12:45:00")
	dvbTimeBytes            = []byte{0xc0, 0x79, 0x12, 0x45, 0x0} // C079124500
)

func TestParseDVBTime(t *testing.T) {
	d, err := parseDVBTime(astibyte.NewIterator(dvbTimeBytes))
	assert.Equal(t, dvbTime, d)
	assert.NoError(t, err)
}

func TestParseDVBDurationMinutes(t *testing.T) {
	d, err := parseDVBDurationMinutes(astibyte.NewIterator(dvbDurationMinutesBytes))
	assert.Equal(t, dvbDurationMinutes, d)
	assert.NoError(t, err)
}

func TestParseDVBDurationSeconds(t *testing.T) {
	d, err := parseDVBDurationSeconds(astibyte.NewIterator(dvbDurationSecondsBytes))
	assert.Equal(t, dvbDurationSeconds, d)
	assert.NoError(t, err)
}
