package logic

import (
	"errors"
	"server/app/rpc/user/internal/pkg/encrypt"
	"testing"
)

func TestCreateTicketBuyerLogic(t *testing.T) {
	enc, err := encrypt.AESEncrypt("12345556656", "123")
	if err != nil {
		println(errors.New("数据加密失败"))
	}
	println("123")
	println(enc)
}
