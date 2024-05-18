package data

import (
	"context"
	"time"

	"github.com/Jhnvlglmlbrt/monitoring-certs/db"
	"github.com/Jhnvlglmlbrt/monitoring-certs/logger"
	"github.com/Jhnvlglmlbrt/monitoring-certs/util"

	"github.com/gofiber/fiber/v2"
	"github.com/uptrace/bun"
)

const (
	domainTrackingTable = "domain_trackings"
	DefaultLimit        = 25
)

type DomainTrackingInfo struct {
	ServerIP      string
	Issuer        string
	SignatureAlgo string
	PublicKeyAlgo string
	EncodedPEM    string
	PublicKey     string
	Signature     string
	DNSNames      string
	KeyUsage      string
	ExtKeyUsages  []string `bun:",array"`
	Expires       time.Time
	Status        string
	LastPollAt    time.Time
	Latency       int
	Error         string
}

type DomainTracking struct {
	ID         int64  `bun:"id,pk,autoincrement"`
	UserID     string `bun:"user_id"`
	DomainName string

	DomainTrackingInfo
}

type Users struct {
	ID                string    `bun:"id,pk,autoincrement"`
	Aud               string    `bun:"aud"`
	Email             string    `bun:"email"`
	EncryptedPassword string    `bun:"encrypted_password"`
	EmailConfirmedAt  time.Time `bun:"email_confirmed_at"`
	CreatedAt         time.Time `bun:"created_at"`
	UpdatedAt         time.Time `bun:"updated_at"`
}

func CountUserDomainTrackings(userID string) (int, error) {
	return db.Bun.NewSelect().
		Model(&DomainTracking{}).
		Where("user_id = ?", userID).
		Count(context.Background())
}

func CountUserFavorites(userID string) (int, error) {
	return db.Bun.NewSelect().
		Model(&Favorites{}).
		Where("user_id = ?", userID).
		Count(context.Background())
}

func CountDomainTrackings() (int, error) {
	return db.Bun.NewSelect().
		Model(&DomainTracking{}).
		Count(context.Background())
}

// func CountUsers(aud string) (int, error) {
// 	return db.Bun.NewSelect().
// 		Model(&Users{}).
// 		Where("aud = ?", aud).
// 		Count(context.Background())
// }

func CountUsers(aud string) (int, error) {
	return db.Bun.NewSelect().
		Table("auth.users").
		Where("aud = ?", aud).
		Count(context.Background()) // Use Value to get the count
}

func GetDomainTrackings(filter fiber.Map, limit int, page int, sortField string, ascending bool) ([]DomainTracking, error) {
	if limit == 0 {
		limit = DefaultLimit
	}
	var trackings []DomainTracking
	builder := db.Bun.NewSelect().Model(&trackings).Limit(limit)
	for k, v := range filter {
		if v != "" {
			builder.Where("? = ?", bun.Ident(k), v)
		}
	}
	offset := (page - 1) * limit
	builder.Offset(offset)
	if ascending {
		builder.OrderExpr("? ASC", bun.Ident(sortField))
	} else {
		builder.OrderExpr("? DESC", bun.Ident(sortField))
	}
	err := builder.Scan(context.Background())
	return trackings, err
}

func GetAdminDomainTrackings(filter fiber.Map, limit int, page int, sortField string, ascending bool) ([]DomainTracking, error) {
	if limit == 0 {
		limit = DefaultLimit
	}
	var trackings []DomainTracking
	builder := db.Bun.NewSelect().Model(&trackings).Limit(limit)
	for k, v := range filter {
		if v != "" {
			builder.Where("? = ?", bun.Ident(k), v)
		}
	}
	offset := (page - 1) * limit
	builder.Offset(offset)
	if ascending {
		builder.OrderExpr("? ASC", bun.Ident(sortField))
	} else {
		builder.OrderExpr("? DESC", bun.Ident(sortField))
	}
	err := builder.Scan(context.Background())
	return trackings, err
}

func GetDomainTracking(query fiber.Map) (*DomainTracking, error) {
	var (
		tracking = new(DomainTracking)
		ctx      = context.Background()
	)
	builder := db.Bun.NewSelect().Model(tracking).QueryBuilder()
	builder = db.WhereMap(builder, query)
	err := builder.Unwrap().(*bun.SelectQuery).Limit(1).Scan(ctx)
	return tracking, err
}

func DeleteDomainTracking(query fiber.Map) error {
	builder := db.Bun.NewDelete().Model(&DomainTracking{}).QueryBuilder()
	builder = db.WhereMap(builder, query)
	_, err := builder.Unwrap().(*bun.DeleteQuery).Exec(context.Background())
	return err
}

func DeleteFavorite(query fiber.Map) error {
	builder := db.Bun.NewDelete().Model(&Favorites{}).QueryBuilder()
	builder = db.WhereMap(builder, query)
	_, err := builder.Unwrap().(*bun.DeleteQuery).Exec(context.Background())
	return err
}

func InsertDomainTracking(tracking *DomainTracking) error {
	_, err := db.Bun.NewInsert().Model(tracking).Exec(context.Background())
	return err
}

func CreateDomainTrackings(trackings []*DomainTracking) (err error) {
	tx, err := db.Bun.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			logger.Log("error", "rollback transaction", "query", "createDomainTrackings", "err", err)
		}
	}()

	for _, tracking := range trackings {
		// Check if already exist. If so, skip.
		query := fiber.Map{
			"domain_name": tracking.DomainName,
			"user_id":     tracking.UserID,
		}
		_, err = GetDomainTracking(query)
		if err != nil {
			if util.IsErrNoRecords(err) {
				if err := InsertDomainTracking(tracking); err != nil {
					return err
				}
			} else {
				logger.Log("error", err)
			}
		}
	}
	return tx.Commit()
}

type TrackingAndAccount struct {
	Account
	DomainTracking
}

func GetAllTrackingsWithAccount() ([]TrackingAndAccount, error) {
	var (
		trackings []TrackingAndAccount
		ctx       = context.Background()
	)
	err := db.Bun.NewSelect().
		ColumnExpr("a.*").
		ColumnExpr("dt.*").
		TableExpr("domain_trackings as dt").
		Join("INNER JOIN accounts AS a").
		JoinOn("a.user_id = dt.user_id").
		Scan(ctx, &trackings)
	return trackings, err
}

func UpdateAllTrackings(trackings []DomainTracking) error {
	_, err := db.Bun.NewUpdate().
		Model(&trackings).
		Column(
			"issuer",
			"expires",
			"signature_algo",
			"public_key_algo",
			"dns_names",
			"last_poll_at",
			"latency",
			"error",
			"status",
			"signature",
			"public_key",
			"key_usage",
			"ext_key_usages",
			"encoded_pem",
			"server_ip",
		).
		Bulk().
		Exec(context.Background())
	return err
}

type Favorites struct {
	ID        int       `bun:"id,pk,autoincrement"`
	UserID    string    `bun:"user_id"`
	DomainID  int64     `bun:"domain_id"`
	CreatedAt time.Time `bun:"created_at,default:current_timestamp"`
}

func InsertFavorite(favorite *Favorites) error {
	_, err := db.Bun.NewInsert().Model(favorite).Exec(context.Background())
	return err
}

func GetFavoriteByDomainID(userID string, domainID int64) (*Favorites, error) {
	favorite := new(Favorites)
	err := db.Bun.NewSelect().Model(favorite).
		Where("user_id = ?", userID).
		Where("domain_id = ?", domainID).
		Limit(1).
		Scan(context.Background())

	if err != nil {
		return nil, err
	}

	return favorite, nil
}

func GetDomainNameByID(domainID int64) (string, error) {
	var domainName string
	err := db.Bun.NewSelect().Model(&DomainTracking{}).ColumnExpr("domain_name").Where("id = ?", domainID).Scan(context.Background(), &domainName)
	if err != nil {
		return "", err
	}
	return domainName, nil
}

// func GetFavoriteDomains(userID string) ([]int64, error) {
// 	var domainIDs []int64
// 	err := db.Bun.NewSelect().Model(&Favorites{}).ColumnExpr("domain_id").Where("user_id = ?", userID).Scan(context.Background(), &domainIDs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return domainIDs, nil
// }

func GetFavoriteDomainTrackings(userID string, filter fiber.Map, limit int, page int, sortField string, ascending bool) ([]DomainTracking, error) {
	if limit == 0 {
		limit = DefaultLimit
	}
	var domainTrackings []DomainTracking

	subquery := db.Bun.NewSelect().
		Column("domain_id").
		Table("favorites").
		Where("user_id = ?", userID)

	builder := db.Bun.NewSelect().
		Model(&domainTrackings).
		ColumnExpr("dt.*").
		TableExpr("domain_trackings as dt").
		Join("INNER JOIN (?) AS sq ON sq.domain_id = dt.id", subquery).
		Limit(limit).Distinct()

	for k, v := range filter {
		if v != "" && k != "user_id" {
			builder.Where("? = ?", bun.Ident("dt."+k), v)
		}
	}

	offset := (page - 1) * limit
	builder.Offset(offset)

	if ascending {
		builder.OrderExpr("dt.id ASC") // Changed from "f.id"
	} else {
		builder.OrderExpr("dt.id DESC") // Changed from "f.id"
	}

	err := builder.Scan(context.Background(), &domainTrackings)
	return domainTrackings, err
}

func RemoveFavoriteDomain(domainID int64) error {
	_, err := db.Bun.NewDelete().
		Model(&Favorites{}).
		Where("domain_id = ?", domainID).
		Exec(context.Background())
	return err
}

func RemoveDomain(userID string, domainID int64) error {
	_, err := db.Bun.NewDelete().
		Model(&DomainTracking{}).
		Where("user_id = ?", userID).
		Where("id = ?", domainID).
		Exec(context.Background())
	return err
}

// func GetUsers(filter fiber.Map, limit int, page int, sortField string, ascending bool) ([]Users, error) {
// 	if limit == 0 {
// 		limit = DefaultLimit
// 	}

// 	var users []Users

// 	builder := db.Bun.NewSelect().Model(&users).Limit(limit)
// 	for k, v := range filter {
// 		if v != "" {
// 			builder.Where("? = ?", bun.Ident(k), v)
// 		}
// 	}
// 	offset := (page - 1) * limit
// 	builder.Offset(offset)
// 	if ascending {
// 		builder.OrderExpr("? ASC", bun.Ident(sortField))
// 	} else {
// 		builder.OrderExpr("? DESC", bun.Ident(sortField))
// 	}
// 	err := builder.Scan(context.Background())
// 	return users, err
// }

func GetUsers(filter fiber.Map, limit int, page int, sortField string, ascending bool) ([]Users, error) {
	if limit == 0 {
		limit = DefaultLimit
	}

	var users []Users
	builder := db.Bun.NewSelect().
		Table("auth.users").
		Column("id", "aud", "email", "encrypted_password", "email_confirmed_at", "created_at", "updated_at").
		Limit(limit)

	for k, v := range filter {
		if v != "" {
			builder.Where("auth.users.? = ?", bun.Ident(k), v)
		}
	}

	offset := (page - 1) * limit
	builder.Offset(offset)

	if ascending {
		builder.OrderExpr("auth.users.? ASC", bun.Ident(sortField))
	} else {
		builder.OrderExpr("auth.users.? DESC", bun.Ident(sortField))
	}

	err := builder.Scan(context.Background(), &users)

	// spew.Dump(users)
	return users, err
}

// func RemoveUser(userID string) error {
// 	_, err := db.Bun.NewDelete().
// 		Model(&Users{}).
// 		Where("user_id = ?", userID).
// 		Exec(context.Background())
// 	return err
// }

func RemoveAccount(userID string) error {
	_, err := db.Bun.NewDelete().
		Model(&Account{}).
		Where("user_id = ?", userID).
		Exec(context.Background())
	return err
}

// func GetUserByID(userID string) (*Users, error) {
// 	user := new(Users)
// 	err := db.Bun.NewSelect().
// 		Model(user).
// 		Where("id = ?", userID).
// 		Scan(context.Background())
// 	if err != nil {
// 		return nil, err
// 	}
// 	return user, nil
// }

func GetAccountByID(userID string) (*Account, error) {
	account := new(Account)
	err := db.Bun.NewSelect().
		Model(account).
		Where("id = ?", userID).
		Scan(context.Background())
	if err != nil {
		return nil, err
	}
	return account, nil
}
