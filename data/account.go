package data

import (
	"context"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/nedpals/supabase-go"
)

const (
	PlanFree     = "STARTER"
	PlanStarter  = "BUSINESS"
	PlanBusiness = "CORPORATIVE"
)

type Account struct {
	ID                 string `bun:",pk,autoincrement"`
	UserID             string
	Plan               string
	NotifyUpfront      int
	NotifyDefaultEmail string
}

func CreateAccountForUserIfNotExists(user *supabase.User) (*Account, error) {
	if acc, err := GetUserAccount(user.ID); err == nil {
		return acc, err
	}

	acc := Account{
		UserID:             user.ID,
		NotifyUpfront:      7,
		NotifyDefaultEmail: user.Email,
		Plan:               PlanFree,
	}

	_, err := db.Bun.NewInsert().Model(&acc).Exec(context.Background())
	if err != nil {
		return nil, err
	}

	logger.Log("event", "new account signup", "id", acc.ID)
	return &acc, nil
}

func GetUserAccount(userID string) (*Account, error) {
	account := new(Account)
	ctx := context.Background()
	err := db.Bun.NewSelect().
		Model(account).
		Where("user_id = ?", userID).
		Scan(ctx)

	return account, err
}

func UpdateAccount(acc *Account) error {
	_, err := db.Bun.NewUpdate().
		Model(acc).
		WherePK().
		Exec(context.Background())

	return err
}
