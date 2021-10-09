package query

import (
	"os"
	"testing"

	"github.com/yaoapp/xun/capsule"
)

// TestAPIRoot
var TestAPIRoot = "/data/apis"
var TestFLWRoot = "/data/flows"
var TestPLGRoot = "/data/plugins"
var TestModRoot = "/data/models"
var TestDriver = "mysql"
var TestDSN = "root:123456@tcp(127.0.0.1:3306)/gou?charset=utf8mb4&parseTime=True&loc=Local"
var TestAESKey = "123456"

func TestMain(m *testing.M) {

	TestAPIRoot = os.Getenv("GOU_TEST_API_ROOT")
	TestFLWRoot = os.Getenv("GOU_TEST_FLW_ROOT")
	TestModRoot = os.Getenv("GOU_TEST_MOD_ROOT")
	TestPLGRoot = os.Getenv("GOU_TEST_PLG_ROOT")
	TestDriver = os.Getenv("GOU_TEST_DB_DRIVER")
	TestDSN = os.Getenv("GOU_TEST_DSN")
	TestAESKey = os.Getenv("GOT_TEST_AES_KEY")

	// 数据库连接
	switch TestDriver {
	case "sqlite3":
		capsule.AddConn("primary", "sqlite3", TestDSN).SetAsGlobal()
		break
	default:
		capsule.AddConn("primary", "mysql", TestDSN).SetAsGlobal()
		break
	}

	// Run test suites
	exitVal := m.Run()

	// we can do clean up code here
	os.Exit(exitVal)

}