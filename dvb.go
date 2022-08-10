package astits

import (
	"fmt"
	"strconv"
	"time"

	"github.com/icza/bitio"
)

// parseDVBTime parses a DVB time
// This field is coded as 16 bits giving the 16 LSBs of MJD
// followed by 24 bits coded as 6 digits in 4 - bit Binary
// Coded Decimal (BCD). If the start time is undefined
// (e.g. for an event in a NVOD reference service)
// all bits of the field are set to "1".
//
// Page: 160 | Annex C | Link:
// https://www.dvb.org/resources/public/standards/a38_dvb-si_specification.pdf
func parseDVBTime(r *bitio.CountReader) (time.Time, error) {
	// Date
	mjd := uint16(r.TryReadBits(16))
	yt := int((float32(mjd) - 15078.2) / 365.25)
	mt := int((float64(mjd) - 14956.1 - float64(uint16(float64(yt)*365.25))) / 30.6001)
	d := int(mjd - 14956 - uint16(float64(yt)*365.25) - uint16(float64(mt)*30.6001))
	var k int
	if mt == 14 || mt == 15 {
		k = 1
	}
	y := yt + k
	m := mt - 1 - k*12

	dateStr := strconv.Itoa(y) + "-" + strconv.Itoa(m) + "-" + strconv.Itoa(d)
	t, _ := time.Parse("06-01-02", dateStr)

	s, err := parseDVBDurationSeconds(r)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing DVB duration seconds failed: %w", err)
	}

	t = t.Add(s)
	return t, r.TryError
}

// parseDVBDurationMinutes parses a minutes duration.
// 16 bit field containing the duration of the event in
// hours, minutes. format: 4 digits, 4 - bit BCD = 18 bit.
func parseDVBDurationMinutes(r *bitio.CountReader) (time.Duration, error) {
	d := parseDVBDurationByte(r.TryReadByte())*time.Hour + //nolint:durationcheck
		parseDVBDurationByte(r.TryReadByte())*time.Minute //nolint:durationcheck

	return d, r.TryError
}

// parseDVBDurationSeconds parses a seconds duration.
// 24 bit field containing the duration of the event in hours,
// minutes, seconds. format: 6 digits, 4 - bit BCD = 24 bit.
func parseDVBDurationSeconds(r *bitio.CountReader) (time.Duration, error) {
	d := parseDVBDurationByte(r.TryReadByte())*time.Hour + //nolint:durationcheck
		parseDVBDurationByte(r.TryReadByte())*time.Minute + //nolint:durationcheck
		parseDVBDurationByte(r.TryReadByte())*time.Second //nolint:durationcheck

	return d, r.TryError
}

// parseDVBDurationByte parses a duration byte.
func parseDVBDurationByte(i byte) time.Duration {
	return time.Duration(i>>4*10 + i&0xf)
}

func writeDVBTime(w *bitio.Writer, t time.Time) (int, error) {
	year := t.Year() - 1900
	month := t.Month()
	day := t.Day()

	l := 0
	if month <= time.February {
		l = 1
	}

	mjd := 14956 + day + int(float64(year-l)*365.25) + int(float64(int(month)+1+l*12)*30.6001)

	d := t.Sub(t.Truncate(24 * time.Hour))

	if err := w.WriteBits(uint64(mjd), 16); err != nil {
		return 0, err
	}
	bytesWritten, err := writeDVBDurationSeconds(w, d)
	if err != nil {
		return 2, err
	}

	return bytesWritten + 2, nil
}

func writeDVBDurationMinutes(w *bitio.Writer, d time.Duration) error {
	hours := uint8(d.Hours())
	minutes := uint8(int(d.Minutes()) % 60)

	w.TryWriteByte(dvbDurationByteRepresentation(hours))
	w.TryWriteByte(dvbDurationByteRepresentation(minutes))

	return w.TryError
}

func writeDVBDurationSeconds(w *bitio.Writer, d time.Duration) (int, error) {
	hours := uint8(d.Hours())
	minutes := uint8(int(d.Minutes()) % 60)
	seconds := uint8(int(d.Seconds()) % 60)

	w.TryWriteByte(dvbDurationByteRepresentation(hours))
	w.TryWriteByte(dvbDurationByteRepresentation(minutes))
	w.TryWriteByte(dvbDurationByteRepresentation(seconds))

	return 3, w.TryError
}

func dvbDurationByteRepresentation(n uint8) uint8 {
	return (n/10)<<4 | n%10
}
