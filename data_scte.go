package astits

import "github.com/asticode/go-astikit"

func parseSCTE35Payload(it *astikit.BytesIterator, h *PSISectionHeader) ([]byte, error) {
	header := make([]byte, 3)
	header[0] = byte(h.TableID)
	if h.SectionSyntaxIndicator {
		header[1] |= 0x80
	}
	if h.PrivateBit {
		header[1] |= 0x40
	}

	// reserved/sap_type for scte, flipping to 1s for now
	header[1] |= 0x30

	header[1] |= uint8((h.SectionLength >> 8) & 0x0f)
	header[2] = uint8(h.SectionLength & 0xff)

	buf, err := it.NextBytes(int(h.SectionLength))
	if err != nil {
		return nil, err
	}

	payload := make([]byte, 0, len(header)+int(h.SectionLength))
	payload = append(payload, header...)
	payload = append(payload, buf...)
	return payload, nil
}

func isSCTE35(payload []byte) bool {
	if len(payload) == 0 {
		return false
	}
	tableIDIndex := int(payload[0]) + 1
	if len(payload) < tableIDIndex {
		return false
	}
	return payload[tableIDIndex] == byte(PSITableIDSCTE35)
}
