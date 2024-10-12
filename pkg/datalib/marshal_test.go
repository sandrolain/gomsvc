package datalib

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
)

func TestMarshalBody(t *testing.T) {
	data := struct {
		Foo string
		Bar int
	}{
		Foo: "foo",
		Bar: 123,
	}

	tests := []struct {
		name    string
		typ     string
		wantErr bool
	}{
		{
			name:    "json",
			typ:     TypeJson,
			wantErr: false,
		},
		{
			name:    "msgpack",
			typ:     TypeMsgpack,
			wantErr: false,
		},
		{
			name:    "x-msgpack",
			typ:     TypeXMsgpack,
			wantErr: false,
		},
		{
			name:    "unknown type",
			typ:     "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBytes, err := MarshalBody(tt.typ, &data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotEmpty(t, reqBytes)
		})
	}
}

type foo struct {
	Foo string
	Bar int
}

func TestUnmarshalJSONBody(t *testing.T) {
	data := []byte(`{"foo":"foo","bar":123}`)

	tests := []struct {
		name    string
		typ     string
		wantErr bool
	}{
		{
			name:    "json",
			typ:     TypeJson,
			wantErr: false,
		},
		{
			name:    "msgpack",
			typ:     TypeMsgpack,
			wantErr: true,
		},
		{
			name:    "x-msgpack",
			typ:     TypeXMsgpack,
			wantErr: true,
		},
		{
			name:    "unknown type",
			typ:     "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := UnmarshalBody[foo](tt.typ, data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, "foo", dst.Foo)
			require.Equal(t, 123, dst.Bar)
		})
	}
}

func TestUnmarshalMsgpackBody(t *testing.T) {
	data, err := msgpack.Marshal(&foo{
		Foo: "foo",
		Bar: 123,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		typ     string
		wantErr bool
	}{
		{
			name:    "json",
			typ:     TypeJson,
			wantErr: true,
		},
		{
			name:    "msgpack",
			typ:     TypeMsgpack,
			wantErr: false,
		},
		{
			name:    "x-msgpack",
			typ:     TypeXMsgpack,
			wantErr: false,
		},
		{
			name:    "unknown type",
			typ:     "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst, err := UnmarshalBody[foo](tt.typ, data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, "foo", dst.Foo)
			require.Equal(t, 123, dst.Bar)
		})
	}
}
