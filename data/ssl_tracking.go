package data

import "time"

type SSLTracking struct {
	ID         int
	DomainName string
	Issuer     string
	Expires    time.Time
	Status     string
}
