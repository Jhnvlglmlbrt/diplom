package util

import (
	"fmt"
	"time"
)

func DaysLeft(t time.Time) string {
	timeZero := time.Time{}
	if t.Equal(timeZero) {
		return "n/a"
	}
	return fmt.Sprintf("%d days", time.Until(t)/(time.Hour*24))
}

func Pluralize(word string, amount int) string {
	if amount == 1 {
		return word
	}

	return word + "s"
}
