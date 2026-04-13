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
