package service

import (
	"archive/zip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"imgcutter/imgprocessing"

	"github.com/magiconair/properties/assert"
)

func Test_FileManager(t *testing.T) {
	fm := &fileManager{
		sessionsMapMutex: sync.Mutex{},
		sessions:         map[string]*Session{},
	}

	fm.RemoveAll()

	testSession1 := fm.New()
	testSession2 := fm.New()
	testSession3 := fm.New()
	testSessiontoDelete := fm.New()
	deletingSessionID := testSessiontoDelete.String()

	t.Run("check four session created, one deleted", func(t *testing.T) {
		s, ok := fm.Find(deletingSessionID)
		assert.Equal(t, ok, true)
		err := fm.TerminateSession(s)
		assert.Equal(t, err, nil)
		_, ok = fm.Find(deletingSessionID)
		assert.Equal(t, ok, false)
		assert.Equal(t, len(fm.sessions), 3)
	})

	testfile, err := os.Open("mem.jpg")
	assert.Equal(t, err, nil)
	defer testfile.Close()

	t.Run("uploading files", func(t *testing.T) {
		err = fm.UploadFile(testSession1, testfile, "testfile1.jpg")
		assert.Equal(t, err, nil)

		testfile.Seek(0, 0)
		err = fm.UploadFile(testSession2, testfile, "testfile2.jpg")
		assert.Equal(t, err, nil)

		testfile.Seek(0, 0)
		err = fm.UploadFile(testSession3, testfile, "testfile3.jpg")
		assert.Equal(t, err, nil)
	})
	defer fm.RemoveAll()

	counter := 0
	walkFunc := func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			counter++
		}
		return nil
	}

	t.Run("three files uploaded", func(t *testing.T) {

		filepath.WalkDir("temp", walkFunc)
		assert.Equal(t, counter, 3)
	})

	t.Run("cutting files", func(t *testing.T) {
		err = fm.CutFile(testSession1, fmt.Sprintf("temp/%s/testfile1.jpg", testSession1.String()), 32, 32)
		assert.Equal(t, err, nil)
		err = fm.CutFile(testSession2, fmt.Sprintf("temp/%s/testfile2.jpg", testSession2.String()), 100, 100)
		assert.Equal(t, err, nil)
		err = fm.CutFile(testSession3, fmt.Sprintf("temp/%s/testfile3.jpg", testSession3.String()), 10, 10)
		assert.Equal(t, err, fmt.Errorf("error on cut img: %w", imgprocessing.ErrSmallCut))
	})

	t.Run("two archives created", func(t *testing.T) {
		counter = 0
		filepath.WalkDir("temp", walkFunc)
		assert.Equal(t, counter, 5) // 3 + 2 = 5
	})

	var archiveName1, archiveName2 string
	t.Run("getting archive names", func(t *testing.T) {
		t.Run("ok file 1", func(t *testing.T) {
			archiveName1, err = fm.GetArchiveName(testSession1, fmt.Sprintf("temp/%s/testfile1.jpg", testSession1.String()))
			assert.Equal(t, err, nil)
			assert.Equal(t, archiveName1, fmt.Sprintf("temp/%s/testfile1.zip", testSession1.String()))
		})
		t.Run("ok file 2", func(t *testing.T) {
			archiveName2, err = fm.GetArchiveName(testSession2, fmt.Sprintf("temp/%s/testfile2.jpg", testSession2.String()))
			assert.Equal(t, err, nil)
			assert.Equal(t, archiveName2, fmt.Sprintf("temp/%s/testfile2.zip", testSession2.String()))
		})

		t.Run("not found wrong file", func(t *testing.T) {
			archiveNotFound1, err := fm.GetArchiveName(testSession2, fmt.Sprintf("temp/%s/wrongfilename.jpg", testSession2.String()))
			assert.Equal(t, err, ErrFileNotFound)
			assert.Equal(t, archiveNotFound1, "")
		})

		t.Run("not found missing archive", func(t *testing.T) {
			archiveNotFound2, err := fm.GetArchiveName(testSession3, fmt.Sprintf("temp/%s/testfile3.jpg", testSession3.String()))
			assert.Equal(t, err, ErrFileNotFound)
			assert.Equal(t, archiveNotFound2, "")
		})
	})

	t.Run("check cut result", func(t *testing.T) {
		archive1, err := zip.OpenReader(archiveName1)
		assert.Equal(t, err, nil)
		defer archive1.Close()
		assert.Equal(t, len(archive1.File), 110) // 320x339px / 32x32px = (320/32) x (339/32) = 10x11 = 110

		archive2, err := zip.OpenReader(archiveName2)
		assert.Equal(t, err, nil)
		defer archive2.Close()
		assert.Equal(t, len(archive2.File), 16) // 320x339px / 100x100px = (320/100) x (339/100) = 4x4 = 16
	})

	t.Run("delete files", func(t *testing.T) {
		// deleted img + archive
		err := fm.TerminateSession(testSession1)
		assert.Equal(t, err, nil)

		// deleted img + archive
		err = fm.DeleteFile(testSession2, fmt.Sprintf("temp/%s/testfile2.jpg", testSession2.String()))
		assert.Equal(t, err, nil)

		counter = 0
		filepath.WalkDir("temp", walkFunc)
		assert.Equal(t, counter, 1) // 5 -2 -2 = 1
	})
}
