package storage

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/stretchr/testify/require"
)

func init() {
	err := ole.RoInitialize(1)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupCOM() {
	ole.CoUninitialize()
}

// GetFileFromPath retrieves a StorageFile from a given file path using StorageFile.GetFileFromPathAsync api
// https://docs.microsoft.com/en-us/uwp/api/windows.storage.storagefile.getfilefrompathasync
func GetFileFromPath(fp string) (*StorageFile, error) {
	// Create an AsyncOperationCompletedHandler to retrieve the StorageFile
	var storageFile *StorageFile
	var err error
	waitChan := make(chan struct{})
	onCompleteCB := func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		defer close(waitChan)
		if asyncStatus != foundation.AsyncStatusCompleted {
			log.Printf("Async operation did not complete successfully: status %d", asyncStatus)
			err = fmt.Errorf("async operation did not complete successfully: status %d", asyncStatus)
			return
		}

		// Retrieve the StorageFile result from asyncInfo
		var resultPtr unsafe.Pointer
		resultPtr, err = asyncInfo.GetResults()
		if err != nil {
			log.Printf("Failed to get async operation result: %v", err)
			return
		}

		// Cast the result to a StorageFile
		storageFile = (*StorageFile)(resultPtr)
		log.Printf("Retrieved StorageFile: %+v", storageFile)
	}
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, SignatureStorageFile)
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), onCompleteCB)
	defer handler.Release()

	// this is an async operation
	fileAsyncOp, err := StorageFileGetFileFromPathAsync(fp)
	if err != nil {
		return nil, err
	}

	err = fileAsyncOp.SetCompleted(handler)
	if err != nil {
		return nil, err
	}

	// Wait until async operation has stopped, and finish.
	<-waitChan
	return storageFile, err
}

func Test_GetStorageFileFromPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "someDir")
	require.NoError(t, err)
	fpath := filepath.Join(tempDir, "someFile.txt")
	f, err := os.Create(fpath)
	require.NoError(t, err)
	require.NotNil(t, f)
	sfile, err := GetFileFromPath(fpath)
	require.NoError(t, err)
	require.NotNil(t, sfile)
	name, err := sfile.GetName()
	require.NoError(t, err)
	require.Equal(t, filepath.Base(fpath), name)
}
