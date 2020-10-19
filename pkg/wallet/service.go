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
		if err := file.Close(); err != nil {
			log.Print(err)
		}
	}()

	result := make([]byte, 0)
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

		result = append(result, buff[:read]...)
	}

	str := string(result)

	for _, line := range strings.Split(str, "|") {
		if len(line) == 0 {
			return err
		}

		item := strings.Split(line, ";")
		ID, err := strconv.ParseInt(item[0], 10, 64)

		if err != nil {
			return err
		}

		balance, err := strconv.ParseInt(item[2], 10, 64)

		if err != nil {
			return err
		}

		s.accounts = append(s.accounts, &types.Account{
			ID:      ID,
			Phone:   types.Phone(item[1]),
			Balance: types.Money(balance),
		})
	}

	return nil
}

// Export is used to save all payments, accounts and favorites into file
func (s *Service) Export(dir string) error {
	err := WriteAccountsToFile(dir+"/accounts.dump", s.accounts)
	if err != nil {
		return err
	}

	err = WritePaymentsToFile(dir+"/payments.dump", s.payments)
	if err != nil {
		return err
	}

	err = WriteFavoritesToFile(dir+"/favorites.dump", s.favorites)

	return err
}

// WriteAccountsToFile is a helper function to write accounts to respective file
func WriteAccountsToFile(filePath string, accounts []*types.Account) error {
	if len(accounts) == 0 {
		return nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	for _, account := range accounts {
		_, err := file.Write([]byte(strconv.FormatInt(account.ID, 10) + ";" +
			string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10) + "\n"))

		if err != nil {
			log.Print(err)
			return err
		}
	}

	return err
}

// WritePaymentsToFile is a helper function to write payments to respective file
func WritePaymentsToFile(filePath string, payments []*types.Payment) error {
	if len(payments) == 0 {
		return nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	for _, payment := range payments {
		_, err := file.Write([]byte(payment.ID + ";" + strconv.FormatInt(payment.AccountID, 10) + ";" +
			strconv.FormatInt(int64(payment.Amount), 10) + ";" + string(payment.Category) + ";" +
			string(payment.Status) + "\n"))

		if err != nil {
			log.Print(err)
			return err
		}
	}

	return err
}

// WriteFavoritesToFile is a helper function to write favorite payments to respective file
func WriteFavoritesToFile(filePath string, favorites []*types.Favorite) error {
	if len(favorites) == 0 {
		return nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		log.Print(err)
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	for _, favorite := range favorites {
		_, err := file.Write([]byte(favorite.ID + ";" + strconv.FormatInt(favorite.AccountID, 10) + ";" +
			favorite.Name + ";" + strconv.FormatInt(int64(favorite.Amount), 10) + ";" +
			string(favorite.Category) + "\n"))

		if err != nil {
			log.Print(err)
			return err
		}
	}

	return err
}

// Import is used to update accounts, payments and favorites state from given files
func (s *Service) Import(dir string) error {
	err := UpdateAccountsFromFile(s, dir+"/accounts.dump")
	if err != nil {
		return err
	}

	err = UpdatePaymentsFromFile(s, dir+"/payments.dump")
	if err != nil {
		return err
	}

	err = UpdateFavoritesFromFile(s, dir+"/favorites.dump")

	return err
}

// UpdateAccountsFromFile is used to update accounts states
func UpdateAccountsFromFile(s *Service, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	content := make([]byte, 0)
	buffer := make([]byte, 4)

	for {
		read, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		content = append(content, buffer[:read]...)
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		account := &types.Account{}
		data := strings.Split(line, ";")

		if len(data) != 3 {
			continue
		}

		id, err := strconv.ParseInt(data[0], 10, 64)
		if err != nil {
			log.Print(err)
			return err
		}

		account.ID = id
		account.Phone = types.Phone(data[1])

		balance, err := strconv.ParseInt(data[2], 10, 64)
		if err != nil {
			log.Print(err)
			return err
		}

		account.Balance = types.Money(balance)

		foundAccount, _ := s.FindAccountByID(account.ID)
		if foundAccount == nil {
			s.accounts = append(s.accounts, account)
			return nil
		}

		foundAccount.Phone = account.Phone
		foundAccount.Balance = account.Balance
	}

	return err
}

// UpdatePaymentsFromFile is used to update payments states
func UpdatePaymentsFromFile(s *Service, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	content := make([]byte, 0)
	buffer := make([]byte, 4)

	for {
		read, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		content = append(content, buffer[:read]...)
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		payment := &types.Payment{}
		data := strings.Split(line, ";")

		if len(data) != 5 {
			continue
		}

		payment.ID = data[0]
		accountID, err := strconv.ParseInt(data[1], 10, 64)
		if err != nil {
			return err
		}

		payment.AccountID = accountID

		amount, err := strconv.ParseInt(data[2], 10, 64)
		if err != nil {
			return err
		}

		payment.Amount = types.Money(amount)
		payment.Category = types.PaymentCategory(data[3])
		payment.Status = types.PaymentStatus(data[4])

		foundPayment, _ := s.FindPaymentByID(payment.ID)

		if foundPayment == nil {
			s.payments = append(s.payments, payment)
			return nil
		}

		foundPayment.AccountID = payment.AccountID
		foundPayment.Amount = payment.Amount
		foundPayment.Category = payment.Category
		foundPayment.Status = payment.Status
	}

	return err
}

// UpdateFavoritesFromFile is used to update payments states
func UpdateFavoritesFromFile(s *Service, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	content := make([]byte, 0)
	buffer := make([]byte, 4)

	for {
		read, err := file.Read(buffer)
		if err == io.EOF {
			break
		}

		content = append(content, buffer[:read]...)
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		favorite := &types.Favorite{}
		data := strings.Split(line, ";")

		if len(data) != 5 {
			continue
		}

		favorite.ID = data[0]
		accountID, err := strconv.ParseInt(data[1], 10, 64)
		if err != nil {
			return err
		}

		favorite.AccountID = accountID
		favorite.Name = data[2]

		amount, err := strconv.ParseInt(data[3], 10, 64)
		if err != nil {
			return err
		}

		favorite.Amount = types.Money(amount)
		favorite.Category = types.PaymentCategory(data[4])

		foundFavorite, _ := s.FindFavoriteByID(favorite.ID)

		if foundFavorite == nil {
			s.favorites = append(s.favorites, favorite)
			return nil
		}

		foundFavorite.AccountID = favorite.AccountID
		foundFavorite.Name = favorite.Name
		foundFavorite.Amount = favorite.Amount
		foundFavorite.Category = favorite.Category
	}

	return err
}
