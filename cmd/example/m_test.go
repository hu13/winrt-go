package main

import (
	"fmt"
	"os/user"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/google/uuid"
	"github.com/saltosystems/winrt-go/windows/storage"
	"github.com/saltosystems/winrt-go/windows/storage/provider"
	"github.com/saltosystems/winrt-go/windows/storage/streams"
	"github.com/stretchr/testify/require"
)

func Test_adsf(t *testing.T) {
	ole.RoInitialize(0)

	roots, err := provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	require.NoError(t, err)
	numRoots, err := roots.GetSize()
	require.NoError(t, err)
	fmt.Println("Number of roots:", numRoots)

	// tempBase, err := os.UserCacheDir()
	// require.NoError(t, err)
	// syncRootPath, err := os.MkdirTemp(tempBase, "syncRootPath")
	// require.NoError(t, err)
	syncRootPath := "C:\\Users\\hangk\\AppData\\Local\\syncRootPath1179387943"
	println(syncRootPath)

	writer, err := streams.NewDataWriter()
	require.NoError(t, err)
	syncRootId := []byte("syncRootIdentity")
	err = writer.WriteBytes(uint32(len(syncRootId)), syncRootId)
	require.NoError(t, err)

	bufferContext, err := writer.DetachBuffer()
	require.NoError(t, err)

	syncRootInfo, err := provider.NewStorageProviderSyncRootInfo()
	require.NoError(t, err)

	err = syncRootInfo.SetContext(bufferContext)
	require.NoError(t, err)
	uuid, err := uuid.NewRandom()
	require.NoError(t, err)

	u, err := user.Current()
	require.NoError(t, err)

	userSid, err := syscall.StringToSid(u.Uid)
	require.NoError(t, err)

	sidString, err := userSid.String()
	require.NoError(t, err)

	err = syncRootInfo.SetId(fmt.Sprintf("%s!%s!%v", uuid.String(), sidString, time.Now().Unix()))
	require.NoError(t, err)

	idd, err := syncRootInfo.GetId()
	fmt.Println(">>>>>>> idddd", idd, err)

	res, err := GetFolderFromPath(syncRootPath)
	require.NoError(t, err)

	dir := (*storage.StorageFolder)(res)
	// dpath, err := dir.GetPath()
	// fmt.Println(">>>>>>> dpath", dpath, err)
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
	err = syncRootInfo.SetInSyncPolicy(provider.StorageProviderInSyncPolicyPreserveInsyncForSyncEngine)
	require.NoError(t, err)
	err = syncRootInfo.SetHardlinkPolicy(0)
	require.NoError(t, err)
	err = syncRootInfo.SetVersion("1.0.0.0")
	require.NoError(t, err)

	v, err := syncRootInfo.GetVersion()
	fmt.Println(">>>>>>> version", v, err)
	// syncRootInfo.SetAllowPinning(true)
	// syncRootInfo.SetShowSiblingsAsGroup(false)
	// syncRootInfo.SetProtectionMode(0)
	// syncRootInfo.SetDisplayNameResource("DisplayNameResource-123")
	//PrintAllFields(syncRootInfo)
	fmt.Println(">>>>>>> sync root info", syncRootInfo)

	// needed for newly created sync root path to be there
	time.Sleep(3 * time.Second)

	err = provider.StorageProviderSyncRootManagerRegister(syncRootInfo)
	require.NoError(t, err)

	roots, err = provider.StorageProviderSyncRootManagerGetCurrentSyncRoots()
	require.NoError(t, err)
	println("done")
	numRoots, err = roots.GetSize()
	require.NoError(t, err)
	fmt.Println("Number of roots:", numRoots)

}
