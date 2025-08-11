package game

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

// newHandlerWithInput creates a LoggerEventHandler with a custom input scanner.
// The LoggerEventHandler's scanner field is unexported, so reflection is used to
// inject our own scanner for testing.
func newHandlerWithInput(input string) *LoggerEventHandler {
	h := NewLoggerEventHandler()
	scanner := bufio.NewScanner(strings.NewReader(input))
	v := reflect.ValueOf(h).Elem().FieldByName("scanner")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(scanner))
	return h
}

func TestGetShopActionParsesBuyWithQuantity(t *testing.T) {
	handler := NewLoggerEventHandlerFromReader(strings.NewReader("buy 12\n"))
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
	handler := NewLoggerEventHandlerFromReader(strings.NewReader("reroll\n"))
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
