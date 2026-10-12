package tests

import (
	"os"
	"testing"

	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

func TestMain(m *testing.M) {
	utils.SetJWTSecret("test-secret-that-is-at-least-32-bytes-long")
	os.Exit(m.Run())
}
