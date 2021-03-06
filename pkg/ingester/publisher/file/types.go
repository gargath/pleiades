package file

import (
	"fmt"

	"github.com/gargath/pleiades/pkg/ingester/sse"
)

// Publisher reads Events and writes them to disk
type Publisher struct {
	destination string
	source      <-chan *sse.Event
	msgCount    int64
	prefix      string
	lastEventID string
}

// Opts hold config options for the file publisher
type Opts struct {
	Destination string
}

// PublisherConfig contains configuration for the file Publisher
type PublisherConfig struct {
	Destination string
}

// ErrNoDest indicates that the FilePublisher has no destination path
var ErrNoDest error = fmt.Errorf("No destination path set")

// ErrNilChan indicates that the FilePublisher has no source channel
var ErrNilChan error = fmt.Errorf("Source channel is nil")
