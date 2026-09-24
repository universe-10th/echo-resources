package services

import "testing"

type testPathInt int16
type testPathUint uint32
type testPathString string

func TestParsePathParamParsesSupportedTypes(t *testing.T) {
	t.Parallel()

	stringValue, err := ParsePathParam[string]("user-1")
	if err != nil {
		t.Fatalf("ParsePathParam[string] returned error: %v", err)
	}
	if stringValue != "user-1" {
		t.Fatalf("unexpected string value: %q", stringValue)
	}

	intValue, err := ParsePathParam[int8]("-12")
	if err != nil {
		t.Fatalf("ParsePathParam[int8] returned error: %v", err)
	}
	if intValue != -12 {
		t.Fatalf("unexpected int8 value: %d", intValue)
	}

	uintValue, err := ParsePathParam[uint16]("42")
	if err != nil {
		t.Fatalf("ParsePathParam[uint16] returned error: %v", err)
	}
	if uintValue != 42 {
		t.Fatalf("unexpected uint16 value: %d", uintValue)
	}
}

func TestParsePathParamParsesDefinedTypes(t *testing.T) {
	t.Parallel()

	stringValue, err := ParsePathParam[testPathString]("account-1")
	if err != nil {
		t.Fatalf("ParsePathParam[testPathString] returned error: %v", err)
	}
	if stringValue != "account-1" {
		t.Fatalf("unexpected defined string value: %q", stringValue)
	}

	intValue, err := ParsePathParam[testPathInt]("-123")
	if err != nil {
		t.Fatalf("ParsePathParam[testPathInt] returned error: %v", err)
	}
	if intValue != -123 {
		t.Fatalf("unexpected defined int value: %d", intValue)
	}

	uintValue, err := ParsePathParam[testPathUint]("123")
	if err != nil {
		t.Fatalf("ParsePathParam[testPathUint] returned error: %v", err)
	}
	if uintValue != 123 {
		t.Fatalf("unexpected defined uint value: %d", uintValue)
	}
}

func TestParsePathParamRejectsInvalidNumbers(t *testing.T) {
	t.Parallel()

	if _, err := ParsePathParam[int]("not-a-number"); err == nil {
		t.Fatal("expected invalid signed integer to return an error")
	}
	if _, err := ParsePathParam[uint]("-1"); err == nil {
		t.Fatal("expected negative unsigned integer to return an error")
	}
	if _, err := ParsePathParam[int8]("128"); err == nil {
		t.Fatal("expected out-of-range signed integer to return an error")
	}
	if _, err := ParsePathParam[uint8]("256"); err == nil {
		t.Fatal("expected out-of-range unsigned integer to return an error")
	}
}
