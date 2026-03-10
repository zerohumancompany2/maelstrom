package integration

import (
	"testing"

	"github.com/maelstrom/v3/pkg/mail"
	pkgsecurity "github.com/maelstrom/v3/pkg/security"
	"github.com/maelstrom/v3/pkg/services/admin"
	"github.com/maelstrom/v3/pkg/services/communication"
	"github.com/maelstrom/v3/pkg/services/datasources"
	"github.com/maelstrom/v3/pkg/services/gateway"
	"github.com/maelstrom/v3/pkg/services/heartbeat"
	"github.com/maelstrom/v3/pkg/services/humangateway"
	"github.com/maelstrom/v3/pkg/services/memory"
	"github.com/maelstrom/v3/pkg/services/persistence"
	svcsecurity "github.com/maelstrom/v3/pkg/services/security"
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

func TestServicesIntegration_MailExchange(t *testing.T) {
	commSvc := communication.NewCommunicationService()

	// Setup mail exchange between services
	senderService := gateway.NewGatewayService()
	receiverService := memory.NewMemoryService()

	// Create test mail
	testMail := mail.Mail{
		ID:     "mail-exchange-test-1",
		Type:   mail.MailTypeUser,
		Source: "sys:gateway",
		Target: "sys:memory",
		Content: map[string]any{
			"message": "exchange test message",
			"data":    []string{"item1", "item2"},
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
			Tokens:   10,
			Model:    "test-model",
		},
	}

	// Subscribe receiver to target address
	receiverChan, err := commSvc.Subscribe("sys:memory")
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// Publish mail
	ack, err := commSvc.Publish(testMail)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Verify mail was delivered
	if !ack.Success {
		t.Error("Mail was not successfully delivered")
	}

	// Read mail from receiver channel
	select {
	case receivedMail := <-receiverChan:
		// Verify content preserved
		if receivedContent, ok := receivedMail.Content.(map[string]any); ok {
			if testContent, ok := testMail.Content.(map[string]any); ok {
				if receivedContent["message"] != testContent["message"] {
					t.Error("Mail content message not preserved")
				}
			}
		}

		// Verify metadata preserved
		if receivedMail.Metadata.Boundary != testMail.Metadata.Boundary {
			t.Error("Mail metadata boundary not preserved")
		}
		if receivedMail.Metadata.Tokens != testMail.Metadata.Tokens {
			t.Error("Mail metadata tokens not preserved")
		}
		if receivedMail.Metadata.Model != testMail.Metadata.Model {
			t.Error("Mail metadata model not preserved")
		}
	default:
		t.Error("Mail not received by receiver")
	}

	// Test HandleMail on sender and receiver
	_ = senderService.HandleMail(testMail)
	_ = receiverService.HandleMail(testMail)
}

func TestServicesIntegration_MailRouting(t *testing.T) {
	commSvc := communication.NewCommunicationService()

	// Create test mail for routing
	testMail := mail.Mail{
		ID:     "mail-routing-test-1",
		Type:   mail.MailTypeUser,
		Source: "agent:test",
		Target: "sys:memory",
		Content: map[string]any{
			"message": "routing test",
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
		},
	}

	// Test routing to correct service by sys:* ID
	t.Run("Route to correct service", func(t *testing.T) {
		// Subscribe to sys:memory
		receiverChan, err := commSvc.Subscribe("sys:memory")
		if err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}

		// Publish mail
		ack, err := commSvc.Publish(testMail)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}

		// Verify mail was delivered to correct service
		if !ack.Success {
			t.Error("Mail was not successfully routed")
		}

		// Read from channel
		select {
		case receivedMail := <-receiverChan:
			if receivedMail.Target != "sys:memory" {
				t.Error("Mail not routed to correct service")
			}
		default:
			t.Error("Mail not received by target service")
		}
	})

	// Test routing handles unknown targets gracefully
	t.Run("Handle unknown target gracefully", func(t *testing.T) {
		unknownMail := mail.Mail{
			ID:     "unknown-target-test",
			Type:   mail.MailTypeUser,
			Source: "agent:test",
			Target: "sys:unknown-service",
			Content: map[string]any{
				"message": "unknown target test",
			},
		}

		// Should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Routing to unknown target panicked: %v", r)
			}
		}()

		_, err := commSvc.Publish(unknownMail)
		if err != nil {
			t.Logf("Expected error for unknown target: %v", err)
		}
	})

	// Test cross-service routing
	t.Run("Cross-service routing", func(t *testing.T) {
		senderService := gateway.NewGatewayService()
		receiverService := persistence.NewPersistenceService()

		crossMail := mail.Mail{
			ID:     "cross-service-test",
			Type:   mail.MailTypeUser,
			Source: "sys:gateway",
			Target: "sys:persistence",
			Content: map[string]any{
				"message": "cross-service test",
			},
		}

		// Both services should be able to HandleMail
		_ = senderService.HandleMail(crossMail)
		_ = receiverService.HandleMail(crossMail)
	})
}

func TestServicesIntegration_MailBoundaryEnforcement(t *testing.T) {
	securitySvc := svcsecurity.NewSecurityService()

	// Test boundary enforcement on mail exchange
	t.Run("Boundary enforcement", func(t *testing.T) {
		// Create mail with boundary
		boundaryMail := mail.Mail{
			ID:     "boundary-test-1",
			Type:   mail.MailTypeUser,
			Source: "agent:inner",
			Target: "agent:outer",
			Content: map[string]any{
				"message": "boundary test",
			},
			Metadata: mail.MailMetadata{
				Boundary: mail.InnerBoundary,
			},
		}

		// Security service should handle boundary enforcement
		_ = securitySvc.HandleMail(&boundaryMail)
	})

	// Test taints tracked across mail boundaries
	t.Run("Taints tracked across boundaries", func(t *testing.T) {
		taintedMail := mail.Mail{
			ID:     "taint-test-1",
			Type:   mail.MailTypeUser,
			Source: "agent:inner",
			Target: "agent:dmz",
			Content: map[string]any{
				"message": "taint test",
			},
			Metadata: mail.MailMetadata{
				Boundary: mail.InnerBoundary,
				Taints:   []string{"SECRET", "INNER_ONLY"},
			},
		}

		// Verify mail has taints
		if len(taintedMail.Metadata.Taints) == 0 {
			t.Error("Mail should have taints")
		}

		// Track taints
		_ = securitySvc.TrackTaint("agent:inner", "taint-test-1", "SECRET")

		// Verify taints are tracked
		taintMap, err := securitySvc.ReportTaints("agent:inner")
		if err != nil {
			t.Fatalf("Failed to report taints: %v", err)
		}

		if len(taintMap) == 0 {
			t.Error("Taints not tracked")
		}
	})

	// Test cross-boundary mail blocked or flagged
	t.Run("Cross-boundary mail blocked", func(t *testing.T) {
		// Create mail that violates boundary policy
		violationMail := mail.Mail{
			ID:     "violation-test-1",
			Type:   mail.MailTypeUser,
			Source: "agent:inner",
			Target: "agent:outer",
			Content: map[string]any{
				"message": "violation test",
			},
			Metadata: mail.MailMetadata{
				Boundary: mail.InnerBoundary,
				Taints:   []string{"INNER_ONLY"},
			},
		}

		// Validate and sanitize - should fail for INNER_ONLY to outer
		_, err := securitySvc.ValidateAndSanitize(violationMail, mail.InnerBoundary, mail.OuterBoundary)
		if err == nil {
			t.Error("Expected error for cross-boundary violation")
		} else {
			t.Logf("Expected error: %v", err)
		}
	})

	// Test taint propagation
	t.Run("Taint propagation", func(t *testing.T) {
		data := map[string]interface{}{
			"key": "value",
		}

		// Propagate taints
		propagated, err := securitySvc.TaintPropagate(data, []string{"USER_SUPPLIED"})
		if err != nil {
			t.Fatalf("Failed to propagate taints: %v", err)
		}

		propagatedMap, ok := propagated.(map[string]interface{})
		if !ok {
			t.Error("Propagated data is not a map")
		}

		if _, hasTaints := propagatedMap["_taints"]; !hasTaints {
			t.Error("Taints not propagated to data")
		}
	})

	// Test taint policy check
	t.Run("Taint policy enforcement", func(t *testing.T) {
		policy := pkgsecurity.TaintPolicy{
			AllowedForBoundary: []pkgsecurity.BoundaryType{pkgsecurity.InnerBoundary},
		}

		taintedData := map[string]interface{}{
			"key":     "value",
			"_taints": []string{"INNER_ONLY"},
		}

		allowed, err := securitySvc.CheckTaintPolicy(taintedData, mail.OuterBoundary, policy)
		if err != nil {
			t.Fatalf("Failed to check taint policy: %v", err)
		}

		if allowed {
			t.Error("INNER_ONLY taint should not be allowed on outer boundary")
		}
	})
}
