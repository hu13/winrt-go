package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/storage"
	"github.com/saltosystems/winrt-go/windows/storage/provider"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
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

/*
Sync Root Information:
Id: 3ea0d29c-377c-47e6-9df5-d24832f63ded!S-1-5-21-219854-3445463206-450736542-1003!-1438710713
AllowPinning: True
DisplayNameResource: CfapiSync
HardlinkPolicy: None
HydrationPolicy: Partial
HydrationPolicyModifier: StreamingAllowed, AutoDehydrationAllowed
InSyncPolicy: FileLastWriteTime
Path: Windows.Storage.StorageFolder
PopulationPolicy: Full
ProtectionMode: Unknown
ProviderId: 3ea0d29c-377c-47e6-9df5-d24832f63ded
Version: 1.0.0.0
IconResource: C:\WINDOWS\system32\imageres.dll,-1043
ShowSiblingsAsGroup: False
RecycleBinUri:
Context: System.__ComObject
*/

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

	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil
	}

	u, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user:", err)
		return err
	}

	userSid, err := syscall.StringToSid(u.Uid)
	if err != nil {
		return err
	}

	sidString, err := userSid.String()
	if err != nil {
		return nil
	}

	err = syncRootInfo.SetId(fmt.Sprintf("%s!%s!-1438710713", uuid.String(), sidString))
	if err != nil {
		fmt.Println("Error setting ID:", err)
		return err
	}

	err = syncRootInfo.SetIconResource("C:\\WINDOWS\\system32\\imageres.dll,-1043")
	if err != nil {
		fmt.Println("Error setting Icon resources:", err)
		return err
	}

	// this is not causing the crash
	tempBase, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	syncRootPath, err := os.MkdirTemp(tempBase, "syncRootPath")
	if err != nil {
		return err
	}

	// syncRootPath := "C:\\Users\\hangk\\AppData\\Local\\syncRootPath1179387943"
	// println(syncRootPath)
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
	err = syncRootInfo.SetVersion("1.0.0.0")
	if err != nil {
		return err
	}

	v, err := syncRootInfo.GetVersion()
	fmt.Println(">>>>>>> version", v, err)
	syncRootInfo.SetAllowPinning(true)
	syncRootInfo.SetShowSiblingsAsGroup(false)
	syncRootInfo.SetProtectionMode(1)
	syncRootInfo.SetDisplayNameResource(filepath.Base(syncRootPath))
	// PrintAllFields(syncRootInfo)
	fmt.Println(">>>>>>> sync root info", syncRootInfo)

	err = provider.StorageProviderSyncRootManagerRegister(syncRootInfo)
	if err != nil {
		return err
	}

	roots, err = provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	if err != nil {
		return err
	}
	println("done")
	numRoots, err = roots.GetSize()
	if err != nil {
		return err
	}
	fmt.Println("Number of roots:", numRoots)

	return nil
}
