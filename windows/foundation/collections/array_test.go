package collections

import (
	"testing"

	"github.com/saltosystems/winrt-go"
	"github.com/stretchr/testify/require"
)

func Test_GetCurrent(t *testing.T) {
	a := NewArrayIterable([]any{1, 2, 3}, winrt.SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)

	var ok bool = false
	for ok, err = it.MoveNext(); err == nil && ok; ok, err = it.MoveNext() {
		b, err := it.GetHasCurrent()
		require.NoError(t, err)
		require.True(t, b)

		ptr, err := it.GetCurrent()
		require.NoError(t, err)

		println(int(uintptr(ptr)))
	}
}

func Test_GetMany(t *testing.T) {
	a := NewArrayIterable([]any{101, 202, 303}, winrt.SignatureInt32)

	it, err := a.First()
	require.NoError(t, err)
	resp, n, err := it.GetMany(12)
	require.NoError(t, err)

	println("RESP", n, resp, len(resp))

	for i := 0; i < len(resp); i++ {
		println(int(uintptr(resp[i])))
	}
}
