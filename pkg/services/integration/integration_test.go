package integration

import (
	"testing"
	"time"

	"github.com/maelstrom/v3/pkg/mail"
	pkgsecurity "github.com/maelstrom/v3/pkg/security"
	svc "github.com/maelstrom/v3/pkg/services"
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

// mockLifecycleService implements the Service interface for testing
type mockLifecycleService struct {
	id string
}

func (m *mockLifecycleService) ID() string {
	return m.id
}

func (m *mockLifecycleService) HandleMail(mail mail.Mail) *svc.OutcomeEvent {
	return nil
}

func (m *mockLifecycleService) Start() error {
	return nil
}

func (m *mockLifecycleService) Stop() error {
	return nil
}

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

func TestLayer7_integrationAllHardcodedServicesStart(t *testing.T) {
	// Create all 4 hardcoded core services
	securitySvc := svcsecurity.NewSecurityService()
	commSvc := communication.NewCommunicationService()
	lifecycleSvc := &mockLifecycleService{id: "sys:lifecycle"}

	// Test that all services can start
	t.Run("security service starts", func(t *testing.T) {
		err := securitySvc.Start()
		if err != nil {
			t.Errorf("Failed to start security service: %v", err)
		}
	})

	t.Run("communication service starts", func(t *testing.T) {
		err := commSvc.Start()
		if err != nil {
			t.Errorf("Failed to start communication service: %v", err)
		}
	})

	t.Run("lifecycle service starts", func(t *testing.T) {
		err := lifecycleSvc.Start()
		if err != nil {
			t.Errorf("Failed to start lifecycle service: %v", err)
		}
	})

	// Test mock services can be registered
	sr := svc.NewServiceRegistry()
	err := sr.Register("sys:lifecycle", lifecycleSvc)
	if err != nil {
		t.Errorf("Failed to register lifecycle service: %v", err)
	}

	// Verify service is discoverable
	svcList := sr.DiscoverServices()
	if len(svcList) != 1 {
		t.Errorf("Expected 1 service, got %d", len(svcList))
	}
}

func TestLayer7_integrationAllHotReloadableServicesLoad(t *testing.T) {
	// Create all 8 hot-reloadable services
	gatewaySvc := gateway.NewGatewayService()
	adminSvc := admin.NewAdminService()
	persistenceSvc := persistence.NewPersistenceService()
	heartbeatSvc := heartbeat.NewHeartbeatService()
	memorySvc := memory.NewMemoryService()
	humanGatewaySvc := humangateway.NewHumanGatewayService()
	toolsSvc := tools.NewToolsService()
	datasourcesSvc := datasources.NewDatasourceService()

	// Test that all services can be instantiated and started
	t.Run("gateway service loads", func(t *testing.T) {
		if gatewaySvc == nil {
			t.Error("Failed to create gateway service")
		}
		err := gatewaySvc.Start()
		if err != nil {
			t.Errorf("Failed to start gateway service: %v", err)
		}
	})

	t.Run("admin service loads", func(t *testing.T) {
		if adminSvc == nil {
			t.Error("Failed to create admin service")
		}
		err := adminSvc.Start()
		if err != nil {
			t.Errorf("Failed to start admin service: %v", err)
		}
	})

	t.Run("persistence service loads", func(t *testing.T) {
		if persistenceSvc == nil {
			t.Error("Failed to create persistence service")
		}
		err := persistenceSvc.Start()
		if err != nil {
			t.Errorf("Failed to start persistence service: %v", err)
		}
	})

	t.Run("heartbeat service loads", func(t *testing.T) {
		if heartbeatSvc == nil {
			t.Error("Failed to create heartbeat service")
		}
		err := heartbeatSvc.Start()
		if err != nil {
			t.Errorf("Failed to start heartbeat service: %v", err)
		}
	})

	t.Run("memory service loads", func(t *testing.T) {
		if memorySvc == nil {
			t.Error("Failed to create memory service")
		}
		err := memorySvc.Start()
		if err != nil {
			t.Errorf("Failed to start memory service: %v", err)
		}
	})

	t.Run("human-gateway service loads", func(t *testing.T) {
		if humanGatewaySvc == nil {
			t.Error("Failed to create human-gateway service")
		}
		err := humanGatewaySvc.Start()
		if err != nil {
			t.Errorf("Failed to start human-gateway service: %v", err)
		}
	})

	t.Run("tools service loads", func(t *testing.T) {
		if toolsSvc == nil {
			t.Error("Failed to create tools service")
		}
		err := toolsSvc.Start()
		if err != nil {
			t.Errorf("Failed to start tools service: %v", err)
		}
	})

	t.Run("datasources service loads", func(t *testing.T) {
		if datasourcesSvc == nil {
			t.Error("Failed to create datasources service")
		}
		err := datasourcesSvc.Start()
		if err != nil {
			t.Errorf("Failed to start datasources service: %v", err)
		}
	})
}

func TestLayer7_integrationEndToEndServiceCommunication(t *testing.T) {
	// Create communication service for mail exchange
	commSvc := communication.NewCommunicationService()
	err := commSvc.Start()
	if err != nil {
		t.Fatalf("Failed to start communication service: %v", err)
	}

	// Create sender and receiver services
	senderSvc := gateway.NewGatewayService()
	receiverSvc := memory.NewMemoryService()

	err = senderSvc.Start()
	if err != nil {
		t.Fatalf("Failed to start sender service: %v", err)
	}

	err = receiverSvc.Start()
	if err != nil {
		t.Fatalf("Failed to start receiver service: %v", err)
	}

	// Subscribe receiver to a topic
	receiverChan, err := commSvc.Subscribe("sys:memory")
	if err != nil {
		t.Fatalf("Failed to subscribe receiver: %v", err)
	}

	// Create test mail
	testMail := mail.Mail{
		ID:     "e2e-test-1",
		Type:   mail.MailTypeUser,
		Source: "sys:gateway",
		Target: "sys:memory",
		Content: map[string]any{
			"message": "end-to-end communication test",
		},
		Metadata: mail.MailMetadata{
			Boundary: mail.InnerBoundary,
		},
	}

	// Publish mail
	ack, err := commSvc.Publish(testMail)
	if err != nil {
		t.Fatalf("Failed to publish mail: %v", err)
	}

	// Verify mail was delivered
	if !ack.Success {
		t.Error("Mail was not successfully delivered")
	}

	// Verify receiver gets the mail
	select {
	case receivedMail := <-receiverChan:
		if receivedMail.ID != testMail.ID {
			t.Errorf("Expected mail ID %s, got %s", testMail.ID, receivedMail.ID)
		}
		if receivedMail.Source != testMail.Source {
			t.Errorf("Expected source %s, got %s", testMail.Source, receivedMail.Source)
		}
		if receivedMail.Target != testMail.Target {
			t.Errorf("Expected target %s, got %s", testMail.Target, receivedMail.Target)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for mail delivery")
	}

	// Test that both services can HandleMail
	_ = senderSvc.HandleMail(testMail)
	_ = receiverSvc.HandleMail(testMail)
}
