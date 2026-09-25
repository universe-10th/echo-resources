package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type testPathInt int16
type testPathUint uint32
type testPathString string
type testPathBool bool
type testUnsupportedPathParam struct {
	value string
}

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

	boolValue, err := ParsePathParam[bool]("true")
	if err != nil {
		t.Fatalf("ParsePathParam[bool] returned error: %v", err)
	}
	if !boolValue {
		t.Fatal("unexpected bool value: false")
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

	boolValue, err := ParsePathParam[testPathBool]("true")
	if err != nil {
		t.Fatalf("ParsePathParam[testPathBool] returned error: %v", err)
	}
	if !boolValue {
		t.Fatal("unexpected defined bool value: false")
	}
}

func TestParsePathParamParsesTextUnmarshalers(t *testing.T) {
	t.Parallel()

	timestamp := time.Date(2026, 9, 25, 10, 30, 0, 0, time.UTC)
	timeValue, err := ParsePathParam[time.Time](timestamp.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("ParsePathParam[time.Time] returned error: %v", err)
	}
	if !timeValue.Equal(timestamp) {
		t.Fatalf("unexpected time value: %s", timeValue.Format(time.RFC3339))
	}

	uuidValue := uuid.New()
	parsedUUID, err := ParsePathParam[uuid.UUID](uuidValue.String())
	if err != nil {
		t.Fatalf("ParsePathParam[uuid.UUID] returned error: %v", err)
	}
	if parsedUUID != uuidValue {
		t.Fatalf("unexpected UUID value: %s", parsedUUID)
	}

	objectID := bson.NewObjectID()
	parsedObjectID, err := ParsePathParam[bson.ObjectID](objectID.Hex())
	if err != nil {
		t.Fatalf("ParsePathParam[bson.ObjectID] returned error: %v", err)
	}
	if parsedObjectID != objectID {
		t.Fatalf("unexpected ObjectID value: %s", parsedObjectID.Hex())
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

func TestParsePathParamRejectsInvalidTextUnmarshalerValue(t *testing.T) {
	t.Parallel()

	if _, err := ParsePathParam[uuid.UUID]("not-a-uuid"); err == nil {
		t.Fatal("expected invalid UUID to return an error")
	}
}

func TestParsePathParamRejectsUnsupportedComparableTypes(t *testing.T) {
	t.Parallel()

	if _, err := ParsePathParam[testUnsupportedPathParam]("value"); err == nil {
		t.Fatal("expected unsupported comparable type to return an error")
	}
}
