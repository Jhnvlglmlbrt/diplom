package data

import (
	"context"
	"fmt"
	"sort"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/nedpals/supabase-go"
	"github.com/uptrace/bun"
)

type Plan int

func (p Plan) String() string {
	switch p {
	case PlanStarter:
		return "Начальный"
	case PlanBusiness:
		return "Бизнес"
	case PlanEnterprise:
		return "Корпоративный"
	default:
		return "unknown"
	}
}

func StringToPlan(planStr string) (Plan, error) {
	switch planStr {
	case "Начальный":
		return PlanStarter, nil
	case "Бизнес":
		return PlanBusiness, nil
	case "Корпоративный":
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
	NotifyUpfront      int
	NotifyDefaultEmail string
	PlanID             Plan
	SubscriptionStatus string `bun:"subscription_status"`
}

type Plans struct {
	ID          int64    `bun:"id,pk,autoincrement"`
	Name        string   `bun:"name"`
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

func CountAccounts() (int, error) {
	return db.Bun.NewSelect().
		Model(&Account{}).
		Count(context.Background())
}

func CountPlans() (int, error) {
	return db.Bun.NewSelect().
		Model(&Plans{}).
		Count(context.Background())
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
		SubscriptionStatus: subscriptionStatus,
		PlanID:             Plan(plan),
	}

	logger.Log("msg", "account", acc)
	spew.Dump(acc)

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

	// Сортировка планов по ID (чтобы бизнес был в центре(костыль))
	sort.Slice(plans, func(i, j int) bool {
		return plans[i].ID < plans[j].ID
	})

	return plans, nil
}

func GetAccounts(limit int, page int, sortField string, ascending bool) ([]Account, error) {
	if limit == 0 {
		limit = DefaultLimit
	}

	var accounts []Account

	builder := db.Bun.NewSelect().Model(&accounts).Limit(limit)

	offset := (page - 1) * limit
	builder.Offset(offset)
	if ascending {
		builder.OrderExpr("? ASC", bun.Ident(sortField))
	} else {
		builder.OrderExpr("? DESC", bun.Ident(sortField))
	}
	err := builder.Scan(context.Background())
	return accounts, err
}

func GetPlans(limit int, page int, sortField string, ascending bool) ([]Plans, error) {
	if limit == 0 {
		limit = DefaultLimit
	}
	var plans []Plans

	builder := db.Bun.NewSelect().Model(&plans).Limit(limit)

	offset := (page - 1) * limit
	builder.Offset(offset)
	if ascending {
		builder.OrderExpr("? ASC", bun.Ident(sortField))
	} else {
		builder.OrderExpr("? DESC", bun.Ident(sortField))
	}
	err := builder.Scan(context.Background())
	return plans, err
}

func GetPlanByID(planID string) (*Plans, error) {
	plan := new(Plans)
	err := db.Bun.NewSelect().
		Model(plan).
		Where("id = ?", planID).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return plan, nil
}

func RemovePlan(planID string) error {
	_, err := db.Bun.NewDelete().
		Model(&Plans{}).
		Where("id = ?", planID).
		Exec(context.Background())
	return err
}

func UpdatePlan(plan *Plans) error {
	_, err := db.Bun.NewUpdate().
		Model(plan).
		WherePK().
		Exec(context.Background())
	return err
}
