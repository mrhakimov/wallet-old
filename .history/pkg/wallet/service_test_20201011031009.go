package wallet

import (
	"reflect"
	"testing"

	"github.com/MrHakimov/wallet/pkg/types"
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

	err := svc.ExportToFile("../../data/accounts.txt")
}
