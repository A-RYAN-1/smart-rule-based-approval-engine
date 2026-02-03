package repositories

import (
	"context"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	aggQueryFetchAllLeaves = `
		SELECT id, leave_type, from_date, to_date, status::TEXT, reason, approval_comment, created_at
		FROM leave_requests WHERE employee_id = $1
	`
	aggQueryFetchAllExpenses = `
		SELECT id, amount, category, status::TEXT, reason, approval_comment, created_at
		FROM expense_requests WHERE employee_id = $1
	`
	aggQueryFetchAllDiscounts = `
		SELECT id, discount_percentage, status::TEXT, reason, approval_comment, created_at
		FROM discount_requests WHERE employee_id = $1
	`
)

type AggregatedRepository interface {
	GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error)
}

type aggregatedRepository struct {
	db *pgxpool.Pool
}

func NewAggregatedRepository(db *pgxpool.Pool) AggregatedRepository {
	return &aggregatedRepository{db: db}
}

type aggCombinedReq struct {
	data      map[string]interface{}
	createdAt time.Time
	reqType   string
}

func (r *aggregatedRepository) GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error) {
	// Initialize slices
	leaves = []map[string]interface{}{}
	expenses = []map[string]interface{}{}
	discounts = []map[string]interface{}{}

	// 1. Fetch all requests for the user from each table
	lReqs, err := r.fetchAllLeaves(ctx, userID)
	if err != nil {
		return
	}
	eReqs, err := r.fetchAllExpenses(ctx, userID)
	if err != nil {
		return
	}
	dReqs, err := r.fetchAllDiscounts(ctx, userID)
	if err != nil {
		return
	}

	// 2. Combine all into a single list
	var combined []aggCombinedReq
	combined = append(combined, lReqs...)
	combined = append(combined, eReqs...)
	combined = append(combined, dReqs...)

	// 3. Set total count
	total = len(combined)

	// 4. Global Sort by CreatedAt DESC
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].createdAt.After(combined[j].createdAt)
	})

	// 5. Apply pagination (limit, offset)
	if offset >= len(combined) {
		return // Return empty buckets and total
	}
	end := offset + limit
	if end > len(combined) {
		end = len(combined)
	}
	paginated := combined[offset:end]

	// 6. Redistribute paginated results back to buckets
	for _, item := range paginated {
		switch item.reqType {
		case "LEAVE":
			leaves = append(leaves, item.data)
		case "EXPENSE":
			expenses = append(expenses, item.data)
		case "DISCOUNT":
			discounts = append(discounts, item.data)
		}
	}

	return
}

func (r *aggregatedRepository) fetchAllLeaves(ctx context.Context, userID int64) ([]aggCombinedReq, error) {
	rows, err := r.db.Query(ctx, aggQueryFetchAllLeaves, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []aggCombinedReq
	for rows.Next() {
		var (
			id       int64
			lType    string
			from, to time.Time
			status   string
			reason   string
			comment  *string
			created  time.Time
		)
		if err := rows.Scan(&id, &lType, &from, &to, &status, &reason, &comment, &created); err != nil {
			return nil, err
		}
		res = append(res, aggCombinedReq{
			reqType:   "LEAVE",
			createdAt: created,
			data: map[string]interface{}{
				"id":               id,
				"leave_type":       lType,
				"from_date":        from.Format("2006-01-02"),
				"to_date":          to.Format("2006-01-02"),
				"status":           status,
				"reason":           reason,
				"approval_comment": comment,
				"created_at":       created.Format(time.RFC3339),
			},
		})
	}
	return res, nil
}

func (r *aggregatedRepository) fetchAllExpenses(ctx context.Context, userID int64) ([]aggCombinedReq, error) {
	rows, err := r.db.Query(ctx, aggQueryFetchAllExpenses, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []aggCombinedReq
	for rows.Next() {
		var (
			id      int64
			amount  float64
			cat     string
			status  string
			reason  string
			comment *string
			created time.Time
		)
		if err := rows.Scan(&id, &amount, &cat, &status, &reason, &comment, &created); err != nil {
			return nil, err
		}
		res = append(res, aggCombinedReq{
			reqType:   "EXPENSE",
			createdAt: created,
			data: map[string]interface{}{
				"id":               id,
				"amount":           amount,
				"category":         cat,
				"status":           status,
				"reason":           reason,
				"approval_comment": comment,
				"created_at":       created.Format(time.RFC3339),
			},
		})
	}
	return res, nil
}

func (r *aggregatedRepository) fetchAllDiscounts(ctx context.Context, userID int64) ([]aggCombinedReq, error) {
	rows, err := r.db.Query(ctx, aggQueryFetchAllDiscounts, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []aggCombinedReq
	for rows.Next() {
		var (
			id      int64
			percent float64
			status  string
			reason  string
			comment *string
			created time.Time
		)
		if err := rows.Scan(&id, &percent, &status, &reason, &comment, &created); err != nil {
			return nil, err
		}
		res = append(res, aggCombinedReq{
			reqType:   "DISCOUNT",
			createdAt: created,
			data: map[string]interface{}{
				"id":                  id,
				"discount_percentage": percent,
				"status":              status,
				"reason":              reason,
				"approval_comment":    comment,
				"created_at":          created.Format(time.RFC3339),
			},
		})
	}
	return res, nil
}
