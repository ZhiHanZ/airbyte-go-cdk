package airbyte

import (
	"bufio"
	"io"
	"log"
	"os"
)

// DestinationRunner acts as an "orchestrator" of sorts to run your destination for you
type DestinationRunner struct {
	w          io.Writer
	dst        Destination
	msgTracker MessageTracker
}

// NewDestinationRunner takes your defined destination and plugs it in with the rest of airbyte
func NewDestinationRunner(dst Destination, w io.Writer) DestinationRunner {
	w = newSafeWriter(w)
	msgTracker := MessageTracker{
		Record: newRecordWriter(w),
		State:  newStateWriter(w),
		Log:    newLogWriter(w),
	}

	return DestinationRunner{
		w:          w,
		dst:        dst,
		msgTracker: msgTracker,
	}
}

// Start starts your destination
// Example usage would look like this in your main.go
//  func() main {
// 	src := newCoolDestination()
// 	runner := airbyte.NewDestinationRunner(src)
// 	err := runner.Start()
// 	if err != nil {
// 		log.Fatal(err)
// 	 }
//  }
// Yes, it really is that easy!
func (dr DestinationRunner) Start() error {
	switch cmd(os.Args[1]) {
	case cmdSpec:
		spec, err := dr.dst.Spec(LogTracker{
			Log: dr.msgTracker.Log,
		})
		if err != nil {
			_ = dr.msgTracker.Log(LogLevelError, "failed"+err.Error())
			return err
		}
		return write(dr.w, &message{
			Type:                   msgTypeSpec,
			ConnectorSpecification: spec,
		})

	case cmdCheck:
		inP, err := getSourceConfigPath()
		if err != nil {
			return err
		}
		err = dr.dst.Check(inP, LogTracker{
			Log: dr.msgTracker.Log,
		})
		if err != nil {
			log.Println(err)
			return write(dr.w, &message{
				Type: msgTypeConnectionStat,
				connectionStatus: &connectionStatus{
					Status: checkStatusFailed,
				},
			})
		}

		return write(dr.w, &message{
			Type: msgTypeConnectionStat,
			connectionStatus: &connectionStatus{
				Status: checkStatusSuccess,
			},
		})

	case cmdWrite:
		inP, err := getSourceConfigPath()
		if err != nil {
			return err
		}
		var incat ConfiguredCatalog
		p, err := getCatalogPath()
		if err != nil {
			return err
		}

		err = UnmarshalFromPath(p, &incat)
		if err != nil {
			return err
		}
		s := bufio.NewScanner(os.Stdin)
		err = dr.dst.Write(inP, &incat, s, LogTracker{Log: dr.msgTracker.Log})
		if err != nil {
			log.Println("failed")
			return err
		}
	}

	return nil
}
