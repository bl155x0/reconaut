package iobuffer

import (
	"sync"
)

type IOBuffer struct {
	outputQueue []Output
	Verbose     bool
}

type Output struct {
	text    string
	verbose bool
}

var (
	singletonIOBuffer *IOBuffer
	once              sync.Once
	mutex             sync.Mutex
)

// GetIOBuffer returns a singleton IOBuffer, i.e. it is ensured that this fuction will
// always return the same instance
func GetIOBuffer() *IOBuffer {
	mutex.Lock()
	defer mutex.Unlock()
	once.Do(func() {
		singletonIOBuffer = &IOBuffer{}
	})
	return singletonIOBuffer
}

// AddOutputVerbose adds output to the buffer which is then printed out to STDOUT
func (ioBuffer *IOBuffer) AddOutputVerbose(outputString string) {
	ioBuffer.addOutput(outputString, true)
}

// AddOutputVerbose adds output to the buffer which is then printed out to STDOUT
func (ioBuffer *IOBuffer) AddOutput(outputString string) {
	ioBuffer.addOutput(outputString, false)
}

// AddOutputVerbose adds output to the buffer which is then printed out to STDOUT
func (ioBuffer *IOBuffer) addOutput(outputString string, verbose bool) {
	mutex.Lock()
	defer mutex.Unlock()
	output := Output{
		text:    outputString,
		verbose: verbose,
	}
	ioBuffer.outputQueue = append(ioBuffer.outputQueue, output)
}

// GetOutput returns the next output from the buffer
func (ioBuffer *IOBuffer) GetOutput() *string {
	mutex.Lock()
	defer mutex.Unlock()
	if len(ioBuffer.outputQueue) == 0 {
		return nil
	}

	output := ioBuffer.outputQueue[0]
	ioBuffer.outputQueue = ioBuffer.outputQueue[1:]

	if output.verbose {
		if ioBuffer.Verbose == false {
			return nil
		}
	}
	return &output.text
}
