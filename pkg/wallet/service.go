package wallet

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

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
	ErrFileNotFound         = errors.New("file not found")
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
	bufferf := make([]byte, 4)

	for {
		read, err := file.Read(bufferf)

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Print(err)

			return err
		}

		result = append(result, bufferf[:read]...)
	}

	data := string(result)

	for _, line := range strings.Split(data, "|") {
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

	for index, account := range accounts {
		nl := ""
		if index != 0 {
			nl = "\n"
		}

		_, err := file.Write([]byte(nl + strconv.FormatInt(account.ID, 10) + ";" +
			string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10)))

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

	for index, payment := range payments {
		nl := ""
		if index != 0 {
			nl = "\n"
		}

		_, err := file.Write([]byte(nl + payment.ID + ";" + strconv.FormatInt(payment.AccountID, 10) + ";" +
			strconv.FormatInt(int64(payment.Amount), 10) + ";" + string(payment.Category) + ";" +
			string(payment.Status)))

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

	for index, favorite := range favorites {
		nl := ""
		if index != 0 {
			nl = "\n"
		}

		_, err := file.Write([]byte(nl + favorite.ID + ";" + strconv.FormatInt(favorite.AccountID, 10) + ";" +
			favorite.Name + ";" + strconv.FormatInt(int64(favorite.Amount), 10) + ";" +
			string(favorite.Category)))

		if err != nil {
			log.Print(err)
			return err
		}
	}

	return err
}

// Import is used to update accounts, payments and favorites state from given files
func (s *Service) Import(dir string) error {
	fileAccounts, err := os.Open(dir + "/accounts.dump")

	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}

	if err != ErrFileNotFound {
		defer func() {
			err := fileAccounts.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		content := make([]byte, 0)
		buffer := make([]byte, 4)

		for {
			read, err := fileAccounts.Read(buffer)
			if err == io.EOF {
				break
			}

			content = append(content, buffer[:read]...)
		}

		data := strings.Split(string(content), "\n")

		for _, line := range data {
			account := &types.Account{}
			words := strings.Split(line, ";")

			for index, word := range words {
				switch index {
				case 0:
					id, _ := strconv.ParseInt(word, 10, 64)
					account.ID = id
					break
				case 1:
					account.Phone = types.Phone(word)
					break
				case 2:
					balance, _ := strconv.ParseInt(word, 10, 64)
					account.Balance = types.Money(balance)
					break
				}
			}

			exists := false
			for _, accountCheck := range s.accounts {
				if accountCheck.ID == account.ID {
					accountCheck.Phone = account.Phone
					accountCheck.Balance = account.Balance
					exists = true
				}

			}

			if !exists {
				s.accounts = append(s.accounts, account)
			}
		}
	}

	filePayments, err := os.Open(dir + "/payments.dump")

	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}

	if err != ErrFileNotFound {
		defer func() {
			err := filePayments.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		log.Printf("%#v", filePayments)

		contentPayment := make([]byte, 0)
		bufferPayment := make([]byte, 4)

		for {
			read, err := filePayments.Read(bufferPayment)
			if err == io.EOF {
				break
			}

			contentPayment = append(contentPayment, bufferPayment[:read]...)
		}

		dataPayment := strings.Split(string(contentPayment), "\n")

		for _, line := range dataPayment {
			payment := &types.Payment{}
			words := strings.Split(line, ";")

			for index, word := range words {
				switch index {
				case 0:
					payment.ID = word
					break
				case 1:
					accountID, _ := strconv.ParseInt(word, 10, 64)
					payment.AccountID = int64(accountID)
					break
				case 2:
					balance, _ := strconv.ParseInt(word, 10, 64)
					payment.Amount = types.Money(balance)
					break
				case 3:
					payment.Category = types.PaymentCategory(word)
					break
				case 4:
					payment.Status = types.PaymentStatus(word)
					break
				}
			}

			exists := false
			for _, paymentCheck := range s.payments {
				if paymentCheck.ID == payment.ID {
					paymentCheck.AccountID = payment.AccountID
					paymentCheck.Amount = payment.Amount
					paymentCheck.Category = payment.Category
					paymentCheck.Status = payment.Status
					exists = true
				}
			}

			if !exists {
				s.payments = append(s.payments, payment)
			}
		}
	}

	fileFavorites, err := os.Open(dir + "/favorites.dump")
	if err != nil {
		log.Print(err)
		err = ErrFileNotFound
	}

	if err != ErrFileNotFound {
		defer func() {
			err := fileFavorites.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		log.Printf("%#v", fileFavorites)

		contentFavorite := make([]byte, 0)
		bufferFavorite := make([]byte, 4)

		for {
			read, err := fileFavorites.Read(bufferFavorite)
			if err == io.EOF {
				break
			}

			contentFavorite = append(contentFavorite, bufferFavorite[:read]...)
		}

		dataFavorite := strings.Split(string(contentFavorite), "\n")

		for _, line := range dataFavorite {
			favorite := &types.Favorite{}
			words := strings.Split(line, ";")
			for index, word := range words {
				switch index {
				case 0:
					favorite.ID = word
					break
				case 1:
					accountID, _ := strconv.ParseInt(word, 10, 64)
					favorite.AccountID = int64(accountID)
					break
				case 2:
					favorite.Name = word
					break
				case 3:
					balance, _ := strconv.ParseInt(word, 10, 64)
					favorite.Amount = types.Money(balance)
					break
				case 4:
					favorite.Category = types.PaymentCategory(word)
					break
				}
			}

			exists := false
			for _, favoriteCheck := range s.favorites {
				if favoriteCheck.ID == favorite.ID {
					favoriteCheck.AccountID = favorite.AccountID
					favoriteCheck.Name = favorite.Name
					favoriteCheck.Amount = favorite.Amount
					favoriteCheck.Category = favorite.Category
					exists = true
				}
			}

			if !exists {
				s.favorites = append(s.favorites, favorite)
			}
		}
	}

	return nil
}

// ExportAccountHistory returns all payments of given user (by their accountID)
func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {
	_, err := s.FindAccountByID(accountID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	var payments []types.Payment = nil
	for _, payment := range s.payments {
		if payment.AccountID == accountID {
			payments = append(payments, *payment)
		}
	}

	if payments == nil {
		return nil, ErrAccountNotFound
	}

	return payments, nil
}

// HistoryToFiles is used to create backup files from payments history
func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {
	if len(payments) == 0 {
		return nil
	}

	var err error
	if len(payments) <= records {
		file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		defer func() {
			err = file.Close()
			if err != nil {
				log.Print(err)
			}
		}()

		var data string
		for _, account := range payments {
			data += fmt.Sprint(account.ID) + ";" + fmt.Sprint(account.AccountID) + ";" +
				fmt.Sprint(account.Amount) + ";" + fmt.Sprint(account.Category) + ";" +
				fmt.Sprint(account.Status) + "\n"
		}

		_, err = file.WriteString("data")
	} else {
		var file *os.File
		var data string
		k := 0
		t := 1

		for _, account := range payments {
			if k == 0 {
				file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			}
			k++
			data = fmt.Sprint(account.ID) + ";" + fmt.Sprint(account.AccountID) + ";" +
				fmt.Sprint(account.Amount) + ";" + fmt.Sprint(account.Category) + ";" +
				fmt.Sprint(account.Status) + "\n"

			_, err = file.Write([]byte(data))

			if k == records {
				data = ""
				k = 0
				t++

				err = file.Close()
			}
		}
	}

	return err
}

// Min returns the smaller of x or y
func Min(x int, y int) int {
	if y <= x {
		return y
	}

	return x
}

// SumPayments calculates the sum of all payments using goroutines
func (s *Service) SumPayments(goroutines int) types.Money {
	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	mu := sync.Mutex{}
	result := types.Money(0)

	paymentPerGoroutine := len(s.payments) / goroutines
	if len(s.payments)%goroutines != 0 {
		paymentPerGoroutine++
	}

	for i := 0; i < goroutines; i++ {
		currentSum := types.Money(0)
		index := i
		payments := s.payments

		go func(currentSum types.Money, index int, payments []*types.Payment) {
			defer wg.Done()

			for j := index * paymentPerGoroutine; j < Min((index+1)*paymentPerGoroutine, len(payments)); j++ {
				currentSum += payments[j].Amount
			}

			mu.Lock()
			result += currentSum
			mu.Unlock()
		}(currentSum, index, payments)
	}

	wg.Wait()

	return result
}

// FilterPaymentsByFn accepts filter function and finds all accounts which return true as a filter result
func (s *Service) FilterPaymentsByFn(filter func(payment types.Payment) bool, goroutines int) ([]types.Payment, error) {
	if goroutines == 0 {
		goroutines = 1
	}

	wg := sync.WaitGroup{}
	wg.Add(goroutines)

	mu := sync.Mutex{}
	var result []types.Payment = nil

	paymentPerGoroutine := len(s.payments) / goroutines
	if len(s.payments)%goroutines != 0 {
		paymentPerGoroutine++
	}

	for i := 0; i < goroutines; i++ {
		var currentAccounts []types.Payment = nil
		index := i
		payments := s.payments

		go func(index int) {
			defer wg.Done()

			for j := index * paymentPerGoroutine; j < Min((index+1)*paymentPerGoroutine, len(payments)); j++ {
				if filter(*payments[j]) {
					currentAccounts = append(currentAccounts, *payments[j])
				}
			}

			mu.Lock()
			result = append(result, currentAccounts...)
			mu.Unlock()
		}(index)
	}

	wg.Wait()

	if result == nil {
		return nil, ErrAccountNotFound
	}

	return result, nil
}

var globalAccountID int64

func filterByAccountID(payment types.Payment) bool {
	return payment.AccountID == globalAccountID
}

// FilterPayments accepts accountID and finds all accounts with such an ID using goroutines
func (s *Service) FilterPayments(accountID int64, goroutines int) ([]types.Payment, error) {
	globalAccountID = accountID

	return s.FilterPaymentsByFn(filterByAccountID, goroutines)
}

//SumPaymentsWithProgress is used to calculate payments' amount using channels
func (s *Service) SumPaymentsWithProgress() <-chan types.Progress {
	ch := make(chan types.Progress, 1)
	defer close(ch)

	if s.payments == nil {
		return ch
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(ch chan types.Progress) {
		defer wg.Done()

		sum := types.Progress{}

		for _, value := range s.payments {
			sum.Result += value.Amount
		}

		ch <- sum
	}(ch)

	wg.Wait()

	return ch
}
