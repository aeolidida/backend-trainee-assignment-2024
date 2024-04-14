package nullable

import (
	"bytes"
	"reflect"
	"testing"
)

func TestNullString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		n       NullString
		want    []byte
		wantErr bool
	}{
		{"valid", NullString{Valid: true, String: "hello"}, []byte(`"hello"`), false},
		{"null", NullString{Valid: false}, []byte("null"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("NullString.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NullString.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    NullString
		wantErr bool
	}{
		{"valid", []byte(`"hello"`), NullString{Valid: true, String: "hello"}, false},
		{"null", []byte("null"), NullString{}, false},
		{"invalid", []byte("invalid"), NullString{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullString
			if err := n.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("NullString.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("NullString.UnmarshalJSON() = %v, want %v", n, tt.want)
			}
		})
	}
}

func TestNullInt64_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		n       NullInt64
		want    []byte
		wantErr bool
	}{
		{"valid", NullInt64{Valid: true, Int64: 42}, []byte("42"), false},
		{"null", NullInt64{Valid: false}, []byte("null"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("NullInt64.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NullInt64.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullInt64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    NullInt64
		wantErr bool
	}{
		{"valid", []byte("42"), NullInt64{Valid: true, Int64: 42}, false},
		{"null", []byte("null"), NullInt64{}, false},
		{"invalid", []byte("invalid"), NullInt64{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullInt64
			if err := n.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("NullInt64.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("NullInt64.UnmarshalJSON() = %v, want %v", n, tt.want)
			}
		})
	}
}

func TestNullUint64_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		n       NullUint64
		want    []byte
		wantErr bool
	}{
		{"valid", NullUint64{Valid: true, Uint64: 42}, []byte("42"), false},
		{"null", NullUint64{Valid: false}, []byte("null"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("NullUint64.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NullUint64.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullUint64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    NullUint64
		wantErr bool
	}{
		{"valid", []byte("42"), NullUint64{Valid: true, Uint64: 42}, false},
		{"null", []byte("null"), NullUint64{}, false},
		{"invalid", []byte("invalid"), NullUint64{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullUint64
			if err := n.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("NullUint64.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("NullUint64.UnmarshalJSON() = %v, want %v", n, tt.want)
			}
		})
	}
}

func TestNullBool_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		n       NullBool
		want    []byte
		wantErr bool
	}{
		{"true", NullBool{Valid: true, Bool: true}, []byte("true"), false},
		{"false", NullBool{Valid: true, Bool: false}, []byte("false"), false},
		{"null", NullBool{Valid: false}, []byte("null"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.n.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("NullBool.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("NullBool.MarshalJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNullBool_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    NullBool
		wantErr bool
	}{
		{"true", []byte("true"), NullBool{Valid: true, Bool: true}, false},
		{"false", []byte("false"), NullBool{Valid: true, Bool: false}, false},
		{"null", []byte("null"), NullBool{}, false},
		{"invalid", []byte("invalid"), NullBool{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var n NullBool
			if err := n.UnmarshalJSON(tt.data); (err != nil) != tt.wantErr {
				t.Errorf("NullBool.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(n, tt.want) {
				t.Errorf("NullBool.UnmarshalJSON() = %v, want %v", n, tt.want)
			}
		})
	}
}
