[![GoReportCard](http://goreportcard.com/badge/github.com/asticode/go-astits)](http://goreportcard.com/report/github.com/asticode/go-astits)
[![GoDoc](https://godoc.org/github.com/asticode/go-astits?status.svg)](https://godoc.org/github.com/asticode/go-astits)
[![Travis](https://travis-ci.org/asticode/go-astits.svg?branch=master)](https://travis-ci.org/asticode/go-astits#)
[![Coveralls](https://coveralls.io/repos/github/asticode/go-astits/badge.svg?branch=master)](https://coveralls.io/github/asticode/go-astits)

This is a Golang library to natively parse and demux MPEG Transport Streams (ts) in GO.

# Installation

To install the library use the following:

    go get -u github.com/asticode/go-astits/...
    
# Before looking at the code...

The transport stream is made of packets.<br>
Each packet has a header, an optional adaptation field and a payload.<br>
Several payloads can be appended and parsed as a data.

```
                                           TRANSPORT STREAM
 +--------------------------------------------------------------------------------------------------+
 |                                                                                                  |
 
                       PACKET                                         PACKET
 +----------------------------------------------+----------------------------------------------+----
 |                                              |                                              |
 
 +--------+---------------------------+---------+--------+---------------------------+---------+
 | HEADER | OPTIONAL ADAPTATION FIELD | PAYLOAD | HEADER | OPTIONAL ADAPTATION FIELD | PAYLOAD | ...
 +--------+---------------------------+---------+--------+---------------------------+---------+
 
                                      |         |                                    |         |
                                      +---------+                                    +---------+
                                           |                                              |
                                           +----------------------------------------------+
                                                                DATA
```
    
# Using the library in your code

WARNING: the code below doesn't handle errors for readibility purposes. However you SHOULD!

```go
// Create a cancellable context in case you want to stop reading packets/data any time you want
ctx, cancel := context.WithCancel(context.Background())

// Handle SIGTERM signal
ch := make(chan os.Signal, 1)
signal.Notify(ch, syscall.SIGTERM)
go func() {
    <-ch
    cancel()
}()

// Open your file or initialize any kind of io.Reader
f, _ := os.Open("/path/to/file.ts")
defer f.Close()

// Create the demuxer
dmx := astits.New(ctx, f)
for {
    // Get the next data
    d, _ := dmx.NextData()
    
    // Data is a PMT data
    if d.PMT != nil {
        // Loop through elementary streams
        for _, es := range d.PMT.ElementaryStreams {
                fmt.Printf("Stream detected: %d\n", es.ElementaryPID)
        }
        return
    }
}
```

## Options

In order to pass options to the demuxer, look for the methods prefixed with `Opt` and add them upon calling `New`:

```go
// This is your custom packets parser
p := func(ps []*astits.Packet) (ds []*astits.Data, skip bool, err error) {
        // This is your logic
        skip = true
        return
}

// Now you can create a demuxer with the proper options
dmx := New(ctx, f, OptPacketSize(192), OptPacketsParser(p))
```

# CLI

This library provides a CLI that will automatically get installed in `GOPATH/bin` on `go get` execution.

## List streams

    $ astits -i <path to your file> -f <format: text|json (default: text)>
    
## List data

    $ astits data -i <path to your file> -d <data type: eit|nit|... (repeatable argument | if empty, all data types are shown)>

# Features and roadmap

- [x] Parse PES packets
- [x] Parse PAT packets
- [x] Parse PMT packets
- [x] Parse EIT packets
- [x] Parse NIT packets
- [x] Parse SDT packets
- [x] Parse TOT packets
- [ ] Parse BAT packets
- [ ] Parse DIT packets
- [ ] Parse RST packets
- [ ] Parse SIT packets
- [ ] Parse ST packets
- [ ] Parse TDT packets
- [ ] Parse TSDT packets
