package gopxgrid

import (
	"github.com/go-stomp/stomp/v3"
)

type internalLogger struct{}

var _ stomp.Logger = (*internalLogger)(nil)

func newInternalLogger() stomp.Logger {
	return &internalLogger{}
}

func (l *internalLogger) Debugf(string, ...interface{})   {}
func (l *internalLogger) Infof(string, ...interface{})    {}
func (l *internalLogger) Warningf(string, ...interface{}) {}
func (l *internalLogger) Errorf(string, ...interface{})   {}
func (l *internalLogger) Debug(string)                    {}
func (l *internalLogger) Info(string)                     {}
func (l *internalLogger) Warning(string)                  {}
func (l *internalLogger) Error(string)                    {}
