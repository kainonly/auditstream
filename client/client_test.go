package client

//
//import (
//	"fmt"
//	"github.com/weplanx/transfer/app"
//	"github.com/weplanx/transfer/bootstrap"
//	"go.mongodb.org/mongo-driver/bson/primitive"
//	"os"
//	"testing"
//)
//
//var x *Transfer
//
//func TestMain(m *testing.M) {
//	os.Chdir("../")
//	values, err := bootstrap.SetValues()
//	if err != nil {
//		panic(err)
//	}
//	if x, err = New(
//		fmt.Sprintf(`transfer.kainonly.com%s`, values.Address),
//		values.TLS,
//	); err != nil {
//		panic(err)
//	}
//	os.Exit(m.Run())
//}
//
//func TestTransfer_CreateLogger(t *testing.T) {
//	defer x.Close()
//	if err := x.CreateLogger(app.CreateLoggerRequest{
//		Topic:       "beta",
//		Description: "Transfer 测试",
//	}); err != nil {
//		t.Error(err)
//	}
//}
//
//var data []app.Logger
//
//func TestTransfer_Logger(t *testing.T) {
//	defer x.Close()
//	result, err := x.Logger()
//	if err != nil {
//		t.Error(err)
//	}
//	t.Log(result)
//	data = result.Data
//}
//
//func TestTransfer_DeleteLogger(t *testing.T) {
//	defer x.Close()
//	oid, err := primitive.ObjectIDFromHex("6208b5dd6cd5114a431e0495")
//	if err != nil {
//		t.Error(err)
//	}
//	if err = x.DeleteLogger(app.DeleteLoggerRequest{Id: oid}); err != nil {
//		t.Error(err)
//	}
//}
