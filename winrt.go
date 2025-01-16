package winrt

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.IDataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter !*
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Streams.DataWriter -method-filter WriteBytes -method-filter DetachBuffer -method-filter DataWriter -method-filter Close -method-filter !*

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageFolder -method-filter !CreateFileQueryOverloadDefault -method-filter !CreateFileQuery -method-filter !CreateFileQueryWithOptions -method-filter !CreateFolderQueryOverloadDefault -method-filter !CreateFolderQuery -method-filter !CreateFolderQueryWithOptions -method-filter !CreateItemQuery -method-filter !CreateItemQueryWithOptions -method-filter !GetFilesAsync -method-filter !GetFilesAsyncOverloadDefaultStartAndCount -method-filter !AreQueryOptionsSupported -method-filter !IsCommonFolderQuerySupported -method-filter !IsCommonFileQuerySupported -method-filter !TryGetChangeTracker -method-filter !GetFoldersAsync -method-filter !GetFoldersAsyncOverloadDefaultStartAndCount -method-filter !GetIndexedStateAsync -method-filter !GetItemsAsync -method-filter !GetFolderFromPathForUserAsync

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageFile -method-filter !OpenSequentialReadAsync -method-filter !OpenReadAsync -method-filter !GetScaledImageAsThumbnailAsyncOverloadDefaultSizeDefaultOptions -method-filter !GetScaledImageAsThumbnailAsyncOverloadDefaultOptions -method-filter !GetScaledImageAsThumbnailAsync -method-filter !GetParentAsync -method-filter !CreateStreamedFileAsync -method-filter !GetFileFromPathForUserAsync -method-filter !IsEqual -method-filter !OpenWithOptionsAsync -method-filter !OpenTransactedWriteWithOptionsAsync -method-filter !ReplaceWithStreamedFileAsync -method-filter !CreateStreamedFileFromUriAsync -method-filter !ReplaceWithStreamedFileFromUriAsync

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.HResult

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.ThumbnailOptions
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.ThumbnailMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileProperties.StorageItemContentProperties -method-filter !RetrievePropertiesAsync -method-filter !SavePropertiesAsync -method-filter !SavePropertiesAsyncOverloadDefault

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageProvider
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFilePropertiesWithAvailability
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemPropertiesWithProvider
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderItemProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderItemProperty
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderItemPropertyDefinition

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Uri
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IStringable
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.WwwFormUrlDecoder
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IIterable`1
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.Collections.IIterator`1

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.IStorageProviderStatusUISource
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderStatusUI
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderSyncRootInfo
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderQuotaUI
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.IStorageProviderUICommand
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderUICommandState
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderMoreInfoUI
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderHydrationPolicy
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderHydrationPolicyModifier
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderPopulationPolicy
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderInSyncPolicy
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderProtectionMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderHardlinkPolicy
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFolder
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFolder2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.CreationCollisionOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.Provider.StorageProviderSyncRootManager

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageFile
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItem
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItem2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.NameCollisionOption
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncAction
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.AsyncActionCompletedHandler
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageDeleteOption

//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileAttributes
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.FileAccessMode
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.StorageItemTypes
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemProperties
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Storage.IStorageItemProperties2
//go:generate go run github.com/saltosystems/winrt-go/cmd/winrt-go-gen -debug -class Windows.Foundation.IAsyncInfo
