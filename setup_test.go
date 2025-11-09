package chunkx

import (
	"os"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/approvals/go-approval-tests/reporters"
)

func TestMain(m *testing.M) {
	r := approvals.UseReporter(reporters.NewContinuousIntegrationReporter())
	defer r.Close()
	approvals.UseFolder("testdata")
	os.Exit(m.Run())
}
