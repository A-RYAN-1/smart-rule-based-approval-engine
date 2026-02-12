package repositories

import (
	"context"
	"sort"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
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
	aggQueryFetchPendingLeavesForManager = `
		SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at
		FROM leave_requests lr JOIN users u ON lr.employee_id = u.id
		WHERE lr.status = 'PENDING' AND u.manager_id = $1
	`
	aggQueryFetchPendingLeavesForAdmin = `
		SELECT lr.id, u.name, lr.from_date, lr.to_date, lr.leave_type, lr.reason, lr.created_at
		FROM leave_requests lr JOIN users u ON lr.employee_id = u.id
		WHERE lr.status = 'PENDING'
	`
	aggQueryFetchPendingExpensesForManager = `
		SELECT er.id, u.name, er.amount, er.category, er.reason, er.created_at
		FROM expense_requests er JOIN users u ON er.employee_id = u.id
		WHERE er.status = 'PENDING' AND u.manager_id = $1
	`
	aggQueryFetchPendingExpensesForAdmin = `
		SELECT er.id, u.name, er.amount, er.category, er.reason, er.created_at
		FROM expense_requests er JOIN users u ON er.employee_id = u.id
		WHERE er.status = 'PENDING'
	`
	aggQueryFetchPendingDiscountsForManager = `
		SELECT dr.id, u.name, dr.discount_percentage, dr.reason, dr.created_at
		FROM discount_requests dr JOIN users u ON dr.employee_id = u.id
		WHERE dr.status = 'PENDING' AND u.manager_id = $1
	`
	aggQueryFetchPendingDiscountsForAdmin = `
		SELECT dr.id, u.name, dr.discount_percentage, dr.reason, dr.created_at
		FROM discount_requests dr JOIN users u ON dr.employee_id = u.id
		WHERE dr.status = 'PENDING'
	`
)

type aggregatedRepository struct {
	db interfaces.DB
}

func NewAggregatedRepository(ctx context.Context, db interfaces.DB) interfaces.MyRequestsRepository {
	return &aggregatedRepository{db: db}
}

type aggCombinedReq struct {
	data      map[string]interface{}
	createdAt time.Time
	reqType   string
}

func (r *aggregatedRepository) GetPendingAllRequests(ctx context.Context, role string, approverID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error) {
	leaves = []map[string]interface{}{}
	expenses = []map[string]interface{}{}
	discounts = []map[string]interface{}{}

	var lReqs, eReqs, dReqs []aggCombinedReq

	if role == "admin" {
		lReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingLeavesForAdmin, "LEAVE", 0)
		if err != nil {
			return
		}
		eReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingExpensesForAdmin, "EXPENSE", 0)
		if err != nil {
			return
		}
		dReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingDiscountsForAdmin, "DISCOUNT", 0)
		if err != nil {
			return
		}
	} else {
		lReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingLeavesForManager, "LEAVE", approverID)
		if err != nil {
			return
		}
		eReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingExpensesForManager, "EXPENSE", approverID)
		if err != nil {
			return
		}
		dReqs, err = r.fetchPendingRequests(ctx, aggQueryFetchPendingDiscountsForManager, "DISCOUNT", approverID)
		if err != nil {
			return
		}
	}

	var combined []aggCombinedReq
	combined = append(combined, lReqs...)
	combined = append(combined, eReqs...)
	combined = append(combined, dReqs...)

	total = len(combined)

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].createdAt.After(combined[j].createdAt)
	})

	if offset >= len(combined) {
		return
	}
	end := offset + limit
	if end > len(combined) {
		end = len(combined)
	}
	paginated := combined[offset:end]

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

func (r *aggregatedRepository) fetchPendingRequests(ctx context.Context, query, reqType string, approverID int64) ([]aggCombinedReq, error) {
	var rows interfaces.Rows
	var err error
	if approverID > 0 {
		rows, err = r.db.Query(ctx, query, approverID)
	} else {
		rows, err = r.db.Query(ctx, query)
	}
	if err != nil {
		return nil, utils.MapPgError(err)
	}
	defer rows.Close()

	var res []aggCombinedReq
	for rows.Next() {
		switch reqType {
		case "LEAVE":
			var (
				id       int64
				name     string
				from, to time.Time
				lType    string
				reason   string
				created  time.Time
			)
			if err := rows.Scan(&id, &name, &from, &to, &lType, &reason, &created); err != nil {
				return nil, utils.MapPgError(err)
			}
			res = append(res, aggCombinedReq{
				reqType:   "LEAVE",
				createdAt: created,
				data: map[string]interface{}{
					"id":         id,
					"employee":   name,
					"from_date":  from.Format("2006-01-02"),
					"to_date":    to.Format("2006-01-02"),
					"leave_type": lType,
					"reason":     reason,
					"created_at": created.Format(time.RFC3339),
				},
			})
		case "EXPENSE":
			var (
				id      int64
				name    string
				amount  float64
				cat     string
				reason  *string
				created time.Time
			)
			if err := rows.Scan(&id, &name, &amount, &cat, &reason, &created); err != nil {
				return nil, utils.MapPgError(err)
			}
			res = append(res, aggCombinedReq{
				reqType:   "EXPENSE",
				createdAt: created,
				data: map[string]interface{}{
					"id":         id,
					"employee":   name,
					"amount":     amount,
					"category":   cat,
					"reason":     reason,
					"created_at": created.Format(time.RFC3339),
				},
			})
		case "DISCOUNT":
			var (
				id      int64
				name    string
				percent float64
				reason  string
				created interface{}
			)
			if err := rows.Scan(&id, &name, &percent, &reason, &created); err != nil {
				return nil, utils.MapPgError(err)
			}

			var createdAt time.Time
			switch v := created.(type) {
			case time.Time:
				createdAt = v
			case *time.Time:
				if v != nil {
					createdAt = *v
				}
			}

			res = append(res, aggCombinedReq{
				reqType:   "DISCOUNT",
				createdAt: createdAt,
				data: map[string]interface{}{
					"id":                  id,
					"employee":            name,
					"discount_percentage": percent,
					"reason":              reason,
					"created_at":          createdAt.Format(time.RFC3339),
				},
			})
		}
	}
	return res, nil
}

func (r *aggregatedRepository) fetchAllLeaves(ctx context.Context, userID int64) ([]aggCombinedReq, error) {
	rows, err := r.db.Query(ctx, aggQueryFetchAllLeaves, userID)
	if err != nil {
		return nil, utils.MapPgError(err)
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
			return nil, utils.MapPgError(err)
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
		return nil, utils.MapPgError(err)
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
			return nil, utils.MapPgError(err)
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
		return nil, utils.MapPgError(err)
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
			return nil, utils.MapPgError(err)
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

// These methods satisfy the interfaces.MyRequestsRepository interface
func (r *aggregatedRepository) GetMyLeaveRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	res, err := r.fetchAllLeaves(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	total := len(res)
	// Sort by created at desc
	sort.Slice(res, func(i, j int) bool {
		return res[i].createdAt.After(res[j].createdAt)
	})

	if offset >= len(res) {
		return []map[string]interface{}{}, total, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	res = res[offset:end]

	data := make([]map[string]interface{}, len(res))
	for i, v := range res {
		data[i] = v.data
	}
	return data, total, nil
}

func (r *aggregatedRepository) GetMyExpenseRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	res, err := r.fetchAllExpenses(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	total := len(res)
	// Sort by created at desc
	sort.Slice(res, func(i, j int) bool {
		return res[i].createdAt.After(res[j].createdAt)
	})

	if offset >= len(res) {
		return []map[string]interface{}{}, total, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	res = res[offset:end]

	data := make([]map[string]interface{}, len(res))
	for i, v := range res {
		data[i] = v.data
	}
	return data, total, nil
}

func (r *aggregatedRepository) GetMyDiscountRequests(ctx context.Context, userID int64, limit, offset int) ([]map[string]interface{}, int, error) {
	res, err := r.fetchAllDiscounts(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	total := len(res)
	// Sort by created at desc
	sort.Slice(res, func(i, j int) bool {
		return res[i].createdAt.After(res[j].createdAt)
	})

	if offset >= len(res) {
		return []map[string]interface{}{}, total, nil
	}
	end := offset + limit
	if end > len(res) {
		end = len(res)
	}
	res = res[offset:end]

	data := make([]map[string]interface{}, len(res))
	for i, v := range res {
		data[i] = v.data
	}
	return data, total, nil
}
func (r *aggregatedRepository) GetMyAllRequests(ctx context.Context, userID int64, limit, offset int) (leaves []map[string]interface{}, expenses []map[string]interface{}, discounts []map[string]interface{}, total int, err error) {
	leaves = []map[string]interface{}{}
	expenses = []map[string]interface{}{}
	discounts = []map[string]interface{}{}

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

	var combined []aggCombinedReq
	combined = append(combined, lReqs...)
	combined = append(combined, eReqs...)
	combined = append(combined, dReqs...)

	total = len(combined)

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].createdAt.After(combined[j].createdAt)
	})

	if offset >= len(combined) {
		return
	}
	end := offset + limit
	if end > len(combined) {
		end = len(combined)
	}
	paginated := combined[offset:end]

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
