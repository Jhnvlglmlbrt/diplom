package engine

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/util"
	"github.com/gofiber/template/django/v3"
)

func CreateEngine() *django.Engine {
	engine := django.New("./views", ".html")
	engine.Reload(true)
	engine.AddFunc("css", func(name string) (res template.HTML) {
		filepath.Walk("./static/assets", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == name {
				res = template.HTML("<link rel=\"stylesheet\" href=\"/" + path + "\">")
			}
			return nil
		})
		return
	})

	engine.AddFunc("badgeForStatus", func(status string) (res string) {
		switch status {
		case data.StatusOffline:
			return fmt.Sprintf(`<div class="badge badge-accent">%s</div>`, status)
		case data.StatusHealthy:
			return fmt.Sprintf(`<div class="badge badge-success">%s</div>`, status)
		case data.StatusExpires:
			return fmt.Sprintf(`<div class="badge badge-warning">%s</div>`, status)
		case data.StatusExpired:
			return fmt.Sprintf(`<div class="badge badge-info">%s</div>`, status)
		case data.StatusUnresponsive:
			return fmt.Sprintf(`<div class="badge badge-accent">%s</div>`, status)
		case data.StatusInvalid:
			return fmt.Sprintf(`<div class="badge badge-error">%s</div>`, status)
		}
		return ""
	})

	engine.AddFunc("formatTime", func(t time.Time) (res string) {
		timeZero := time.Time{}
		if t.Equal(timeZero) {
			return "n/a"
		}
		return t.Format(time.DateTime)
	})

	engine.AddFunc("timeAgo", func(t time.Time) (res string) {
		x := time.Since(t).Seconds()
		return fmt.Sprintf("%v seconds ago", math.Round(x))
	})

	engine.AddFunc("daysLeft", func(t time.Time) (res string) {
		return util.DaysLeft(t)
	})

	engine.AddFunc("pluralize", func(s string, n int) (res string) {
		return util.Pluralize(s, n)
	})

	engine.AddFunc("getEmailForUserID", func(userID string) (res string) {
		user, err := data.GetUser(userID)
		if err != nil {
			return "Can't get a user by ID"
		}
		return user.Email
	})

	return engine
}
