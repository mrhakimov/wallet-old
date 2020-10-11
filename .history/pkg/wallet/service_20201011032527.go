package wallet

import (
	"errors"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/MrHakimov/wallet/pkg/types"
)

// Error type for handling errors
type Error string

func (e Error) Error() string {
	return string(e)
}

// Errors
var (
	ErrPhoneNumberRegistred = errors.New("phone already registred")
	ErrAmountMustBePositive = errors.New("amount must be greater that zero")
	ErrAccountNotFound      = errors.New("account not found")
	ErrNotEnoughBalance     = errors.New("not enough balance")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrFavoriteNotFound     = errors.New("favorite not found")
)

// Service represents type for storing accounts and payments
type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

// RegisterAccount is used to register user by phone number
func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneNumberRegistred
		}
	}

	s.nextAccountID++

	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}

	s.accounts = append(s.accounts, account)

	return account, nil
}

// Deposit is used to create deposits
func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

// Pay is used for payments
func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}

	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	if account.Balance < amount {
		return nil, ErrNotEnoughBalance

	}
	account.Balance -= amount

	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}

	s.payments = append(s.payments, payment)
	return payment, nil

}

// FindAccountByID returns user by accountID
func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, nil
		}
	}

	return nil, ErrAccountNotFound
}

// FindPaymentByID returns payment by paymentID
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
		}
	}

	return nil, ErrPaymentNotFound
}

// FindFavoriteByID returns favorite payment by id
func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favoriteID {
			return favorite, nil
		}
	}

	return nil, ErrFavoriteNotFound
}

// Reject is used to reject payments
func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)

	if err != nil {
		return ErrPaymentNotFound
	}

	payment.Status = types.PaymentStatusFail

	account, err := s.FindAccountByID(payment.AccountID)

	if err != nil {
		return ErrAccountNotFound
	}

	account.Balance += payment.Amount

	return nil
}

// Repeat is used to make one more same payment
func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	payment, err := s.FindPaymentByID(paymentID)

	if err != nil {
		return nil, ErrPaymentNotFound
	}

	newPayment, err := s.Pay(payment.AccountID, payment.Amount, payment.Category)
	if err != nil {
		return nil, err
	}

	return newPayment, nil
}

// FavoritePayment is used to create new favorite payment
func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)

	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

// PayFromFavorite is just a wrapper for Pay
func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoriteByID(favoriteID)

	if err != nil {
		return nil, ErrFavoriteNotFound
	}

	payment, err := s.Pay(favorite.AccountID, favorite.Amount, favorite.Category)

	if err != nil {
		return nil, ErrPaymentNotFound
	}

	return payment, nil
}

// ExportToFile is used to export accounts data to file
func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)

	if err != nil {
		log.Print(err)

		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Print(err)
		}
	}()

	for _, account := range s.accounts {
		ID := strconv.FormatInt(int64(account.ID), 10) + ";"
		phone := string(account.Phone) + ";"
		balance := strconv.FormatInt(int64(account.Balance), 10)

		_, err = file.Write([]byte(ID + phone + balance + "|"))

		if err != nil {
			log.Print(err)

			return err
		}
	}

	return nil
}

// ImportFromFile is used to read accounts from file
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)

	if err != nil {
		log.Print(err)

		return err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Print(closeErr)
		}
	}()

	content := make([]byte, 0)
	buff := make([]byte, 4)

	for {
		read, err := file.Read(buff)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print(err)
			return err
		}
		content = append(content, buff[:read]...)
	}
	str := string(content)
	for _, line := range strings.Split(str, "|") {
		if len(line) <= 0 {
			return err
		}

		item := strings.Split(line, ";")
		ID, _ := strconv.ParseInt(item[0], 10, 64)
		balance, _ := strconv.ParseInt(item[2], 10, 64)

		s.accounts = append(s.accounts, &types.Account{
			ID:      ID,
			Phone:   types.Phone(item[1]),
			Balance: types.Money(balance),
		})
	}

	return err
}
