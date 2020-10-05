package wallet

import (
	"testing"
)

func TestService_FindbyAccountById_success(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+992000000000")
	_, err := svc.FindAccountByID(1)
	if err != nil {
		t.Error(err)
	}

}

func TestService_FindByAccountByID_notFound(t *testing.T) {
	svc := Service{}
	svc.RegisterAccount("+992000000000")
	_, err := svc.FindAccountByID(2)

	if err != ErrAccountNotFound {
		t.Error(err)
	}
}
