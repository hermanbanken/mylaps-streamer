package main

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type jsonFormatterAppEngine struct {
	*log.JSONFormatter
}

func (f *jsonFormatterAppEngine) Format(entry *log.Entry) ([]byte, error) {
	entry.Data["severity"] = strings.ToUpper(entry.Level.String())
	return f.JSONFormatter.Format(entry)
}
func init() {
	log.SetFormatter(&jsonFormatterAppEngine{
		&log.JSONFormatter{FieldMap: log.FieldMap{
			log.FieldKeyTime:  "timestamp",
			log.FieldKeyLevel: "level",
			log.FieldKeyMsg:   "message",
			log.FieldKeyFunc:  "function",
		},
		},
	})
	log.SetLevel(log.DebugLevel)
}
