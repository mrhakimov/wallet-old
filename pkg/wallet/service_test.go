package wallet

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/MrHakimov/wallet/pkg/types"
	"github.com/google/uuid"
)

func TestService_RegisterAccount_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestService_FindAccoundByIdmethod_notFound(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(2)
	if err == nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestDeposit(t *testing.T) {
	svc := Service{}

	svc.RegisterAccount("+992000000001")

	err := svc.Deposit(1, 100_00)
	if err != nil {
		t.Error("something wrong while paying")
	}

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", account)
	}
}

func TestService_FindbyAccountById_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+992000000000")
	_, err := svc.FindAccountByID(1)

	if err != nil {
		t.Error(err)
	}
}

func TestFindPaymentByID_success(t *testing.T) {
	svc := &Service{}

	phone := types.Phone("+992000000000")

	account, err := svc.RegisterAccount(phone)
	if err != nil {
		t.Error(err)
		return
	}

	err = svc.Deposit(account.ID, 1000)
	if err != nil {
		t.Error(err)
		return
	}

	pay, err := svc.Pay(account.ID, 500, "auto")
	if err != nil {
		t.Error(err)
		return
	}

	got, err := svc.FindPaymentByID(pay.ID)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(got, pay) {
		t.Error(err)
		return
	}
}

func TestFindFavoriteByID_success(t *testing.T) {
	svc := &Service{}

	phone := types.Phone("+992000000000")

	account, err := svc.RegisterAccount(phone)
	if err != nil {
		t.Error(err)
		return
	}

	err = svc.Deposit(account.ID, 1000)
	if err != nil {
		t.Error(err)
		return
	}

	pay, err := svc.Pay(account.ID, 500, "auto")
	if err != nil {
		t.Error(err)
		return
	}

	_, err = svc.FindFavoriteByID(pay.ID)
	if err == nil {
		t.Error(err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	svc := &Service{}

	phone := types.Phone("+992000000000")

	account, err := svc.RegisterAccount(phone)
	if err != nil {
		t.Error(err)
		return
	}

	err = svc.Deposit(account.ID, 1000)
	if err != nil {
		t.Error(err)
		return
	}

	pay, err := svc.Pay(account.ID, 500, "auto")
	if err != nil {
		t.Error(err)
		return
	}

	err = svc.Reject(pay.ID)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestService_Reject_fail(t *testing.T) {
	svc := Service{}

	svc.RegisterAccount("+992000000000")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Error(err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Error(err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Error(err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Error(pay)
	}

	editPayID := "4"

	err = svc.Reject(editPayID)
	if err != ErrPaymentNotFound {
		t.Error(err)
	}
}

func TestService_Repeat_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+9920000001")

	account, err := svc.FindAccountByID(1)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	err = svc.Deposit(account.ID, 1000_00)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	payment, err := svc.Pay(account.ID, 100_00, "auto")
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err := svc.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("\ngot > %v \nwant > nil", err)
	}

	pay, err = svc.Repeat(pay.ID)
	if err != nil {
		t.Errorf("Repeat(): Error(): can't pay for an account(%v): %v", pay.ID, err)
	}
}

func TestService_Favorite_success_user(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 10_00, "auto")
	if err != nil {
		t.Errorf("Pay() Error() can't pay for an account(%v): %v", account, err)
	}

	favorite, err := svc.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		t.Errorf("FavoritePayment() Error() can't for an favorite(%v): %v", favorite, err)
	}

	paymentFavorite, err := svc.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavorite, err)
	}

	paymentFavoriteFail, err := svc.PayFromFavorite(payment.ID)
	if err == nil {
		t.Errorf("PayFromFavorite() Error() can't for an favorite(%v): %v", paymentFavoriteFail, err)
	}
}

func TestService_ExportToFile(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	err = svc.ExportToFile("../../data/accounts.txt")
	if err != nil {
		t.Error("Error occurred while exporting to file!", err)
	}
}

func TestService_ImportFromFile(t *testing.T) {
	svc := Service{}

	err := svc.ImportFromFile("../../data/accounts.txt")
	if err != nil {
		t.Error("Error occurred while importing from file!", err)
	}

	err = svc.ImportFromFile("../../data/accountsFake.txt")
	if err == nil {
		t.Error("Error occurred while importing from file!", err)
	}
}

func TestService_Export_Success(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 100_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	payment, err := svc.Pay(account.ID, 500, "auto")
	if err != nil {
		t.Error(err)
		return
	}

	favorite, err := svc.FavoritePayment(payment.ID, "megafon")
	if err != nil {
		t.Errorf("FavoritePayment() Error() can't for an favorite(%v): %v", favorite, err)
	}

	account, err = svc.RegisterAccount("+992000000002")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 200_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = svc.Export(wd + "/test1")
	if err != nil {
		t.Error(err)
	}
}

func TestService_Import_Success(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000003")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	err = svc.Deposit(account.ID, 300_00)
	if err != nil {
		t.Errorf("method Deposit returned not nil error, error => %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = svc.Import(wd + "/test1")
	if err != nil {
		t.Error(err)
	}
}

func TestService_Import_Fail(t *testing.T) {
	svc := Service{}

	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = svc.Import(wd + "/..")
	if err != nil {
		t.Error(err)
	}
}

func TestService_ExportAccountHistory_Success(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	account.Balance = 300_000_00

	var payments []types.Payment = nil
	payment, err := svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)
	foundPayments, err := svc.ExportAccountHistory(account.ID)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(foundPayments, payments) {
		t.Error(err)
	}
}

func TestService_HistoryToFiles_Success(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	account.Balance = 300_000_00

	var payments []types.Payment = nil
	payment, err := svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)

	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = svc.HistoryToFiles(payments, dir+"/test1", 1)
	if err != nil {
		t.Error(err)
	}
}

func TestService_HistoryToFiles_Multiple(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	account.Balance = 300_000_00

	var payments []types.Payment = nil
	payment, err := svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)

	payment, err = svc.Pay(account.ID, types.Money(1000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	err = svc.HistoryToFiles(payments, dir+"/test1", 1)
	if err != nil {
		t.Error(err)
	}
}

func TestService_SumPayments(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	account.Balance = 300_000_00

	var payments []types.Payment = nil
	payment, err := svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)

	payment, err = svc.Pay(account.ID, types.Money(1000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	payments = append(payments, *payment)

	for _, payment := range payments {
		fmt.Println(payment.Amount)
	}

	sum := svc.SumPayments(2)
	fmt.Println(sum)
	if sum != 4000_00 {
		t.Error(err)
		return
	}
}

func BenchmarkSumPayments(b *testing.B) {
	want := types.Money(3000_00)
	for i := 0; i < b.N; i++ {
		svc := Service{}

		account, err := svc.RegisterAccount("+992000000001")
		if err != nil {
			b.Errorf("method RegisterAccount returned not nil error, account => %v", account)
		}

		account.Balance = 300_000_00

		_, err = svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
		if err != nil {
			b.Error(err)
			return
		}

		result := svc.SumPayments(1)
		if want != result {
			b.Fatalf("invalid result, got %v, want %v", result, want)
		}
	}
}

// testFilter is just a test function
func testFilter(payment types.Payment) bool {
	return payment.AccountID == globalAccountID
}

func BenchmarkFilterPayments(b *testing.B) {
	for i := 0; i < b.N; i++ {
		svc := Service{}

		account, err := svc.RegisterAccount("+992000000001")
		if err != nil {
			b.Errorf("method RegisterAccount returned not nil error, account => %v", account)
		}

		account.Balance = 300_000_00

		_, err = svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
		if err != nil {
			b.Error(err)
			return
		}

		svc.FilterPaymentsByFn(testFilter, 5)

		result, _ := svc.FilterPayments(account.ID, 5)
		if len(result) != 1 {
			b.Fatal()
		}
	}
}

func TestService_FilterPaymentsByFn(t *testing.T) {
	svc := Service{}

	account, err := svc.RegisterAccount("+992000000001")
	if err != nil {
		t.Errorf("method RegisterAccount returned not nil error, account => %v", account)
	}

	account.Balance = 300_000_00

	_, err = svc.Pay(account.ID, types.Money(3000_00), types.PaymentCategory("OK"))
	if err != nil {
		t.Error(err)
		return
	}

	svc.FilterPaymentsByFn(testFilter, 5)
}

func TestService_SumPaymentsWithProgress(t *testing.T) {
	svc := Service{}
	for i := 0; i < 300_000; i++ {
		payment := &types.Payment{
			ID:     uuid.New().String(),
			Amount: types.Money(100_00),
		}
		svc.payments = append(svc.payments, payment)
	}

	svc.SumPaymentsWithProgress()
}
