package gatewayapi

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	platformconfig "github.com/orcastack/orcastackapi/internal/platform/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type signupRequestRecord struct {
	ID         string     `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Status     string     `json:"status"`
	CreatedAt  string     `json:"created_at"`
	ReviewedAt *string    `json:"reviewed_at,omitempty"`
	ReviewedBy string     `json:"reviewed_by,omitempty"`
	ReviewNote string     `json:"review_note,omitempty"`
}

type signupDecisionRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

var signupStore = struct {
	sync.Mutex
	db *sql.DB
}{ }

func signupDB() (*sql.DB, error) {
	signupStore.Lock()
	defer signupStore.Unlock()

	if signupStore.db != nil {
		return signupStore.db, nil
	}

	dsn := platformconfig.String("DATABASE_URL", "")
	if dsn == "" {
		host := platformconfig.String("ORCASTACK_POSTGRES_HOST", "postgres")
		port := platformconfig.String("ORCASTACK_POSTGRES_PORT", "5432")
		user := platformconfig.String("ORCASTACK_POSTGRES_USER", platformconfig.String("POSTGRES_USER", "orcastack"))
		password := platformconfig.String("ORCASTACK_POSTGRES_PASSWORD", platformconfig.String("POSTGRES_PASSWORD", "orcastack"))
		database := platformconfig.String("ORCASTACK_POSTGRES_DB", platformconfig.String("POSTGRES_DB", "orcastack"))
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, database)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open signup request store: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect signup request store: %w", err)
	}

	signupStore.db = db
	return db, nil
}

func createSignupRequestRecord(ctx context.Context, record signupRequestRecord) error {
	db, err := signupDB()
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx, `
		insert into signup_requests (id, username, email, status, reviewed_by, review_note)
		values ($1, $2, $3, $4, '', '')
	`, record.ID, record.Username, record.Email, record.Status)
	if err != nil {
		return fmt.Errorf("insert signup request: %w", err)
	}

	return nil
}

func listSignupRequestRecords(ctx context.Context) ([]signupRequestRecord, error) {
	db, err := signupDB()
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
		select id, username, email, status, created_at, reviewed_at, coalesce(reviewed_by, ''), coalesce(review_note, '')
		from signup_requests
		order by created_at desc
	`)
	if err != nil {
		return nil, fmt.Errorf("list signup requests: %w", err)
	}
	defer rows.Close()

	requests := make([]signupRequestRecord, 0)
	for rows.Next() {
		var record signupRequestRecord
		var createdAt time.Time
		var reviewedAt sql.NullTime
		if err := rows.Scan(&record.ID, &record.Username, &record.Email, &record.Status, &createdAt, &reviewedAt, &record.ReviewedBy, &record.ReviewNote); err != nil {
			return nil, fmt.Errorf("scan signup request: %w", err)
		}
		record.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		if reviewedAt.Valid {
			formatted := reviewedAt.Time.UTC().Format(time.RFC3339)
			record.ReviewedAt = &formatted
		}
		requests = append(requests, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signup requests: %w", err)
	}

	return requests, nil
}

func reviewSignupRequestRecord(ctx context.Context, requestID, reviewer string, decision signupDecisionRequest) (signupRequestRecord, error) {
	db, err := signupDB()
	if err != nil {
		return signupRequestRecord{}, err
	}

	now := time.Now().UTC()
	row := db.QueryRowContext(ctx, `
		update signup_requests
		set status = $2,
		    reviewed_at = $3,
		    reviewed_by = $4,
		    review_note = $5
		where id = $1
		returning id, username, email, status, created_at, reviewed_at, reviewed_by, review_note
	`, requestID, decision.Status, now, reviewer, decision.Note)

	var record signupRequestRecord
	var createdAt time.Time
	var reviewedAt sql.NullTime
	if err := row.Scan(&record.ID, &record.Username, &record.Email, &record.Status, &createdAt, &reviewedAt, &record.ReviewedBy, &record.ReviewNote); err != nil {
		if err == sql.ErrNoRows {
			return signupRequestRecord{}, fmt.Errorf("signup request not found")
		}
		return signupRequestRecord{}, fmt.Errorf("review signup request: %w", err)
	}

	record.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	if reviewedAt.Valid {
		formatted := reviewedAt.Time.UTC().Format(time.RFC3339)
		record.ReviewedAt = &formatted
	}

	return record, nil
}
