package money

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("should create money with valid cents", func(t *testing.T) {
		money, err := New(1000)
		require.NoError(t, err)
		require.Equal(t, Money{Cents: 1000}, money)
	})

	t.Run("should return error for negative cents", func(t *testing.T) {
		_, err := New(-100)
		require.ErrorIs(t, err, ErrMoneyNegative)
	})

	t.Run("should create money with zero cents", func(t *testing.T) {
		money := Money{}
		require.Equal(t, Money{Cents: 0}, money)
	})
}

func TestAdd(t *testing.T) {
	t.Parallel()

	t.Run("should add two money values", func(t *testing.T) {
		money1, _ := New(1500)
		money2, _ := New(2500)
		result := money1.Add(money2)
		require.Equal(t, Money{Cents: 4000}, result)
	})
}

func TestMoneyMarshalAndUnmarshal(t *testing.T) {
	t.Parallel()

	t.Run("should marshal correctly", func(t *testing.T) {
		m := Money{Cents: 999}

		data, err := json.Marshal(m)

		require.NoError(t, err)
		require.Equal(t, "999", string(data))
	})

	t.Run("should unmarshal", func(t *testing.T) {
		var m Money

		t.Run("should unmarshal correctly", func(t *testing.T) {
			err := json.Unmarshal([]byte("12345"), &m)

			require.NoError(t, err)
			require.Equal(t, int64(12345), m.Cents)
		})

		t.Run("should unmarshal zero", func(t *testing.T) {
			err := json.Unmarshal([]byte("0"), &m)

			require.NoError(t, err)
			require.Equal(t, int64(0), m.Cents)
		})

		t.Run("should unmarshal negative", func(t *testing.T) {
			err := json.Unmarshal([]byte("-100"), &m)

			require.Error(t, err)
			require.Equal(t, ErrMoneyNegative, err)
		})

		t.Run("should unmarshal invalid type", func(t *testing.T) {
			err := json.Unmarshal([]byte(`"abc"`), &m)

			require.Error(t, err)
		})
	})

}
