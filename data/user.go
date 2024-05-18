package data

import (
	"context"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
)

type User struct {
	ID    string
	Email string
	// IsAdmin bool
}

func CreateUser(users *Users) error {
	_, err := db.Bun.NewInsert().Model(users).Exec(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func DeleteUser(userID string) error {
	_, err := db.Bun.Exec("DELETE FROM auth.users WHERE id = ?", userID)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(userID string) (*Users, error) {
	var user Users

	query := "SELECT id, aud, email, encrypted_password, email_confirmed_at, created_at, updated_at FROM auth.users WHERE id = ?"

	err := db.Bun.NewRaw(query, userID).Scan(context.Background(), &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
