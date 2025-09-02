package main

import "time"

type externalProcessCredentialResult struct {
	Version         int
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}
