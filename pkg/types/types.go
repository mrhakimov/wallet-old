package types

// Money represents type for storing money
type Money int64

// PaymentCategory represents possible categories of payments
type PaymentCategory string

// PaymentStatus represents status of payment
type PaymentStatus string

// statuses
const (
	PaymentStatusOk         PaymentStatus = "OK"
	PaymentStatusFail       PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

// Payment represents payment data
type Payment struct {
	ID        string
	AccountID int64
	Amount    Money
	Category  PaymentCategory
	Status    PaymentStatus
}

// Favorite is used for featured payments
type Favorite struct {
	ID        string
	AccountID int64
	Name      string
	Amount    Money
	Category  PaymentCategory
}

// Phone is used for telephone numbers
type Phone string

// Account is used to store user's data
type Account struct {
	ID      int64
	Phone   Phone
	Balance Money
}

// Progress is used to store sum calculation progress
type Progress struct {
	Part   int
	Result Money
}
