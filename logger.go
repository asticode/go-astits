package astits

import "github.com/asticode/go-astikit"

// Right now we use a global logger because it feels weird to inject a logger in pure functions
// Indeed, logger is only needed to let the developer know when an unhandled descriptor or id has been found
// in the stream
var logger = astikit.AdaptStdLogger(nil)

func SetLogger(l astikit.StdLogger) { logger = astikit.AdaptStdLogger(l) }
