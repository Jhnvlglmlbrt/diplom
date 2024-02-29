package app

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/gofiber/template/django/v3"
)

func CreateEngine() *django.Engine {
	engine := django.New("./views", ".html")
	engine.SetAutoEscape(false)
	engine.Reload(true)

	engine.AddFunc("css", func(name string) (res template.HTML) {
		filepath.Walk("static/assets", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == name {
				res = template.HTML("<link rel=\"stylesheet\" href=\"" + path + "\">")
			}
			return nil
		})
		return
	})

	engine.AddFunc("badgeForStatus", func(status string) (res string) {
		switch status {
		case data.StatusOffline:
			return fmt.Sprintf(`<div class="badge badge-error">%s</div>`, status)
		case data.StatusHealthy:
			return fmt.Sprintf(`<div class="badge badge-success">%s</div>`, status)
		case data.StatusExpires:
			return fmt.Sprintf(`<div class="badge badge-warning">%s</div>`, status)
		case data.StatusExpired:
			return fmt.Sprintf(`<div class="badge badge-error">%s</div>`, status)
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
	engine.AddFunc("daysLeft", func(t time.Time) (res string) {
		timeZero := time.Time{}
		if t.Equal(timeZero) {
			return "n/a"
		}
		return fmt.Sprintf("%d days", time.Until(t)/(time.Hour*24))
	})
	return engine
}
