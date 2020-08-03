package file

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gargath/pleiades/pkg/log"
	"github.com/gargath/pleiades/pkg/publisher"
	"github.com/gargath/pleiades/pkg/sse"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const moduleName = "filepublisher"

var (
	eventsPublished = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pleiades_publish_events_total",
			Help: "The total number of events published to filesystem"})

	pubErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pleiades_publish_file_errors_total",
			Help: "Total numbers of errors encountered while publishing to filesystem",
		},
		[]string{"type"})

	logger = log.MustGetLogger(moduleName)
)

// NewPublisher returns a Publisher initialized with the source channel and destination path provided
func NewPublisher(src <-chan *sse.Event, dest string) (publisher.Publisher, error) {
	if src == nil {
		return nil, ErrNilChan
	}
	if dest == "" {
		return nil, ErrNoDest
	}
	o, err := os.Stat(dest)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(dest, 0755)
		if errDir != nil {
			logger.Fatalf("failed to create destination directory: %v", errDir)
			panic(err)
		}
	} else if o.Mode().IsRegular() {
		logger.Errorf("destination path %s exists and is file", dest)
		return nil, fmt.Errorf("destination path %s exists as file", dest)
	}
	f := &Publisher{
		source:      src,
		destination: dest,
	}
	return f, nil
}

// ReadAndPublish will read Events from the input channel and write them to file
// File names are sequential and relative to the destination directory
// If the FilePublisher's destionation directory is not set, ReadAndPublish returns ErrNoDest
//
// Calling ReadAndPublish() will reset the processed message counter of the underlying Publisher and
// returns the value of the counter when the Publisher's source channel is closed
func (f *Publisher) ReadAndPublish() (int64, error) {
	f.msgCount = 0
	for e := range f.source {
		f.msgCount++
		if e != nil {
			err := f.ProcessEvent(e)
			if err != nil {
				return f.msgCount, fmt.Errorf("error processing event: %v", err)
			}
		}
	}
	return f.msgCount, nil
}

// ProcessEvent writes a single event to a file
func (f *Publisher) ProcessEvent(e *sse.Event) error {
	eventsPublished.Inc()
	d, err := ioutil.ReadAll(e.GetData())
	if err != nil {
		pubErrors.WithLabelValues("event_data_read").Inc()
		return fmt.Errorf("error reading event data: %v", err)
	}
	err = ioutil.WriteFile(fmt.Sprintf("%s/event-%d.dat", f.destination, f.msgCount), d, 0644)
	if err != nil {
		pubErrors.WithLabelValues("file_write").Inc()
		return fmt.Errorf("error writing file: %v", err)
	}
	return nil
}
