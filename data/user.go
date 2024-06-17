package data

import (
	"context"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/uptrace/bun"
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

func GetEmailsForUserIDs(userIDs []string) (map[string]string, error) {
	users, err := GetAllUsers(userIDs)
	if err != nil {
		return nil, err
	}
	emailMap := make(map[string]string)
	for _, user := range users {
		emailMap[user.ID] = user.Email
	}
	return emailMap, nil
}

func GetAllUsers(userIDs []string) (map[string]*Users, error) {
	users := make(map[string]*Users)

	query := `
        SELECT id, aud, email, encrypted_password, email_confirmed_at, created_at, updated_at
        FROM auth.users
        WHERE id IN (?)
    `

	var userList []Users
	err := db.Bun.NewRaw(query, bun.In(userIDs)).Scan(context.Background(), &userList)
	if err != nil {
		return nil, err
	}

	for _, user := range userList {
		users[user.ID] = &user
	}

	return users, nil
}
