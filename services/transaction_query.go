package services

type TransactionQuery struct {
	MemberID    string  `form:"member_id"`
	AccountID   string  `form:"account_id"`
	CategoryID  string  `form:"category_id"`
	Type        string  `form:"type"`
	StartDate   string  `form:"start_date"`
	EndDate     string  `form:"end_date"`
	MinAmount   float64 `form:"min_amount"`
	MaxAmount   float64 `form:"max_amount"`
	Description string  `form:"description"`
	SortBy      string  `form:"sort_by,default=date"`
	SortOrder   string  `form:"sort_order,default=desc"`
	Limit       int     `form:"limit,default=20"`
	Page        int     `form:"page,default=1"`
}
