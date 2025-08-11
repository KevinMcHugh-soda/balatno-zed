package game

import (
	"bufio"
	"strings"
	"testing"
)

func TestGetShopActionParsesBuyWithQuantity(t *testing.T) {
	handler := &LoggerEventHandler{scanner: bufio.NewScanner(strings.NewReader("buy 12\n"))}
	action, params, quit := handler.GetShopAction()
	if quit {
		t.Fatalf("expected quit to be false, got true")
	}
	if action != PlayerActionBuy {
		t.Fatalf("expected action %s, got %s", PlayerActionBuy, action)
	}
	if len(params) != 1 || params[0] != "12" {
		t.Fatalf("expected params [\"12\"], got %v", params)
	}
}

func TestGetShopActionParsesReroll(t *testing.T) {
	handler := &LoggerEventHandler{scanner: bufio.NewScanner(strings.NewReader("reroll\n"))}
	action, params, quit := handler.GetShopAction()
	if quit {
		t.Fatalf("expected quit to be false, got true")
	}
	if action != PlayerActionReroll {
		t.Fatalf("expected action %s, got %s", PlayerActionReroll, action)
	}
	if len(params) != 0 {
		t.Fatalf("expected no params, got %v", params)
	}
}
