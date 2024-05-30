# go-astits Go MPEGTS Parser

## Unreleased

 - Length check array length before index use
 - [SA-2570] Add ParsePacketWithoutPayload function
 - Implement Serialise for Packet Adaptation Field

## v1.5.0
 - Add serialisation methods for PAT and PMT packets
 - Make ParseData public
 - Make `PacketPool` and `(p *PacketPool) Add()` public
 - Add all stream types values
 - Copy data into `d.originalBytes` in `parseDescriptors`, as we cannot count on it persisting
 - Update go.mod and go.sum with correct package names
