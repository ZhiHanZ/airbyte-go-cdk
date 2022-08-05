package airbyte

import (
	"bufio"
)

// Destination is the only interface you need to define to create your destination!
type Destination interface {
	// Spec returns the input "form" spec needed for your source
	Spec(logTracker LogTracker) (*ConnectorSpecification, error)
	// Check verifies the source - usually verify creds/connection etc.
	Check(srcCfgPath string, logTracker LogTracker) error
	// Write would receive synced messages and produce state message when sync go successful
	Write(srcCfgPath string, configuredCat *ConfiguredCatalog, messages *bufio.Scanner, logTracker LogTracker) error
}
