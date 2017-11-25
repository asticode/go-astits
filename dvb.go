package astits

import (
	"fmt"
	"time"
)

// parseDVBTime parses a DVB time
// This field is coded as 16 bits giving the 16 LSBs of MJD followed by 24 bits coded as 6 digits in 4 - bit Binary
// Coded Decimal (BCD). If the start time is undefined (e.g. for an event in a NVOD reference service) all bits of the
// field are set to "1".
// I apologize for the computation which is really messy but details are given in the documentation
// Page: 160 | Annex C | Link: https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
func parseDVBTime(i []byte, offset *int) (t time.Time) {
	// Date
	var mjd = uint16(i[*offset])<<8 | uint16(i[*offset+1])
	var yt = int((float64(mjd) - 15078.2) / 365.25)
	var mt = int((float64(mjd) - 14956.1 - float64(int(float64(yt)*365.25))) / 30.6001)
	var d = int(float64(mjd) - 14956 - float64(int(float64(yt)*365.25)) - float64(int(float64(mt)*30.6001)))
	var k int
	if mt == 14 || mt == 15 {
		k = 1
	}
	var y = yt + k
	var m = mt - 1 - k*12
	t, _ = time.Parse("06-01-02", fmt.Sprintf("%d-%d-%d", y, m, d))
	*offset += 2

	// Time
	t = t.Add(parseDVBDuration(i, offset))
	return
}

// parseDVBDuration parses a duration
// 24 bit field containing the duration of the event in hours, minutes, seconds. format: 6 digits, 4 - bit BCD = 24 bit
func parseDVBDuration(i []byte, offset *int) (d time.Duration) {
	d = parseDVBDurationByte(i[*offset])*time.Hour + parseDVBDurationByte(i[*offset+1])*time.Minute + parseDVBDurationByte(i[*offset+2])*time.Second
	*offset += 3
	return
}

// parseDVBDurationByte parses a duration byte
func parseDVBDurationByte(i byte) time.Duration {
	return time.Duration(uint8(i)>>4*10 + uint8(i)&0xf)
}
