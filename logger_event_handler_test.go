package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	game "balatno/internal/game"
)

// newHandlerWithInput creates a LoggerEventHandler with a custom input scanner.
// The LoggerEventHandler's scanner field is unexported, so reflection is used to
// inject our own scanner for testing.
func newHandlerWithInput(input string) *game.LoggerEventHandler {
	h := game.NewLoggerEventHandler()
	scanner := bufio.NewScanner(strings.NewReader(input))
	v := reflect.ValueOf(h).Elem().FieldByName("scanner")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(scanner))
	return h
}

func TestGetShopActionParsesBuyWithQuantity(t *testing.T) {
	handler := game.NewLoggerEventHandlerFromReader(strings.NewReader("buy 12\n"))
	action, params, quit := handler.GetShopAction()
	if quit {
		t.Fatalf("expected quit to be false, got true")
	}
	if action != game.PlayerActionBuy {
		t.Fatalf("expected action %s, got %s", game.PlayerActionBuy, action)
	}
	if len(params) != 1 || params[0] != "12" {
		t.Fatalf("expected params [\"12\"], got %v", params)
	}
}

func TestGetShopActionParsesReroll(t *testing.T) {
	handler := game.NewLoggerEventHandlerFromReader(strings.NewReader("reroll\n"))
	action, params, quit := handler.GetShopAction()
	if quit {
		t.Fatalf("expected quit to be false, got true")
	}
	if action != game.PlayerActionReroll {
		t.Fatalf("expected action %s, got %s", game.PlayerActionReroll, action)
	}
	if len(params) != 0 {
		t.Fatalf("expected no params, got %v", params)
	}
}
