package pkg

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func Test_initExifTool(t *testing.T) {
	o, err := GetNewOperator()
	require.NoError(t, err)
	require.NotNil(t, o.Storage.Exif)
}

func Test_getFileDate(t *testing.T) {
	o, err := GetNewOperator()
	require.NoError(t, err)

	// injected a date (for testing purposes) into this file with following call:
	// exiftool -CreateDate="2022:12:08 19:09:53" pkg/evil_gopher.png
	// evil_gopher.png belongs to https://github.com/MariaLetta/free-gophers-pack/blob/master/characters/png/1.png
	{
		res, err := o.getFileDate("./evil_gopher.png", "month")
		require.NoError(t, err)
		assert.Equal(t, "2022/12", res)
	}
	{
		res, err := o.getFileDate("./evil_gopher.png", "year")
		require.NoError(t, err)
		assert.Equal(t, "2022", res)
	}

	f, err := os.Create("random.png")
	require.NoError(t, err)
	_, err = o.getFileDate(f.Name(), "year")
	require.Equal(t, err, ErrNoCreateDate)
	require.NoError(t, os.Remove(f.Name()))
}
