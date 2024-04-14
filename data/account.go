package data

import (
	"context"
	"fmt"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/nedpals/supabase-go"
)

type Plan int

func (p Plan) String() string {
	switch p {
	case PlanStarter:
		return "starter"
	case PlanBusiness:
		return "business"
	case PlanEnterprise:
		return "enterprise"
	default:
		return "unknown"
	}
}

func StringToPlan(planStr string) (Plan, error) {
	switch planStr {
	case "starter":
		return PlanStarter, nil
	case "business":
		return PlanBusiness, nil
	case "enterprise":
		return PlanEnterprise, nil
	default:
		return 0, fmt.Errorf("unknown plan: %s", planStr)
	}
}

const (
	PlanStarter Plan = iota + 1
	PlanBusiness
	PlanEnterprise
)

type Account struct {
	ID                 int64 `bun:",pk,autoincrement"`
	UserID             string
	SubscriptionStatus string
	Plan               Plan
	NotifyUpfront      int
	NotifyDefaultEmail string
}

type Plans struct {
	ID   int64 `bun:",pk,autoincrement"`
	Name string
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

// func GetAccount(query fiber.Map) (*Account, error) {
// 	account := new(Account)
// 	builder := db.Bun.NewSelect().Model(account)
// 	for k, v := range query {
// 		builder.Where("? = ?", bun.Ident(k), v)
// 	}
// 	err := builder.Scan(context.Background())
// 	return account, err
// }

func UpdateAccount(acc *Account) error {
	_, err := db.Bun.NewUpdate().
		Model(acc).
		WherePK().
		Exec(context.Background())
	return err
}

func CreateAccountForUserIfNotExist(user *supabase.User, selectedPlan string, subscriptionStatus string) (*Account, error) {
	if acc, err := GetUserAccount(user.ID); err == nil {
		return acc, nil
	}

	plan, err := StringToPlan(selectedPlan)
	if err != nil {
		return nil, err
	}

	acc := Account{
		UserID:             user.ID,
		NotifyUpfront:      7,
		NotifyDefaultEmail: user.Email,
		Plan:               Plan(plan),
		SubscriptionStatus: subscriptionStatus,
	}
	_, err = db.Bun.NewInsert().Model(&acc).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	logger.Log("event", "new account signup", "id", acc.ID)
	return &acc, nil
}
