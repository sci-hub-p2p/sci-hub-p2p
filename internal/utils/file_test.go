// Copyright 2021 Trim21<trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/constants"
)

func TestFileExist(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	tmpFile := filepath.Join(tmp, "filename")
	assert.Nil(t, os.WriteFile(tmpFile, []byte("s"), constants.DefaultFilePerm))
	re, err := utils.FileExist(tmpFile)
	assert.Nil(t, err)
	assert.True(t, re, tmpFile)
}

func TestFileExistDirErr(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	tmpDir := filepath.Join(tmp, "dirname")
	assert.Nil(t, os.MkdirAll(tmpDir, os.ModeDir))
	_, err := utils.FileExist(tmpDir)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, utils.ErrNotAFile)
}

func TestDirExist(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	tmpDir := filepath.Join(tmp, "dirname")
	assert.Nil(t, os.MkdirAll(tmpDir, os.ModeDir))
	re, err := utils.DirExist(tmpDir)
	assert.Nil(t, err)
	assert.True(t, re)
}

func TestDirIsFileErr(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	tmpFile := filepath.Join(tmp, "filename")
	assert.Nil(t, os.WriteFile(tmpFile, []byte("s"), constants.DefaultFilePerm))
	_, err := utils.DirExist(tmpFile)
	assert.NotNil(t, err)
	assert.ErrorIs(t, err, utils.ErrNotADir)
}
