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
	ID                 int64  `bun:"id,pk,autoincrement"`
	UserID             string `bun:"user_id"`
	SubscriptionStatus string
	NotifyUpfront      int
	NotifyDefaultEmail string
	PlanID             Plan
}

type Plans struct {
	ID          int64    `bun:"id,pk,autoincrement"`
	Name        string   `bun:"name"`
	Price       float64  `bun:"price"`
	Description string   `bun:"description"`
	Features    []string `bun:"features,type:text[]"`
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

func GetAccountByEmail(email string) (*Account, error) {
	account := new(Account)
	ctx := context.Background()
	err := db.Bun.NewSelect().
		Model(account).
		Where("notify_default_email = ?", email).
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

func CreateAccountForUserIfNotExist(user *supabase.User, selectedPlan string, subscriptionStatus string) (*Account, error) {
	if exists, err := PlanExistsByName(selectedPlan); err != nil || !exists {
		return nil, err
	}

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
		PlanID:             Plan(plan),
		SubscriptionStatus: subscriptionStatus,
	}
	_, err = db.Bun.NewInsert().Model(&acc).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	logger.Log("event", "new account signup", "id", acc.ID)
	return &acc, nil
}

func PlanExistsByName(planName string) (bool, error) {
	plan := new(Plans)
	err := db.Bun.NewSelect().Model(plan).Where("name = ?", planName).Scan(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed to check if plan exists by name: %w", err)
	}

	return true, nil
}

func GetAllPlans() ([]Plans, error) {
	var plans []Plans

	err := db.Bun.NewSelect().Model(&plans).Scan(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all plans: %w", err)
	}

	return plans, nil
}
