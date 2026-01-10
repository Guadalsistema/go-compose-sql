package typeconv

import (
	"database/sql"
	"reflect"
	"testing"
	"time"
)

func TestRegistry_Convert(t *testing.T) {
	tests := []struct {
		name        string
		setupFn     func(r *Registry)
		source      interface{}
		targetType  reflect.Type
		wantErr     bool
		validateFn  func(t *testing.T, result interface{})
	}{
		{
			name: "direct type match - no conversion",
			setupFn: func(r *Registry) {
				// No converters registered
			},
			source:     "hello",
			targetType: reflect.TypeOf(""),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				if result != "hello" {
					t.Errorf("expected 'hello', got %v", result)
				}
			},
		},
		{
			name: "string to time.Time conversion",
			setupFn: func(r *Registry) {
				r.Register(
					reflect.TypeOf(""),
					reflect.TypeOf(time.Time{}),
					StringToTime,
				)
			},
			source:     "2024-01-15 10:30:00",
			targetType: reflect.TypeOf(time.Time{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				tm, ok := result.(time.Time)
				if !ok {
					t.Errorf("expected time.Time, got %T", result)
					return
				}
				if tm.Year() != 2024 || tm.Month() != 1 || tm.Day() != 15 {
					t.Errorf("unexpected time value: %v", tm)
				}
			},
		},
		{
			name: "int64 to time.Time conversion",
			setupFn: func(r *Registry) {
				r.Register(
					reflect.TypeOf(int64(0)),
					reflect.TypeOf(time.Time{}),
					Int64ToTime,
				)
			},
			source:     int64(1705318200),
			targetType: reflect.TypeOf(time.Time{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				_, ok := result.(time.Time)
				if !ok {
					t.Errorf("expected time.Time, got %T", result)
				}
			},
		},
		{
			name: "default converter - string to time.Time",
			setupFn: func(r *Registry) {
				r.RegisterDefault(
					reflect.TypeOf(time.Time{}),
					DefaultTimeConverter,
				)
			},
			source:     "2024-01-15 10:30:00",
			targetType: reflect.TypeOf(time.Time{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				_, ok := result.(time.Time)
				if !ok {
					t.Errorf("expected time.Time, got %T", result)
				}
			},
		},
		{
			name: "default converter - int64 to time.Time",
			setupFn: func(r *Registry) {
				r.RegisterDefault(
					reflect.TypeOf(time.Time{}),
					DefaultTimeConverter,
				)
			},
			source:     int64(1705318200),
			targetType: reflect.TypeOf(time.Time{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				_, ok := result.(time.Time)
				if !ok {
					t.Errorf("expected time.Time, got %T", result)
				}
			},
		},
		{
			name: "string to sql.NullTime conversion",
			setupFn: func(r *Registry) {
				r.Register(
					reflect.TypeOf(""),
					reflect.TypeOf(sql.NullTime{}),
					StringToNullTime,
				)
			},
			source:     "2024-01-15 10:30:00",
			targetType: reflect.TypeOf(sql.NullTime{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				nt, ok := result.(sql.NullTime)
				if !ok {
					t.Errorf("expected sql.NullTime, got %T", result)
					return
				}
				if !nt.Valid {
					t.Errorf("expected Valid=true, got false")
				}
			},
		},
		{
			name: "nil to sql.NullTime conversion",
			setupFn: func(r *Registry) {
				r.RegisterDefault(
					reflect.TypeOf(sql.NullTime{}),
					DefaultNullTimeConverter,
				)
			},
			source:     nil,
			targetType: reflect.TypeOf(sql.NullTime{}),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				nt, ok := result.(sql.NullTime)
				if !ok {
					t.Errorf("expected sql.NullTime, got %T", result)
					return
				}
				if nt.Valid {
					t.Errorf("expected Valid=false, got true")
				}
			},
		},
		{
			name: "int64 to bool conversion",
			setupFn: func(r *Registry) {
				r.Register(
					reflect.TypeOf(int64(0)),
					reflect.TypeOf(true),
					Int64ToBool,
				)
			},
			source:     int64(1),
			targetType: reflect.TypeOf(true),
			wantErr:    false,
			validateFn: func(t *testing.T, result interface{}) {
				b, ok := result.(bool)
				if !ok {
					t.Errorf("expected bool, got %T", result)
					return
				}
				if !b {
					t.Errorf("expected true, got false")
				}
			},
		},
		{
			name: "no converter registered - error",
			setupFn: func(r *Registry) {
				// No converters registered
			},
			source:     "hello",
			targetType: reflect.TypeOf(int64(0)),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setupFn(r)

			result, err := r.Convert(tt.source, tt.targetType)

			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateFn != nil {
				tt.validateFn(t, result)
			}
		})
	}
}

func TestRegistry_NeedsConversion(t *testing.T) {
	tests := []struct {
		name       string
		setupFn    func(r *Registry)
		sourceType reflect.Type
		targetType reflect.Type
		want       bool
	}{
		{
			name: "same type - no conversion needed",
			setupFn: func(r *Registry) {
				// No converters needed
			},
			sourceType: reflect.TypeOf(""),
			targetType: reflect.TypeOf(""),
			want:       false,
		},
		{
			name: "specific converter registered",
			setupFn: func(r *Registry) {
				r.Register(
					reflect.TypeOf(""),
					reflect.TypeOf(time.Time{}),
					StringToTime,
				)
			},
			sourceType: reflect.TypeOf(""),
			targetType: reflect.TypeOf(time.Time{}),
			want:       true,
		},
		{
			name: "default converter registered",
			setupFn: func(r *Registry) {
				r.RegisterDefault(
					reflect.TypeOf(time.Time{}),
					DefaultTimeConverter,
				)
			},
			sourceType: reflect.TypeOf(""),
			targetType: reflect.TypeOf(time.Time{}),
			want:       true,
		},
		{
			name: "no converter registered",
			setupFn: func(r *Registry) {
				// No converters
			},
			sourceType: reflect.TypeOf(""),
			targetType: reflect.TypeOf(int64(0)),
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setupFn(r)

			got := r.NeedsConversion(tt.sourceType, tt.targetType)
			if got != tt.want {
				t.Errorf("NeedsConversion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertingScanner(t *testing.T) {
	r := NewRegistry()
	r.Register(
		reflect.TypeOf(""),
		reflect.TypeOf(time.Time{}),
		StringToTime,
	)

	scanner := r.CreateScanner(reflect.TypeOf(time.Time{}))

	// Simulate scanning a string value (like SQLite would return)
	err := scanner.Scan("2024-01-15 10:30:00")
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Extract the result
	cs, ok := scanner.(*convertingScanner)
	if !ok {
		t.Fatalf("expected *convertingScanner, got %T", scanner)
	}

	result := cs.Result()
	tm, ok := result.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T", result)
	}

	if tm.Year() != 2024 {
		t.Errorf("unexpected year: %d", tm.Year())
	}
}
