package provider

import (
	"fmt"
	"log"
	"os"
	"testing"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"
	"github.com/saltosystems/winrt-go/windows/foundation"
	"github.com/saltosystems/winrt-go/windows/storage"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
	"github.com/stretchr/testify/require"
)

func init() {
	ole.RoInitialize(1)
}

func PrintAllFields(impl *StorageProviderSyncRootInfo) {
	fields := []struct {
		Name  string
		Value func() (interface{}, error)
	}{
		{Name: "ID", Value: func() (interface{}, error) { return impl.GetId() }},
		{"Path", func() (interface{}, error) { return impl.GetPath() }},
		{"DisplayNameResource", func() (interface{}, error) { return impl.GetDisplayNameResource() }},
		{"IconResource", func() (interface{}, error) { return impl.GetIconResource() }},
		{"HydrationPolicy", func() (interface{}, error) { return impl.GetHydrationPolicy() }},
		{"HydrationPolicyModifier", func() (interface{}, error) { return impl.GetHydrationPolicyModifier() }},
		{"PopulationPolicy", func() (interface{}, error) { return impl.GetPopulationPolicy() }},
		{"InSyncPolicy", func() (interface{}, error) { return impl.GetInSyncPolicy() }},
		{"HardlinkPolicy", func() (interface{}, error) { return impl.GetHardlinkPolicy() }},
		{"ShowSiblingsAsGroup", func() (interface{}, error) { return impl.GetShowSiblingsAsGroup() }},
		{"Version", func() (interface{}, error) { return impl.GetVersion() }},
		{"ProtectionMode", func() (interface{}, error) { return impl.GetProtectionMode() }},
		{"AllowPinning", func() (interface{}, error) { return impl.GetAllowPinning() }},
		{"StorageProviderItemPropertyDefinitions", func() (interface{}, error) { return impl.GetStorageProviderItemPropertyDefinitions() }},
		{"RecycleBinUri", func() (interface{}, error) { return impl.GetRecycleBinUri() }},
		{"Context", func() (interface{}, error) { return impl.GetContext() }},
	}

	fmt.Println("StorageProviderSyncRootInfo Fields:")
	for _, field := range fields {
		value, err := field.Value()
		if err != nil {
			fmt.Printf("  %s: Error (%v)\n", field.Name, err)
		} else {
			fmt.Printf("  %s: %v\n", field.Name, value)
		}
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
		log.Printf("Retrieved StorageFile: %+v", folder)
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

func Test_GetCurrentSyncRoots(t *testing.T) {
	// tr := initTestResource(t, withTestBrowseDirFn(defaultBrowseDirTestFunc), withConnectSyncRoot())
	// defer tr.cleanUp()

	ok, err := StorageProviderSyncRootManagerIsSupported()
	require.NoError(t, err)
	require.True(t, ok)

	roots, err := StorageProviderSyncRootManagerGetCurrentSyncRoots()
	require.NoError(t, err)
	numRoots, err := roots.GetSize()
	require.NoError(t, err)
	fmt.Println("Number of roots:", numRoots)
	require.True(t, numRoots == 0)

	tempBase, err := os.UserCacheDir()
	require.NoError(t, err)
	syncRootPath, err := os.MkdirTemp(tempBase, "syncRootPath")
	require.NoError(t, err)

	writer, err := streams.NewDataWriter()
	require.NoError(t, err)
	syncRootId := []byte("syncRootIdentity")
	err = writer.WriteBytes(uint32(len(syncRootId)), syncRootId)
	require.NoError(t, err)

	bufferContext, err := writer.DetachBuffer()
	require.NoError(t, err)

	syncRootInfo, err := NewStorageProviderSyncRootInfo()
	require.NoError(t, err)

	err = syncRootInfo.SetContext(bufferContext)
	require.NoError(t, err)

	syncRootInfo.SetId("{00000000-1234-0000-0000-000000000001}")
	require.NoError(t, err)
	dir, err := GetFolderFromPath(syncRootPath)
	require.NoError(t, err)
	itf3 := dir.MustQueryInterface(ole.NewGUID(storage.GUIDIStorageFolder))
	defer itf3.Release()
	iStorageDir := (*storage.IStorageFolder)(unsafe.Pointer(itf3))
	err = syncRootInfo.SetPath(iStorageDir)
	require.NoError(t, err)
	err = syncRootInfo.SetHydrationPolicy(2)
	require.NoError(t, err)
	err = syncRootInfo.SetHydrationPolicyModifier(0)
	require.NoError(t, err)
	err = syncRootInfo.SetPopulationPolicy(2)
	require.NoError(t, err)
	err = syncRootInfo.SetInSyncPolicy(StorageProviderInSyncPolicyPreserveInsyncForSyncEngine)
	require.NoError(t, err)
	err = syncRootInfo.SetHardlinkPolicy(0)
	require.NoError(t, err)
	err = syncRootInfo.SetVersion("1.0")
	require.NoError(t, err)
	syncRootInfo.SetAllowPinning(true)
	syncRootInfo.SetShowSiblingsAsGroup(false)
	syncRootInfo.SetProtectionMode(0)
	syncRootInfo.SetDisplayNameResource("DisplayNameResource")
	PrintAllFields(syncRootInfo)
	fmt.Println(">>>>>>> sync root info", syncRootInfo)

	err = StorageProviderSyncRootManagerRegister(syncRootInfo)
	require.NoError(t, err)

	roots, err = StorageProviderSyncRootManagerGetCurrentSyncRoots()
	require.NoError(t, err)
	numRoots, err = roots.GetSize()
	require.NoError(t, err)
	fmt.Println("Number of roots:", numRoots)
	require.True(t, numRoots == 1)
}

func xTest_GetManyStorageProperyItem(t *testing.T) {

	prop1, err := NewStorageProviderItemProperty()
	require.NoError(t, err)
	prop1.SetId(1)
	prop1.SetValue("Value1")
	prop1.SetIconResource("shell32.dll,-44")

	prop2, err := NewStorageProviderItemProperty()
	require.NoError(t, err)
	prop2.SetId(2)
	prop2.SetValue("Value2")
	prop2.SetIconResource("shell32.dll,-44")

	log.Println("prop1", prop1)
	log.Println("prop2", prop2)

	a := winrt.NewArrayIterable([]any{prop1, prop2}, SignatureStorageProviderItemProperty)

	it, err := a.First()
	require.NoError(t, err)
	resp, n, err := it.GetMany(3) // only 2 is available
	require.NoError(t, err)
	require.Equal(t, uint32(2), n)

	println("RESP", n, resp, len(resp))

	// Extract and print the StorageProviderItemProperty objects
	for i := uint32(0); i < n; i++ {
		itemPtr := unsafe.Pointer(resp[i]) // Convert uintptr to pointer

		item := (*StorageProviderItemProperty)(itemPtr) // Cast pointer to StorageProviderItemProperty

		// Access properties of the StorageProviderItemProperty
		id, _ := item.GetId()
		value, _ := item.GetValue()
		iconResource, _ := item.GetIconResource()

		fmt.Printf("Item %d: Id=%d, Value=%s, IconResource=%s\n", i+1, id, value, iconResource)
	}
}
