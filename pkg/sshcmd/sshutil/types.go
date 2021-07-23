package sshutil

import "time"

type SSH struct {
	User         string
	Password     string
	PkFile       string
	PkPassword   string
	OriginalPass string
	UserPass     map[string]string
	Timeout      *time.Duration
}
