package cmd

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestResolveLogLevel(t *testing.T) {
	cases := []struct {
		name           string
		debug, verbose bool
		logLevel       string
		want           logrus.Level
	}{
		{"default info", false, false, "info", logrus.InfoLevel},
		{"debug overrides", true, false, "warn", logrus.DebugLevel},
		{"verbose overrides logLevel", false, true, "error", logrus.InfoLevel},
		{"explicit warn", false, false, "warn", logrus.WarnLevel},
		{"explicit debug string", false, false, "debug", logrus.DebugLevel},
		{"unparseable falls back to info", false, false, "bogus", logrus.InfoLevel},
	}
	for _, c := range cases {
		if got := resolveLogLevel(c.debug, c.verbose, c.logLevel); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}
}
