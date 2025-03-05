# go-astits Go MPEGTS Parser

## Unreleased

- [SA-3344] guard against nil adaptation field

## v1.9.0
 - [SA-3019] Add UnmarshalPacketWithoutPayload function

## v1.8.0
 - Length check array length before index use

## v1.7.0
 - [SA-2570] Add ParsePacketWithoutPayload function

## v1.6.0 
 - Implement Serialise for Packet Adaptation Field

## v1.5.0
 - Add serialisation methods for PAT and PMT packets
 - Make ParseData public
 - Make `PacketPool` and `(p *PacketPool) Add()` public
 - Add all stream types values
 - Copy data into `d.originalBytes` in `parseDescriptors`, as we cannot count on it persisting
 - Update go.mod and go.sum with correct package names
