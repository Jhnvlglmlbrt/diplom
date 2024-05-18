package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/gofiber/fiber/v2"
	"github.com/nedpals/supabase-go"
)

func HandleUsersList(c *fiber.Ctx) error {
	aud := "authenticated"
	count, err := data.CountUsers(aud)
	if err != nil {
		return err
	}
	logger.Log("msg", "count", count)
	if count == 0 {
		return c.Render("admin/users", fiber.Map{"usersExist": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}
	filterContext := buildFilterContext(filter)

	query := fiber.Map{
		"aud": aud,
	}

	users, err := data.GetUsers(query, filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	data := fiber.Map{
		"users":       users,
		"filters":     filterContext,
		"pages":       buildPages(count, filter.Limit),
		"usersExist":  true,
		"queryParams": filter.encode(),
	}
	return c.Render("admin/users", data)
}

func HandleUsersDelete(c *fiber.Ctx) error {
	var req struct {
		UsersIDs []string `json:"users_ids"`
	}

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	tx, err := db.Bun.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, userID := range req.UsersIDs {
		user, err := data.GetUser(userID)
		if err != nil {
			return err
		}

		if err := data.RemoveAccount(userID); err != nil {
			return err
		}

		// метод для удаления пользователя из auth.users (поскольку внешний ключ к аккаунт, то сначала удаление account)
		if err := data.DeleteUser(userID); err != nil {
			return err
		}

		// if err := data.RemoveUser(user.UserID); err != nil {
		// 	return err
		// }
		logger.Log("msg", "user deleted", user.Email)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Redirect("/admin/users")
}

func HandleAccountsList(c *fiber.Ctx) error {
	count, err := data.CountAccounts()
	if err != nil {
		return err
	}
	if count == 0 {
		return c.Render("admin/accounts", fiber.Map{"accountsExist": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}
	filterContext := buildFilterContext(filter)

	accounts, err := data.GetAccounts(filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	context := fiber.Map{
		"accounts":      accounts,
		"filters":       filterContext,
		"pages":         buildPages(count, filter.Limit),
		"accountsExist": true,
	}

	return c.Render("admin/accounts", context)
}

func HandleAccountsUpdate(c *fiber.Ctx) error {
	var req struct {
		AccountsIDs        []string `json:"accounts_ids"`
		NotifyUpfront      []string `json:"notify_upfront"`
		NotifyDefaultEmail []string `json:"notify_default_email"`
		PlanID             []string `json:"plan_id"`
		SubscriptionStatus []string `json:"subscription_status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	var notifyUpfrontInt []int
	for _, val := range req.NotifyUpfront {
		num, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("ошибка при преобразовании значения поля notify_upfront: %v", err)
		}
		notifyUpfrontInt = append(notifyUpfrontInt, num)
	}

	tx, err := db.Bun.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, accountID := range req.AccountsIDs {
		account, err := data.GetAccountByID(accountID)
		if err != nil {
			return err
		}

		account.NotifyUpfront = notifyUpfrontInt[i]
		account.NotifyDefaultEmail = req.NotifyDefaultEmail[i]
		account.PlanID, _ = data.StringToPlan(req.PlanID[i])
		account.SubscriptionStatus = req.SubscriptionStatus[i]

		if err := data.UpdateAccount(account); err != nil {
			return err
		}
		logger.Log("msg", "account changed", "accountID ", accountID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Redirect("/admin/accounts")
}

func HandleAccountsDelete(c *fiber.Ctx) error {
	var req struct {
		UsersIDs []string `json:"accounts_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	// spew.Dump(req)

	tx, err := db.Bun.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, userID := range req.UsersIDs {
		account, err := data.GetAccountByID(userID)
		logger.Log(userID)
		if err != nil {
			return err
		}

		if err := data.RemoveAccount(account.UserID); err != nil {
			return err
		}
		logger.Log("msg", "account deleted by admin", account.UserID)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Redirect("/admin/accounts")
}

func HandleUsersCreate(c *fiber.Ctx) error {
	var (
		selectedPlan = "Начальный"
	)
	email := c.FormValue("email")
	password := c.FormValue("password")

	// Создание пользователя
	admin := createSupabaseAdmin()
	userParams := supabase.AdminUserParams{
		Email:        email,
		Password:     &password,
		EmailConfirm: true,
	}
	user, err := admin.Admin.CreateUser(context.Background(), userParams)
	if err != nil {
		return err
	}

	users, err := data.GetUser(user.ID)
	if err != nil {
		return err
	}

	// создавать пользователя в таблице users
	// создавать аккаунт
	// вернуть на admin/users
	// Добавить создание аккаунта в users при  auth callback!!!

	users = &data.Users{
		ID:                users.ID,
		Aud:               users.Aud,
		Email:             users.Email,
		EncryptedPassword: users.EncryptedPassword,
		EmailConfirmedAt:  users.EmailConfirmedAt,
		CreatedAt:         users.CreatedAt,
		UpdatedAt:         users.UpdatedAt,
	}
	// err = data.CreateUser(users)
	// if err != nil {
	// 	return err
	// }

	// logger.Log("msg", "user created by admin", user.Email)

	_, err = CreateAccountForAdminIfNotExist(user, selectedPlan, data.SubscriptionStatusActive)
	if err != nil {
		return err
	}

	logger.Log("msg", "account created by admin", user.Email)

	return c.Redirect("/admin/users")
}

func CreateAccountForAdminIfNotExist(user *supabase.AdminUser, selectedPlan string, subscriptionStatus string) (*data.Account, error) {
	if exists, err := data.PlanExistsByName(selectedPlan); err != nil || !exists {
		return nil, err
	}

	if acc, err := data.GetUserAccount(user.ID); err == nil {
		return acc, nil
	}

	plan, err := data.StringToPlan(selectedPlan)
	if err != nil {
		return nil, err
	}

	acc := data.Account{
		UserID:             user.ID,
		NotifyUpfront:      7,
		NotifyDefaultEmail: user.Email,
		SubscriptionStatus: subscriptionStatus,
		PlanID:             data.Plan(plan),
	}

	logger.Log("msg", "account", acc)
	// spew.Dump(acc)

	_, err = db.Bun.NewInsert().Model(&acc).Exec(context.Background())
	if err != nil {
		return nil, err
	}
	logger.Log("event", "new account signup", "id", acc.ID)
	return &acc, nil
}

// func HandlePlansList(c *fiber.Ctx) error {
// 	count, err := data.CountPlans()
// 	if err != nil {
// 		return err
// 	}
// 	if count == 0 {
// 		return c.Render("admin/plans", fiber.Map{"accountsExist": false})
// 	}

// 	filter, err := buildTrackingFilter(c)
// 	if err != nil {
// 		return err
// 	}
// 	filterContext := buildFilterContext(filter)

// 	accounts, err := handlers.GetPlans(filter.Limit, filter.Page, "id", true)
// 	if err != nil {
// 		return err
// 	}

// 	context := fiber.Map{
// 		"accounts":      accounts,
// 		"filters":       filterContext,
// 		"pages":         buildPages(count, filter.Limit),
// 		"accountsExist": true,
// 	}

// 	return c.Render("admin/accounts", context)
// }

func HandlePlansList(c *fiber.Ctx) error {
	count, err := data.CountPlans()
	if err != nil {
		return err
	}
	if count == 0 {
		return c.Render("admin/plans", fiber.Map{"accountsExist": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}
	filterContext := buildFilterContext(filter)

	plans, err := data.GetPlans(filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	// spew.Dump(plans)
	context := fiber.Map{
		"plans":      plans,
		"filters":    filterContext,
		"pages":      buildPages(count, filter.Limit),
		"plansExist": true,
	}

	return c.Render("admin/plans", context)
}

func HandlePlansDelete(c *fiber.Ctx) error {
	var req struct {
		PlansIDs []string `json:"plans_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	// spew.Dump(req)

	tx, err := db.Bun.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, planID := range req.PlansIDs {
		plan, err := data.GetPlanByID(planID)
		// logger.Log(plan)
		if err != nil {
			return err
		}
		planID := strconv.FormatInt(plan.ID, 10)

		if err := data.RemovePlan(planID); err != nil {
			return err
		}
		logger.Log("msg", "plan deleted by admin", plan.Name)
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Redirect("/admin/plans")
}

func HandlePlansUpdate(c *fiber.Ctx) error {
	var req struct {
		PlansIDs    []string `json:"plans_ids"`
		Name        []string `json:"name"`
		Description []string `json:"description"`
		Features    []string `json:"features"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	spew.Dump(req)
	tx, err := db.Bun.BeginTx(c.Context(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, planID := range req.PlansIDs {
		plan, err := data.GetPlanByID(planID)
		if err != nil {
			return err
		}

		plan.Name = req.Name[i]
		plan.Description = req.Description[i]

		// деление строки на отдельные значения функций
		planFeatures := strings.Split(req.Features[i], ",")

		// удаление пробелов по краям каждого значения
		for j, feature := range planFeatures {
			planFeatures[j] = strings.TrimSpace(feature)
		}

		plan.Features = planFeatures

		if err := data.UpdatePlan(plan); err != nil {
			return err
		}
		logger.Log("msg", "plan changed", "planID ", planID)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return c.Redirect("/admin/plans")
}

func HandlePlansCreate(c *fiber.Ctx) error {
	planName := c.FormValue("name")
	planDescription := c.FormValue("description")
	planFeaturesString := c.FormValue("features")

	planFeatures := strings.Split(planFeaturesString, ",")

	for i, feature := range planFeatures {
		planFeatures[i] = strings.TrimSpace(feature)
	}

	plan := &data.Plans{
		Name:        planName,
		Description: planDescription,
		Features:    planFeatures,
	}

	_, err := db.Bun.NewInsert().Model(plan).Exec(context.Background())
	if err != nil {
		return err
	}

	return c.Redirect("/admin/plans")
}
