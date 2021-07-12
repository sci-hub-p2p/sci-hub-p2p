package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertSlice(t *testing.T) {
	type T1 struct {
		Host  string   `tuple:"0"`
		Port  int      `tuple:"1"`
		Names []string `tuple:"3"`
		Raw   []byte   `tuple:"2"`
	}

	var row = []interface{}{"233", 66, []byte("string"), []string{"1", "2", "3"}}
	var v T1
	err := ScanSlice(row, &v)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, row[0], v.Host)
	assert.Equal(t, row[1], v.Port)
	assert.Equal(t, row[3], v.Names)
	assert.Equal(t, row[2], v.Raw)
}

func TestDontPanic(t *testing.T) {
	type T1 struct {
		Host  string   `tuple:"0"`
		Port  int      `tuple:"1"`
		Names []string `tuple:"3"`
		Raw   []byte   `tuple:"2"`
	}

	var row = []interface{}{"1"}
	var v T1
	err := ScanSlice(row, &v)
	assert.NotNil(t, err)
}
