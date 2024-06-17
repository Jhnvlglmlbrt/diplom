package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/data"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/Jhnvlglmlbrt/monitoring-certs/pkg/notify"
	"github.com/Jhnvlglmlbrt/monitoring-certs/pkg/ssl"
	"github.com/Jhnvlglmlbrt/monitoring-certs/settings"
	"github.com/Jhnvlglmlbrt/monitoring-certs/util"
	"github.com/gofiber/fiber/v2"
	"github.com/sujit-baniya/flash"
)

var limitFilters = []int{
	5,
	10,
	25,
	50,
	100,
}

var statusFilters = []string{
	"all",
	data.StatusHealthy,
	data.StatusExpires,
	data.StatusExpired,
	data.StatusInvalid,
	data.StatusOffline,
	data.StatusUnresponsive,
}

func HandleDomainList(c *fiber.Ctx) error {
	var (
		domainTrackings []data.DomainTracking
	)

	user := getAuthenticatedUser(c)
	count, err := data.CountUserDomainTrackings(user.ID)
	if err != nil {
		return err
	}
	if count == 0 {
		return c.Render("domains/index", fiber.Map{"userHasTrackings": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}

	filterContext := buildFilterContext(filter)
	query := fiber.Map{
		"user_id": user.ID,
	}
	if filter.Status != "all" {
		query["status"] = filter.Status
	}

	// чтоьы показывать кол-во доменов по плану подписки
	account, err := data.GetUserAccount(user.ID)
	if err != nil {
		return err
	}

	maxTrackings := settings.Account[account.PlanID].MaxTrackings

	if count > maxTrackings {
		filter.Limit = maxTrackings
		count = maxTrackings
	}

	domainTrackings, err = data.GetDomainTrackings(query, filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	// вывод доменов, если есть поиск
	if searchQuery := c.Query("q"); searchQuery != "" {
		searchQuery = strings.TrimSpace(searchQuery)
		query["domain_name"] = searchQuery
		domainTrackings, err = data.GetDomainTrackings(query, filter.Limit, filter.Page, "id", true)
		if err != nil {
			return err
		}
	}

	data := fiber.Map{
		"trackings":        domainTrackings,
		"filters":          filterContext,
		"userHasTrackings": true,
		"pages":            buildPages(count, filter.Limit),
		"queryParams":      filter.encode(),
	}
	return c.Render("domains/index", data)
}

func HandleDomainNew(c *fiber.Ctx) error {
	return c.Render("domains/new", fiber.Map{})
}

func HandleDomainsDelete(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)

	var req struct {
		DomainIDs []string `json:"domain_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	for _, domainID := range req.DomainIDs {
		domainIDInt, err := strconv.ParseInt(domainID, 10, 64)
		if err != nil {
			return err
		}

		err = data.RemoveDomain(user.ID, domainIDInt)
		if err != nil {
			continue
		}
		logger.Log("msg", "domain deleted", domainIDInt)

	}

	return c.Redirect("/domains")
}

func HandleAdminDomainsDelete(c *fiber.Ctx) error {
	var req struct {
		DomainIDs []string `json:"domain_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	for _, domainID := range req.DomainIDs {

		query := fiber.Map{
			"id": domainID,
		}

		domainIDInt, err := strconv.ParseInt(domainID, 10, 64)
		if err != nil {
			return err
		}

		tracking, err := data.GetDomainTracking(query)
		if err != nil {
			return err
		}

		err = data.RemoveDomain(tracking.UserID, domainIDInt)
		if err != nil {
			continue
		}
		logger.Log("msg", "domain deleted", domainIDInt)

	}

	return c.Redirect("/domains")
}

func HandleDomainDelete(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)
	domainID := c.Params("id")

	query := fiber.Map{
		"user_id": user.ID,
		"id":      c.Params("id"),
	}

	if err := data.DeleteFavorite(fiber.Map{"user_id": user.ID, "domain_id": domainID}); err != nil {
		return err
	}

	if err := data.DeleteDomainTracking(query); err != nil {
		return err
	}
	return c.Redirect("/domains")
}

func HandleAdminDomainDelete(c *fiber.Ctx) error {
	domainID := c.Params("id")

	query := fiber.Map{
		"id": c.Params("id"),
	}

	tracking, err := data.GetDomainTracking(query)
	if err != nil {
		return err
	}

	if err := data.DeleteFavorite(fiber.Map{"user_id": tracking.UserID, "domain_id": domainID}); err != nil {
		return err
	}

	if err := data.DeleteDomainTracking(query); err != nil {
		return err
	}
	return c.Redirect("/domains")
}

func HandleDomainShowRaw(c *fiber.Ctx) error {
	trackingID := c.Params("id")
	user := getAuthenticatedUser(c)
	query := fiber.Map{
		"user_id": user.ID,
		"id":      trackingID,
	}
	tracking, err := data.GetDomainTracking(query)
	if err != nil {
		return err
	}
	return c.Send([]byte(tracking.EncodedPEM))
}

func HandleAdminDomainShowRaw(c *fiber.Ctx) error {
	trackingID := c.Params("id")
	query := fiber.Map{
		"id": trackingID,
	}
	tracking, err := data.GetDomainTracking(query)
	if err != nil {
		return err
	}
	return c.Send([]byte(tracking.EncodedPEM))
}

func HandleDomainShow(c *fiber.Ctx) error {
	trackingID := c.Params("id")
	user := getAuthenticatedUser(c)
	query := fiber.Map{
		"user_id": user.ID,
		"id":      trackingID,
	}
	tracking, err := data.GetDomainTracking(query)
	if err != nil {
		return err
	}
	context := fiber.Map{
		"tracking": tracking,
	}
	return c.Render("admin/show", context)
}

func HandleAdminDomainShow(c *fiber.Ctx) error {
	trackingID := c.Params("id")
	query := fiber.Map{
		"id": trackingID,
	}
	tracking, err := data.GetDomainTracking(query)
	if err != nil {
		return err
	}
	user, err := GetEmailForUserID(tracking.UserID)
	if err != nil {
		return err
	}

	context := fiber.Map{
		"userEmail": user,
		"tracking":  tracking,
	}
	return c.Render("admin/show", context)
}

// Обработчик для отправки ручного уведомления
func HandleSendTestNotification(c *fiber.Ctx) error {
	trackingID := c.Params("id")
	// logger.Log("id", trackingID)

	tracking, err := data.GetDomainTracking(fiber.Map{"id": trackingID})
	if err != nil {
		return err
	}

	account, err := data.GetUserAccount(tracking.UserID)
	if err != nil {
		return err
	}

	notifier := notify.NewEmailNotifier(
		os.Getenv("EMAIL_FROM"),
		os.Getenv("EMAIL_PASSWORD"),
		os.Getenv("SMTP_SERVER"),
		os.Getenv("SMTP_PORT"),
	)

	daysUntilExpiration := int(time.Until(tracking.Expires).Hours() / 24)

	subject := `Уведомление о состоянии ваших сертификатов!`
	body := fmt.Sprintf("Ваш домен - %s. Дней до истечения срока действия его SSL сертификата: %d.", tracking.DomainName, daysUntilExpiration)

	if err := notifier.JustNotify(context.Background(), account.NotifyDefaultEmail, subject, body); err != nil {
		logger.Log("error", err)
		return err
	}

	logger.Log("user", tracking.UserID, "email", account.NotifyDefaultEmail, "message", "Уведомление успешно отправлено.")
	return c.Send([]byte("отправлено"))
}

func HandleDomainCreate(c *fiber.Ctx) error {
	flashData := fiber.Map{}
	userDomainsInput := c.FormValue("domains")
	userDomainsInput = strings.ReplaceAll(userDomainsInput, " ", "")

	if len(userDomainsInput) == 0 {
		flashData["domainsError"] = "Please provide at least 1 valid domain name"
		return flash.WithData(c, flashData).Redirect("/domains/new")
	}
	domains := strings.Split(userDomainsInput, ",")
	if len(domains) == 0 {
		flashData["domainsError"] = "Invalid domain list input. Make sure to use a comma seperated list (domain1.com, domain2.com, ..)"
		flashData["domains"] = userDomainsInput
		return flash.WithData(c, flashData).Redirect("/domains/new")
	}
	for _, domain := range domains {
		if !util.IsValidDomainName(domain) {
			flashData["domainsError"] = fmt.Sprintf("%s is not a valid domain", domain)
			flashData["domains"] = userDomainsInput
			return flash.WithData(c, flashData).Redirect("/domains/new")
		}
	}

	user := getAuthenticatedUser(c)
	account, err := data.GetUserAccount(user.ID)
	if err != nil {
		return err
	}

	maxTrackings := settings.Account[account.PlanID].MaxTrackings
	count, err := data.CountUserDomainTrackings(user.ID)
	if err != nil {
		return err
	}
	if account.PlanID > data.PlanStarter && account.SubscriptionStatus != "active" {
		logger.Log("error", "subscription status not active", "status", account.SubscriptionStatus)
		return flash.WithData(c, flashData).Redirect("/domains/new")
	}
	if len(domains)+count > maxTrackings {
		flashData["maxTrackings"] = maxTrackings
		flashData["domains"] = userDomainsInput
		return flash.WithData(c, flashData).Redirect("/domains/new")
	}
	if err := createAllDomainTrackings(user.ID, domains); err != nil {
		return err
	}
	return c.Redirect("/domains")
}

func createAllDomainTrackings(userID string, domains []string) error {
	var (
		trackings = []*data.DomainTracking{}
		wg        = sync.WaitGroup{}
	)
	for _, domain := range domains {
		wg.Add(1)
		go func(domain string) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer func() {
				cancel()
				wg.Done()
			}()
			trackingInfo, err := ssl.PollDomain(ctx, domain)
			if err != nil {
				logger.Log("error", "polling domain failed", "err", err, "domain", domain)
				return
			}
			tracking := &data.DomainTracking{
				DomainName:         domain,
				UserID:             userID,
				DomainTrackingInfo: *trackingInfo,
			}
			trackings = append(trackings, tracking)
		}(domain)
	}
	wg.Wait()

	fmt.Println("inserting domains into the database", len(trackings))

	return data.CreateDomainTrackings(trackings)
}

type TrackingFilter struct {
	Limit  int
	Page   int
	Status string
}

func (f *TrackingFilter) encode() string {
	values := url.Values{}
	if f.Limit != 0 {
		values.Set("limit", strconv.Itoa(f.Limit))
	}
	if f.Page != 0 {
		values.Set("page", strconv.Itoa(f.Page))
	}
	values.Set("status", f.Status)
	return values.Encode()
}

func buildTrackingFilter(c *fiber.Ctx) (*TrackingFilter, error) {
	filter := new(TrackingFilter)
	if err := c.QueryParser(filter); err != nil {
		return nil, err
	}
	if filter.Limit == 0 {
		filter.Limit = data.DefaultLimit
	}
	return filter, nil
}

func buildFilterContext(filter *TrackingFilter) fiber.Map {
	return fiber.Map{
		"statuses":       statusFilters,
		"limits":         limitFilters,
		"selectedStatus": filter.Status,
		"selectedLimit":  filter.Limit,
		"selectedPage":   filter.Page,
	}
}

func buildPages(results int, limit int) []int {
	numPages := int(math.Ceil(float64(results) / float64(limit)))
	// fmt.Println("Total number of pages:", numPages)
	pages := make([]int, numPages)
	for i := range pages {
		pages[i] = i + 1
	}
	// fmt.Println("Generated pages:", pages)
	return pages
}

func HandleFavoritesList(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)
	count, err := data.CountUserFavorites(user.ID)
	if err != nil {
		return err
	}
	if count == 0 {
		return c.Render("domains/favorites", fiber.Map{"userHasTrackings": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}

	filterContext := buildFilterContext(filter)
	query := fiber.Map{
		"user_id": user.ID,
	}
	if filter.Status != "all" {
		query["status"] = filter.Status
	}

	domainTrackings, err := data.GetFavoriteDomainTrackings(user.ID, query, filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	// spew.Dump(domainTrackings)

	if searchQuery := c.Query("q"); searchQuery != "" {
		searchQuery = strings.TrimSpace(searchQuery)
		query := fiber.Map{
			"domain_name": searchQuery,
		}
		domainTrackings, err = data.GetFavoriteDomainTrackings(user.ID, query, filter.Limit, filter.Page, "id", true)
		// spew.Dump(domainTrackings)
		if err != nil {
			return err
		}
	}

	data := fiber.Map{
		"trackings":        domainTrackings,
		"filters":          filterContext,
		"userHasTrackings": true,
		"pages":            buildPages(count, filter.Limit),
		"queryParams":      filter.encode(),
	}
	return c.Render("domains/favorites", data)
}

// HandleAddFavorite обрабатывает запрос на добавление избранных доменов
func HandleAddFavorite(c *fiber.Ctx) error {
	user := getAuthenticatedUser(c)

	var req struct {
		DomainIDs []string `json:"domain_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	for _, domainID := range req.DomainIDs {
		domainID, err := strconv.ParseInt(domainID, 10, 64)
		if err != nil {
			return err
		}

		domain, err := data.GetFavoriteByDomainID(user.ID, domainID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				fmt.Printf("Домен с id - %d не существует.\n", domainID)
			} else {
				return err
			}
		} else {
			if domain != nil {
				domainName, err := data.GetDomainNameByID(domainID)
				if err != nil {
					return err
				}
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Домен (" + domainName + ") уже существует в избранном"})
			}
		}
		favorite := data.Favorites{
			UserID:   user.ID,
			DomainID: domainID,
		}
		if err := data.InsertFavorite(&favorite); err != nil {
			return err
		}

		logger.Log("msg", "Домен добавлен в избранное", "domainID", domainID)
	}

	return c.Redirect("/favorites")
}

func HandleRemoveFavorite(c *fiber.Ctx) error {
	var req struct {
		DomainIDs []string `json:"domain_ids"`
	}
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	for _, domainID := range req.DomainIDs {
		domainIDInt, err := strconv.ParseInt(domainID, 10, 64)
		if err != nil {
			return err
		}

		err = data.RemoveFavoriteDomain(domainIDInt)
		if err != nil {
			continue
		}
		logger.Log("msg", "favorite domain deleted", domainIDInt)

	}

	return c.Redirect("/favorites")
}

func HandleCheckDomainStatus(c *fiber.Ctx) error {
	domain := c.FormValue("domain")

	trackingInfo, err := ssl.PollDomain(context.Background(), domain)
	if err != nil {
		return err
	}

	return c.Render("home/status", fiber.Map{"status": trackingInfo.Status})
}

func HandleAdminDomainList(c *fiber.Ctx) error {
	start := time.Now()
	var (
		domainTrackings []data.DomainTracking
		aud             = "authenticated"
	)

	countUsers, err := data.CountUsers(aud)
	if err != nil {
		return err
	}

	count, err := data.CountDomainTrackings()
	if err != nil {
		return err
	}

	if count*countUsers == 0 {
		return c.Render("admin/domains", fiber.Map{"noTrackings": false})
	}

	filter, err := buildTrackingFilter(c)
	if err != nil {
		return err
	}

	filterContext := buildFilterContext(filter)
	query := fiber.Map{}
	if filter.Status != "all" {
		query["status"] = filter.Status
	}

	domainTrackings, err = data.GetAdminDomainTrackings(query, filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	params := fiber.Map{
		"aud": aud,
	}

	users, err := data.GetUsers(params, filter.Limit, filter.Page, "id", true)
	if err != nil {
		return err
	}

	// вывод доменов, если есть поиск
	if searchQuery := c.Query("q"); searchQuery != "" {
		searchQuery = strings.TrimSpace(searchQuery)
		query["domain_name"] = searchQuery
		domainTrackings, err = data.GetAdminDomainTrackings(query, filter.Limit, filter.Page, "id", true)
		if err != nil {
			return err
		}
	}

	var userIDs []string
	for _, tracking := range domainTrackings {
		userIDs = append(userIDs, tracking.UserID)
	}

	emailMap, err := data.GetEmailsForUserIDs(userIDs)
	if err != nil {
		return err
	}

	data := fiber.Map{
		"emailMap":    emailMap,
		"users":       users,
		"trackings":   domainTrackings,
		"filters":     filterContext,
		"noTrackings": true,
		"pages":       buildPages(count, filter.Limit),
		"queryParams": filter.encode(),
	}
	log.Printf("HandleAdminDomainList took %s", time.Since(start))
	return c.Render("admin/domains", data)
}

func HandleAdminDomainCreate(c *fiber.Ctx) error {
	flashData := fiber.Map{}
	userDomainsInput := c.FormValue("domains")
	userDomainsInput = strings.ReplaceAll(userDomainsInput, " ", "")

	if len(userDomainsInput) == 0 {
		flashData["domainsError"] = "Please provide at least 1 valid domain name"
		return flash.WithData(c, flashData).Redirect("/admin/new")
	}
	domains := strings.Split(userDomainsInput, ",")
	if len(domains) == 0 {
		flashData["domainsError"] = "Invalid domain list input. Make sure to use a comma seperated list (domain1.com, domain2.com, ..)"
		flashData["domains"] = userDomainsInput
		return flash.WithData(c, flashData).Redirect("/admin/new")
	}
	for _, domain := range domains {
		if !util.IsValidDomainName(domain) {
			flashData["domainsError"] = fmt.Sprintf("%s is not a valid domain", domain)
			flashData["domains"] = userDomainsInput
			return flash.WithData(c, flashData).Redirect("/admin/new")
		}
	}

	user := getAuthenticatedUser(c)

	if err := createAllDomainTrackings(user.ID, domains); err != nil {
		return err
	}
	return c.Redirect("/admin/domains")
}

func GetEmailForUserID(userID string) (string, error) {
	user, err := data.GetUser(userID)
	if err != nil {
		return "", err
	}
	return user.Email, nil
}
