package sub

import (
	"net/url"
	"strconv"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
 "github.com/mhsanaei/3x-ui/v3/internal/database"
)

func TestMultilocationHostNameOverridesGlobalTemplate(t *testing.T) {
	for _, template := range []string{"", "{{INBOUND}}-{{EMAIL}}", "{{EMAIL}}|{{TRAFFIC_LEFT}}"} {
		for _, flag := range []string{"🇹🇷", "🇺🇸", "🇫🇷", "🇪🇸", "🇳🇱", "🇮🇹", "🇸🇪", "🇬🇧"} {
			for _, email := range []string{"farnaz_skm-direct", "farnaz_skm-tunnel", "farnaz_skm"} {
				s := NewSubService(template)
				inbound := &model.Inbound{Port: 2083, Protocol: model.VLESS, Remark: "Andromeda (clone)"}
				client := model.Client{Email: email, SubID: "tg-123-unchanged"}
				before := client
				if got := s.genHostRemark(inbound, client, flag, "ws"); got != "farnaz_skm"+flag {
					t.Fatalf("template=%q email=%q flag=%q: got %q", template, email, flag, got)
				}
				if client.Email != before.Email || client.SubID != before.SubID {
					t.Fatal("rendering changed the stored client identity")
				}
			}
		}
	}
}

func TestMultilocationHostNameKeepsOtherLabels(t *testing.T) {
 seedSubDB(t)
	for _, tc := range []struct {
		port     int
		protocol model.Protocol
		host     string
	}{
		{443, model.VLESS, "🇺🇸"},
		{8080, model.VLESS, "🇺🇸"},
		{2083, model.Trojan, "🇺🇸"},
		{2083, model.VLESS, "CDN"},
		{2083, model.VLESS, ""},
	} {
		inbound := &model.Inbound{Port: tc.port, Protocol: tc.protocol, Remark: "Original"}
		client := model.Client{Email: "farnaz_skm-direct"}
		if got := NewSubService("{{INBOUND}}-{{EMAIL}}").genHostRemark(inbound, client, tc.host, "ws"); got != "Original-farnaz_skm-direct" {
			t.Fatalf("port=%d host=%q: unrelated label changed to %q", tc.port, tc.host, got)
		}
	}
}

func TestMultilocationHostLinksHaveDistinctFlagsAndRoutes(t *testing.T) {
 seedSubDB(t)
	inbound := &model.Inbound{
		Port: 2083, Protocol: model.VLESS, Remark: "Andromeda (clone)",
		Settings: `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"farnaz_skm-direct","enable":true}],"decryption":"none"}`,
		StreamSettings: `{"network":"ws","security":"tls","wsSettings":{"path":"/","host":"example.com"},"tlsSettings":{"serverName":"example.com"}}`,
	}
	client := model.Client{Email: "farnaz_skm-direct", SubID: "tg-123-test"}
 db := database.GetDB()
 if err := db.Create(inbound).Error; err != nil { t.Fatal(err) }
 record := &model.ClientRecord{Email: client.Email, SubID: client.SubID, UUID: "11111111-2222-4333-8444-555555555555", Enable: true}
 if err := db.Create(record).Error; err != nil { t.Fatal(err) }
 if err := db.Create(&model.ClientInbound{ClientId: record.Id, InboundId: inbound.Id}).Error; err != nil { t.Fatal(err) }
 for flag, route := range map[string]string{
		"🇹🇷": "9081", "🇺🇸": "9082", "🇫🇷": "9083", "🇪🇸": "9092",
		"🇳🇱": "9094", "🇮🇹": "9095", "🇸🇪": "9097", "🇬🇧": "9116",
	} {
		ep := map[string]any{
			"dest": "proxy.example.com", "port": 2083, "forceTls": "same",
			"remark": flag, "vlessRoute": route, "isHost": true,
		}
		original := inbound.Settings
		link := NewSubService("{{INBOUND}}-{{EMAIL}}").linkFromHosts(inbound, client, []map[string]any{ep})
		u, err := url.Parse(link)
		if err != nil {
			t.Fatal(err)
		}
		if u.Fragment != "farnaz_skm"+flag {
			t.Fatalf("flag=%s: fragment=%q", flag, u.Fragment)
		}
		if u.User == nil { t.Fatalf("missing VLESS credentials: %q", link) }
		if u.Port() != strconv.Itoa(2083) || u.User.Username() != applyVlessRoute("11111111-2222-4333-8444-555555555555", route) {
			t.Fatalf("flag=%s: port or route changed", flag)
		}
		if inbound.Settings != original || ep["remark"] != flag {
			t.Fatal("rendering mutated shared input")
		}
	}
}

func TestBotServiceHostNameByPort(t *testing.T) {
	for _, email := range []string{"farnaz_skm-direct", "farnaz_skm-tunnel"} {
		for _, tc := range []struct {
			port int
			want string
		}{
			{8080, "farnaz_skm-direct"},
			{443, "farnaz_skm-tunnel"},
			{2083, "farnaz_skm🇺🇸"},
		} {
			client := model.Client{Email: email, SubID: "tg-123-test"}
			inbound := &model.Inbound{Port: tc.port, Protocol: model.VLESS}
			got, ok := botServiceHostName(inbound, client, "🇺🇸")
			if !ok || got != tc.want {
				t.Fatalf("email=%s port=%d: got %q, %v", email, tc.port, got, ok)
			}
		}
	}
}

func TestBotServiceHostNameExcludesUnrelatedAccounts(t *testing.T) {
	for _, client := range []model.Client{
		{Email: "manual-direct", SubID: "manual-sub"},
		{Email: "free-test-tunnel", SubID: "tg-123-free"},
		{Email: "channel-free-test-tunnel", SubID: "tg-123-free"},
		{Email: "giftly-test-direct", SubID: "tg-123-gift"},
	} {
		if _, ok := botServiceHostName(&model.Inbound{Port: 443, Protocol: model.VLESS}, client, "🇺🇸"); ok {
			t.Fatalf("unexpected override for %s", client.Email)
		}
	}
}
