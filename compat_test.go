package stash

import (
	"testing"
	"time"
)

func TestUntypedBasic(t *testing.T) {
	uc := NewUntyped(5*time.Minute, 0)
	uc.Set("foo", "bar", DefaultTTL)
	v, ok := uc.Get("foo")
	if !ok || v != "bar" {
		t.Fatalf("expected bar, got %v", v)
	}
}

func TestUntypedNoExpiration(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("key", 42, 0)
	v, ok := uc.Get("key")
	if !ok || v != 42 {
		t.Fatal("expected 42")
	}
}

func TestUntypedExpiration(t *testing.T) {
	uc := NewUntyped(50*time.Millisecond, 0)
	uc.SetDefault("key", "val")
	time.Sleep(60 * time.Millisecond)
	_, ok := uc.Get("key")
	if ok {
		t.Fatal("expected expired")
	}
}

func TestIncrementInt(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", int(10), 0)
	err := uc.Increment("c", 5)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(int) != 15 {
		t.Fatalf("expected 15, got %v", v)
	}
}

func TestIncrementInt64(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", int64(100), 0)
	err := uc.Increment("c", 50)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(int64) != 150 {
		t.Fatalf("expected 150, got %v", v)
	}
}

func TestIncrementFloat64(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", float64(1.5), 0)
	err := uc.Increment("c", 2)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(float64) != 3.5 {
		t.Fatalf("expected 3.5, got %v", v)
	}
}

func TestIncrementUint(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", uint(5), 0)
	err := uc.Increment("c", 3)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(uint) != 8 {
		t.Fatalf("expected 8, got %v", v)
	}
}

func TestDecrementInt(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", int(10), 0)
	err := uc.Decrement("c", 3)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(int) != 7 {
		t.Fatalf("expected 7, got %v", v)
	}
}

func TestDecrementFloat32(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", float32(10.0), 0)
	err := uc.Decrement("c", 3)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := uc.Get("c")
	if v.(float32) != 7.0 {
		t.Fatalf("expected 7.0, got %v", v)
	}
}

func TestIncrementNonNumeric(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", "not a number", 0)
	err := uc.Increment("c", 1)
	if err == nil {
		t.Fatal("expected error for non-numeric increment")
	}
}

func TestDecrementNonNumeric(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", "not a number", 0)
	err := uc.Decrement("c", 1)
	if err == nil {
		t.Fatal("expected error for non-numeric decrement")
	}
}

func TestIncrementMissing(t *testing.T) {
	uc := NewUntyped(0, 0)
	err := uc.Increment("nope", 1)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestDecrementMissing(t *testing.T) {
	uc := NewUntyped(0, 0)
	err := uc.Decrement("nope", 1)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestIncrementExpired(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", int(10), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	err := uc.Increment("c", 1)
	if err == nil {
		t.Fatal("expected error for expired key")
	}
}

func TestIncrementAllIntTypes(t *testing.T) {
	uc := NewUntyped(0, 0)
	tests := []struct {
		name string
		val  any
	}{
		{"int8", int8(1)},
		{"int16", int16(1)},
		{"int32", int32(1)},
		{"uint8", uint8(1)},
		{"uint16", uint16(1)},
		{"uint32", uint32(1)},
		{"uint64", uint64(1)},
	}
	for _, tt := range tests {
		uc.Set(tt.name, tt.val, 0)
		if err := uc.Increment(tt.name, 1); err != nil {
			t.Fatalf("Increment(%s): %v", tt.name, err)
		}
	}
}

func TestUntypedDeleteAndFlush(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("a", 1, 0)
	uc.Set("b", 2, 0)
	uc.Delete("a")
	if _, ok := uc.Get("a"); ok {
		t.Fatal("expected a deleted")
	}
	uc.Flush()
	if uc.Count() != 0 {
		t.Fatal("expected empty after flush")
	}
}

func TestUntypedWithCleanup(t *testing.T) {
	uc := NewUntyped(50*time.Millisecond, 25*time.Millisecond)
	defer uc.Stop()
	uc.SetDefault("key", "val")
	time.Sleep(80 * time.Millisecond)
	if uc.Count() != 0 {
		t.Fatal("expected janitor to clean up")
	}
}

func TestDecrementAllTypes(t *testing.T) {
	uc := NewUntyped(0, 0)
	tests := []struct {
		name string
		val  any
	}{
		{"int", int(10)},
		{"int8", int8(10)},
		{"int16", int16(10)},
		{"int32", int32(10)},
		{"int64", int64(10)},
		{"uint", uint(10)},
		{"uint8", uint8(10)},
		{"uint16", uint16(10)},
		{"uint32", uint32(10)},
		{"uint64", uint64(10)},
		{"float32", float32(10)},
		{"float64", float64(10)},
	}
	for _, tt := range tests {
		uc.Set(tt.name, tt.val, 0)
		if err := uc.Decrement(tt.name, 1); err != nil {
			t.Fatalf("Decrement(%s): %v", tt.name, err)
		}
	}
}

func TestDecrementExpired(t *testing.T) {
	uc := NewUntyped(0, 0)
	uc.Set("c", int(10), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	err := uc.Decrement("c", 1)
	if err == nil {
		t.Fatal("expected error for expired key")
	}
}
