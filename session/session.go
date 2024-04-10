package session

import (
	"github.com/gofiber/fiber/v2/middleware/session"
)

// InitSessionMiddleware инициализирует middleware для работы с сессиями.
func InitSessionMiddleware() *session.Store {
	return session.New()
}

func InitSessionMiddleware3() *session.Store {
	return session.New()
}
