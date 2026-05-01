package errorx

const (
	OK                 = 0
	ServerError        = 1001
	ParamError         = 1002
	AuthError          = 1003
	NotFound           = 1004
	UserNotFound       = 2001
	UserAlreadyExist   = 2002
	PasswordError      = 2003
	ProductNotFound    = 3001
	StockNotEnough     = 3002
	OrderNotFound      = 4001
	OrderStatusError   = 4002
	CartEmpty          = 5001
	PaymentNotFound    = 6001
	PaymentStatusError = 6002

	// Activity center error codes (7xxx)
	ActivityNotFound            = 7001
	ActivityNotPublished        = 7002
	ActivityEnded               = 7003
	ActivityNotEligible         = 7004
	ActivityRateLimited         = 7005
	ActivityBlacklisted         = 7006
	ActivityTokenInvalid        = 7007
	ActivityStockEmpty          = 7008
	ActivityAlreadyParticipated = 7009
	RuleEvalFailed              = 7101
	WorkflowStateError          = 7201
	WorkflowTransitionError     = 7202
	RewardDispatchFailed        = 7301
	RewardAlreadyClaimed        = 7302
	RiskCheckFailed             = 7401
	SagaCompensated             = 7501

	// Review service error codes (8xxx)
	ReviewOrderNotFound      = 8001
	ReviewOrderNotCompleted  = 8002
	ReviewAlreadyExists      = 8003
	ReviewNotFound           = 8004
	ReviewFollowupNotAllowed = 8005
	ReviewRiskBlocked        = 8006
	ReviewMediaInvalid       = 8007
	ReviewLimitExceeded      = 8008
	AdminTokenInvalid        = 8009
)

var message = map[int]string{
	OK:                 "success",
	ServerError:        "server error",
	ParamError:         "invalid parameter",
	AuthError:          "unauthorized",
	NotFound:           "not found",
	UserNotFound:       "user not found",
	UserAlreadyExist:   "user already exists",
	PasswordError:      "wrong password",
	ProductNotFound:    "product not found",
	StockNotEnough:     "stock not enough",
	OrderNotFound:      "order not found",
	OrderStatusError:   "invalid order status",
	CartEmpty:          "cart is empty",
	PaymentNotFound:    "payment not found",
	PaymentStatusError: "invalid payment status",

	ActivityNotFound:            "activity not found",
	ActivityNotPublished:        "activity not published",
	ActivityEnded:               "activity ended",
	ActivityNotEligible:         "not eligible to participate",
	ActivityRateLimited:         "too many requests",
	ActivityBlacklisted:         "user blacklisted",
	ActivityTokenInvalid:        "invalid participation token",
	ActivityStockEmpty:          "activity stock empty",
	ActivityAlreadyParticipated: "already participated",
	RuleEvalFailed:              "rule evaluation failed",
	WorkflowStateError:          "invalid workflow state",
	WorkflowTransitionError:     "invalid workflow transition",
	RewardDispatchFailed:        "reward dispatch failed",
	RewardAlreadyClaimed:        "reward already claimed",
	RiskCheckFailed:             "risk check failed",
	SagaCompensated:             "distributed transaction compensated",

	ReviewOrderNotFound:      "review: order not found or not owned by user",
	ReviewOrderNotCompleted:  "review: order not completed",
	ReviewAlreadyExists:      "review: order item already reviewed",
	ReviewNotFound:           "review not found or deleted",
	ReviewFollowupNotAllowed: "review: followup not allowed",
	ReviewRiskBlocked:        "review: blocked by risk control",
	ReviewMediaInvalid:       "review: invalid media url",
	ReviewLimitExceeded:      "review: content/media size limit exceeded",
	AdminTokenInvalid:        "invalid admin token",
}

type CodeError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewCodeError(code int) *CodeError {
	return &CodeError{Code: code, Msg: message[code]}
}

func NewCodeErrorMsg(code int, msg string) *CodeError {
	return &CodeError{Code: code, Msg: msg}
}

func (e *CodeError) Error() string {
	return e.Msg
}
