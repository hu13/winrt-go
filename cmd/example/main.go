package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/storage"
	"github.com/saltosystems/winrt-go/windows/storage/provider"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
	"golang.org/x/sys/windows"
)

func main() {
	ole.RoInitialize(0)
	if err := run2(); err != nil {
		panic(err)
	}
}

func GetFolderFromPath(fp string) (*storage.StorageFolder, error) {
	var folder *storage.StorageFolder
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
		folder = (*storage.StorageFolder)(resultPtr)
		folderPath, _ := folder.GetPath()
		createdDate, _ := folder.GetDateCreated()
		log.Printf("Retrieved StorageFolder: %v, path: %v, createdDate: %v", folder, folderPath, createdDate)
	}
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, storage.SignatureStorageFolder)
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), onCompleteCB)
	defer handler.Release()

	// this is an async operation
	fileAsyncOp, err := storage.StorageFolderGetFolderFromPathAsync(fp)
	if err != nil {
		return nil, err
	}

	err = fileAsyncOp.SetCompleted(handler)
	if err != nil {
		return nil, err
	}

	// Wait until async operation has stopped, and finish.
	<-waitChan
	return folder, err
}

// GetFileFromPath retrieves a StorageFile from a given file path using StorageFile.GetFileFromPathAsync api
// https://docs.microsoft.com/en-us/uwp/api/windows.storage.storagefile.getfilefrompathasync
func GetFileFromPath(fp string) (*storage.StorageFile, error) {
	// Create an AsyncOperationCompletedHandler to retrieve the StorageFile
	var storageFile *storage.StorageFile
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
		storageFile = (*storage.StorageFile)(resultPtr)
		log.Printf("Retrieved StorageFile: %+v", storageFile)
	}
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, storage.SignatureStorageFile)
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), onCompleteCB)
	defer handler.Release()

	// this is an async operation
	fileAsyncOp, err := storage.StorageFileGetFileFromPathAsync(fp)
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

func run() error {
	// create info
	infoID := "infoID"

	info, err := provider.NewStorageProviderSyncRootInfo()
	if err != nil {
		return err
	}

	err = info.SetId(infoID)
	if err != nil {
		return err
	}

	storageFolderAsync, err := storage.StorageFolderGetFolderFromPathAsync(`C:\Users\hangk\work\winrt-go`)
	if err != nil {
		return err
	}

	if err := awaitAsyncOperation(storageFolderAsync, storage.SignatureStorageFolder); err != nil {
		return err
	}

	res, err := storageFolderAsync.GetResults()
	if err != nil {
		return err
	}

	// res, err := GetFolderFromPath(`C:\Users\hangk\work\windows\pv_cloud_drive_root`)
	// if err != nil {
	// 	return err
	// }

	folder := (*storage.StorageFolder)(res)
	itf := folder.MustQueryInterface(ole.NewGUID(storage.GUIDIStorageFolder))
	defer itf.Release()
	f := (*storage.IStorageFolder)(unsafe.Pointer(itf))
	if err := info.SetPath(f); err != nil {
		return err
	}

	// register info
	err = provider.StorageProviderSyncRootManagerRegister(info)
	if err != nil {
		return err
	}

	// unregister info
	err = provider.StorageProviderSyncRootManagerUnregister(infoID)
	if err != nil {
		return err
	}

	// release info
	info.Release()

	return nil
}

func run2() error {
	roots, err := provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	if err != nil {
		return err
	}
	numRoots, err := roots.GetSize()
	if err != nil {
		return err
	}
	fmt.Println("Number of roots:", numRoots)

	tempBase, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	syncRootPath, err := os.MkdirTemp(tempBase, "syncRootPath")
	if err != nil {
		return err
	}

	writer, err := streams.NewDataWriter()
	if err != nil {
		return err
	}
	syncRootId := []byte("syncRootIdentity")
	err = writer.WriteBytes(uint32(len(syncRootId)), syncRootId)
	if err != nil {
		return err
	}

	bufferContext, err := writer.DetachBuffer()
	if err != nil {
		return err
	}

	reader, err := streams.DataReaderFromBuffer(bufferContext)
	if err != nil {
		return err
	}
	bufferContent, err := reader.ReadBytes(uint32(len(syncRootId)))
	if err != nil {
		return err
	}
	fmt.Println(">>>>>>> buffer content", bufferContent, string(bufferContent))

	syncRootInfo, err := provider.NewStorageProviderSyncRootInfo()
	if err != nil {
		return err
	}

	err = syncRootInfo.SetContext(bufferContext)
	if err != nil {
		return err
	}

	providerGUID := uuid.New().String()
	fmt.Println(">>>>>>> set provider guid", providerGUID)
	// Open the current process token
	var token windows.Token
	err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return err
	}
	defer token.Close() // Ensure the token handle is closed
	user, err := token.GetTokenUser()
	if err != nil {
		return err
	}
	sid := user.User.Sid.String()
	syncRootId2 := fmt.Sprintf("%s!%s!%d", providerGUID, sid, -1438710713)
	fmt.Println(">>>>>>> set sync root id", syncRootId2)
	// required
	err = syncRootInfo.SetId(syncRootId2)
	if err != nil {
		return err
	}
	idd, err := syncRootInfo.GetId()
	fmt.Println(">>>>>>> get sync root id", idd, err)

	parsedGUID := ole.NewGUID(providerGUID)
	sysGUID := syscall.GUID{
		Data1: uint32(parsedGUID.Data1),
		Data2: parsedGUID.Data2,
		Data3: parsedGUID.Data3,
		Data4: [8]byte{parsedGUID.Data4[0], parsedGUID.Data4[1], parsedGUID.Data4[2], parsedGUID.Data4[3], parsedGUID.Data4[4], parsedGUID.Data4[5], parsedGUID.Data4[6], parsedGUID.Data4[7]},
	}
	err = syncRootInfo.SetProviderId(sysGUID)
	if err != nil {
		return err
	}

	getid, err := syncRootInfo.GetProviderId()
	if err != nil {
		return err
	}
	getidStr := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", getid.Data1, getid.Data2, getid.Data3, getid.Data4[0:2], getid.Data4[2:])
	fmt.Println(">>>>>>> get providerID:", getidStr)

	// this is not causing the crash
	res, err := GetFolderFromPath(syncRootPath)
	if err != nil {
		return err
	}
	dir := (*storage.StorageFolder)(res)
	itf3 := dir.MustQueryInterface(ole.NewGUID(storage.GUIDIStorageFolder))
	defer itf3.Release()
	iStorageDir := (*storage.IStorageFolder)(unsafe.Pointer(itf3))
	err = syncRootInfo.SetPath(iStorageDir)
	if err != nil {
		return err
	}

	// not required coz still crashes without them
	err = syncRootInfo.SetHydrationPolicy(2)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetHydrationPolicyModifier(0)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetPopulationPolicy(1)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetInSyncPolicy(provider.StorageProviderInSyncPolicyPreserveInsyncForSyncEngine)
	if err != nil {
		return err
	}
	err = syncRootInfo.SetHardlinkPolicy(0)
	if err != nil {
		return err
	}

	// required
	err = syncRootInfo.SetVersion("1.0")
	if err != nil {
		return err
	}
	syncRootInfo.SetAllowPinning(true)
	syncRootInfo.SetShowSiblingsAsGroup(false)
	syncRootInfo.SetProtectionMode(1)
	syncRootInfo.SetDisplayNameResource(filepath.Base(syncRootPath))
	fmt.Printf(">>>>>>>>>> syncRootInfo: %+v\n", syncRootInfo)

	err = provider.StorageProviderSyncRootManagerRegister(syncRootInfo)
	if err != nil {
		return err
	}
	runtime.KeepAlive(syncRootInfo)

	fmt.Println(">>>>>>> registered, err", err)

	roots, err = provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	if err != nil {
		return err
	}
	fmt.Println(">>>>>>> got current sync roots", roots)
	fmt.Println(">>>>>>> err", err)
	fmt.Println("done")
	numRoots, err = roots.GetSize()
	if err != nil {
		return err
	}
	fmt.Println("Number of roots:", numRoots)

	return nil
}

func awaitAsyncOperation(asyncOperation *foundation.IAsyncOperation, genericParamSignature string) error {
	var status foundation.AsyncStatus

	// We need to obtain the GUID of the AsyncOperationCompletedHandler, but its a generic delegate
	// so we also need the generic parameter type's signature:
	// AsyncOperationCompletedHandler<genericParamSignature>
	iid := winrt.ParameterizedInstanceGUID(foundation.GUIDAsyncOperationCompletedHandler, genericParamSignature)

	// Wait until the async operation completes.
	waitChan := make(chan struct{})
	handler := foundation.NewAsyncOperationCompletedHandler(ole.NewGUID(iid), func(instance *foundation.AsyncOperationCompletedHandler, asyncInfo *foundation.IAsyncOperation, asyncStatus foundation.AsyncStatus) {
		status = asyncStatus
		close(waitChan)
	})
	defer handler.Release()

	asyncOperation.SetCompleted(handler)

	// Wait until async operation has stopped, and finish.
	asyncWait := true
	for asyncWait {
		select {
		case <-time.After(30 * time.Second):
			itf, err := asyncOperation.QueryInterface(ole.NewGUID(foundation.GUIDIAsyncInfo))
			if err != nil {
				return err
			}
			defer itf.Release()
			v := (*foundation.IAsyncInfo)(unsafe.Pointer(itf))
			if err := v.Cancel(); err != nil {
				return err
			}
			println("Waiting for operation cancel")
		case <-waitChan:
			asyncWait = false
		}
	}

	if status != foundation.AsyncStatusCompleted {
		return fmt.Errorf("async operation failed with status %d", status)
	}
	return nil
}
