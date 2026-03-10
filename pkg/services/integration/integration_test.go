package integration

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	"github.com/maelstrom/v3/pkg/services/admin"
	"github.com/maelstrom/v3/pkg/services/datasources"
	"github.com/maelstrom/v3/pkg/services/gateway"
	"github.com/maelstrom/v3/pkg/services/heartbeat"
	"github.com/maelstrom/v3/pkg/services/humangateway"
	"github.com/maelstrom/v3/pkg/services/memory"
	"github.com/maelstrom/v3/pkg/services/persistence"
	"github.com/maelstrom/v3/pkg/services/tools"
)

func TestServicesIntegration_HandleMail(t *testing.T) {
	// Create all 8 services
	services := []interface{}{
		gateway.NewGatewayService(),
		admin.NewAdminService(),
		persistence.NewPersistenceService(),
		heartbeat.NewHeartbeatService(),
		memory.NewMemoryService(),
		tools.NewToolsService(),
		datasources.NewDatasourceService(),
		humangateway.NewHumanGatewayService(),
	}

	// Test that all services implement HandleMail
	testMail := mail.Mail{
		ID:     "test-mail-1",
		Type:   mail.MailTypeUser,
		Source: "agent:test",
		Target: "sys:test",
		Content: map[string]any{
			"message": "test",
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
		},
	}

	for _, svc := range services {
		t.Run("HandleMail exists", func(t *testing.T) {
			// Use reflection to check HandleMail method exists
			handleMailFunc := svc.(interface{ HandleMail(mail.Mail) error })
			t.Logf("Service with HandleMail: %T", svc)
			_ = handleMailFunc
		})

		t.Run("HandleMail no panic", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("HandleMail panicked: %v", r)
				}
			}()

			handleMailFunc := svc.(interface{ HandleMail(mail.Mail) error })
			_ = handleMailFunc.HandleMail(testMail)
		})
	}
}
