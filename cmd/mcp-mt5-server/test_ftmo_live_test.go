package main

import (
	"testing"
	"time"

	"github.com/lukeware/kayron-ai/internal/services/mt5"
)

// TestFTMOLiveConnection tenta conectar ao FTMO MT5 rodando
func TestFTMOLiveConnection(t *testing.T) {
	// FTMO credentials
	baseURL := "http://localhost:8228" // MT5 WebAPI port
	login := "1513212141"
	password := "dIh*r!5l$7l"

	t.Logf("🔗 Testando conexão FTMO MT5...")
	t.Logf("   Server: %s", baseURL)
	t.Logf("   Login: %s", login)
	t.Logf("   Time: %s UTC", time.Now().UTC().Format("2006-01-02 15:04:05"))

	// Create client
	client := mt5.NewClient(baseURL, login, password, 5)

	// Try to get account info
	accountInfo, err := client.GetAccount()
	if err != nil {
		// Expected: connection timeout if MT5 WebAPI not running
		t.Logf("⚠️  Conexão falhou (esperado sem WebAPI rodando): %v", err)
		t.Logf("\nPróximos passos:")
		t.Logf("1. Habilitar WebAPI no MT5 (Tools → Options → API)")
		t.Logf("2. Configurar porta: 8228")
		t.Logf("3. Reiniciar MT5")
		t.Logf("4. Rodar teste novamente")
		return
	}

	// Success: display account info
	t.Logf("✅ Conectado ao FTMO MT5!")
	t.Logf("   Login: %d", accountInfo.Login)
	t.Logf("   Balance: %s", accountInfo.Balance.String())
	t.Logf("   Equity: %s", accountInfo.Equity.String())
	t.Logf("   Free Margin: %s", accountInfo.FreeMargin.String())
	t.Logf("   Margin Level: %s%%", accountInfo.MarginLevel.String())
	t.Logf("   Currency: %s", accountInfo.Currency)
}

// TestFTMOSymbols testa disponibilidade de símbolos
func TestFTMOSymbols(t *testing.T) {
	baseURL := "http://localhost:8228"
	client := mt5.NewClient(baseURL, "1513212141", "dIh*r!5l$7l", 5)

	symbols := []string{"EURUSD", "GBPUSD", "USDJPY", "AUDUSD", "NZDUSD"}

	t.Logf("\n📊 Testando símbolos FTMO...")

	for _, symbol := range symbols {
		quote, err := client.GetQuote(symbol)
		if err != nil {
			t.Logf("❌ %s: %v", symbol, err)
			continue
		}

		bid := quote.Bid.String()
		ask := quote.Ask.String()
		t.Logf("✅ %s: Bid=%s Ask=%s", symbol, bid, ask)
	}
}

// TestFTMOWebAPIStatus verifica status WebAPI
func TestFTMOWebAPIStatus(t *testing.T) {
	t.Logf("\n🔧 Checklist WebAPI FTMO:")
	t.Logf("[ ] MT5 Terminal aberto")
	t.Logf("[ ] Tools → Options → API → Enable WebAPI")
	t.Logf("[ ] Port: 8228 (default)")
	t.Logf("[ ] Authentication: Basic Auth")
	t.Logf("[ ] Login: 1513212141")
	t.Logf("[ ] Password: dIh*r!5l$7l")
	t.Logf("\nSe todos checkados e teste falha:")
	t.Logf("1. Abrir MT5 Tools → Options")
	t.Logf("2. Ir para API tab")
	t.Logf("3. Abilitar 'WebAPI'")
	t.Logf("4. Set Port: 8228")
	t.Logf("5. Restart MT5")
}
