package collections

import (
	"log"
	"sync"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/saltosystems/winrt-go"

	"github.com/saltosystems/winrt-go/internal/delegate"
	"github.com/saltosystems/winrt-go/internal/kernel32"
)

var (
	ole32                     = syscall.NewLazyDLL("ole32.dll")
	procCoRegisterClassObject = ole32.NewProc("CoRegisterClassObject")
)

func RegisterInstance(clsid *ole.GUID, obj unsafe.Pointer) (*ole.IUnknown, error) {
	var cookie uintptr
	ret, _, err := procCoRegisterClassObject.Call(
		uintptr(unsafe.Pointer(clsid)),   // CLSID
		uintptr(obj),                     // Object implementing the interface
		uintptr(ole.CLSCTX_LOCAL_SERVER), // Context (adjust as needed)
		uintptr(1),                       // Flags
		uintptr(unsafe.Pointer(&cookie)), // Out cookie
	)
	log.Println(">>>>>", ret, err)
	if ret != 0 {
		return nil, ole.NewError(ret)
	}
	return nil, nil

}

var (
	firstCallback = syscall.NewCallback(first)
)

// We cannot pass a pointer that includes Go pointers to WinRT
// so we either copy the arrays manually into the Heap
// or store them somewhere in Go (which is simpler).
//
// This struct stores each array in a map that uses the Pointer passed
// to WinRT as a key. This pointer is Heap allocated and will never change.
// Once the pointer is Released, the array is also removed.
type syncArrayIterables struct {
	mu        *sync.Mutex
	iterables map[unsafe.Pointer][]any
}

var arrayItems = &syncArrayIterables{
	mu:        &sync.Mutex{},
	iterables: make(map[unsafe.Pointer][]any),
}

func (i *syncArrayIterables) add(instPtr unsafe.Pointer, items []any) {
	arrayItems.mu.Lock()
	defer arrayItems.mu.Unlock()

	i.iterables[instPtr] = items
}

func (i *syncArrayIterables) get(instPtr unsafe.Pointer) ([]any, bool) {
	arrayItems.mu.Lock()
	defer arrayItems.mu.Unlock()

	it, ok := arrayItems.iterables[instPtr]
	return it, ok
}

func (i *syncArrayIterables) remove(instPtr unsafe.Pointer) {
	arrayItems.mu.Lock()
	defer arrayItems.mu.Unlock()

	delete(i.iterables, instPtr)
}

type arrayIterable struct {
	IIterable
	sync.Mutex
	refs          uintptr
	IID           ole.GUID
	itemSignature string
}

func init() {
	ole.CoInitialize(0)
}

func NewArrayIterable(items []any, itemSignature string) *IIterable {
	iid := ole.NewGUID(winrt.ParameterizedInstanceGUID(GUIDIIterable, itemSignature))

	// create type instance
	size := unsafe.Sizeof(*(*arrayIterable)(nil))
	instPtr := kernel32.Malloc(size)
	inst := (*arrayIterable)(instPtr)

	// get the callbacks for the VTable
	callbacks := delegate.RegisterCallbacks(instPtr, inst)

	log.Printf("Registering instance with CLSID: %s", iid.String())

	// the VTable should also be allocated in the heap
	sizeVTable := unsafe.Sizeof(*(*IIterableVtbl)(nil))
	vTablePtr := kernel32.Malloc(sizeVTable)

	inst.RawVTable = (*interface{})(vTablePtr)

	vTable := (*IIterableVtbl)(vTablePtr)
	vTable.IUnknownVtbl = ole.IUnknownVtbl{
		QueryInterface: callbacks.QueryInterface,
		AddRef:         callbacks.AddRef,
		Release:        callbacks.Release,
	}
	vTable.First = firstCallback

	arrayItems.add(instPtr, items)
	// Initialize all properties: the malloc may contain garbage
	inst.IID = *iid // copy contents
	inst.Mutex = sync.Mutex{}
	inst.refs = 0
	inst.itemSignature = itemSignature

	inst.AddRef()
	return &inst.IIterable // ugly but works
}

func (r *arrayIterable) GetIID() *ole.GUID {
	return &r.IID
}

// not sure
func (r *arrayIterable) Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr {
	_, ok := arrayItems.get(instancePtr)
	if !ok {
		// instance not found
		return ole.E_FAIL
	}
	return ole.S_OK
}

// addRef increments the reference counter by one
func (r *arrayIterable) AddRef() uintptr {
	r.Lock()
	defer r.Unlock()
	r.refs++
	return r.refs
}

func (r *arrayIterable) removeRef() uintptr {
	r.Lock()
	defer r.Unlock()

	if r.refs > 0 {
		r.refs--
	}

	return r.refs
}

// removeRef decrements the reference counter by one. If it was already zero, it will just return zero.
func (r *arrayIterable) Release() uintptr {
	rem := r.removeRef()
	if rem == 0 {
		// We're done.
		instancePtr := unsafe.Pointer(r)
		arrayItems.remove(instancePtr)
		kernel32.Free(unsafe.Pointer(r.RawVTable))
		kernel32.Free(instancePtr)
	}
	return rem
}

func first(inst unsafe.Pointer, out **IIterator) uintptr {
	offset := unsafe.Offsetof(arrayIterable{}.IIterable)
	i := (*arrayIterable)(unsafe.Pointer(uintptr(inst) - offset))
	arrIt, ok := arrayItems.get(inst)
	if !ok {
		return ole.E_FAIL
	}

	it := NewArrayIterator(arrIt, i.itemSignature)

	m := (**IIterator)(out)
	*m = it
	return ole.S_OK
}

var (
	getCurrentCallback    = syscall.NewCallback(getCurrent)
	getHasCurrentCallback = syscall.NewCallback(getHasCurrent)
	moveNextCallback      = syscall.NewCallback(moveNext)
	getManyCallback       = syscall.NewCallback(getMany)
)

type collectionsIterator struct {
	IIterator
	sync.Mutex
	refs          uintptr
	IID           ole.GUID
	index         int
	itemSignature string
}

func NewArrayIterator(items []any, itemSignature string) *IIterator {
	iid := ole.NewGUID(winrt.ParameterizedInstanceGUID(GUIDIIterator, itemSignature))

	// create type instance
	size := unsafe.Sizeof(*(*collectionsIterator)(nil))
	instPtr := kernel32.Malloc(size)
	inst := (*collectionsIterator)(instPtr)

	// get the callbacks for the VTable
	callbacks := delegate.RegisterCallbacks(instPtr, inst)

	// the VTable should also be allocated in the heap
	sizeVTable := unsafe.Sizeof(*(*IIterableVtbl)(nil))
	vTablePtr := kernel32.Malloc(sizeVTable)

	inst.RawVTable = (*interface{})(vTablePtr)

	vTable := (*IIteratorVtbl)(vTablePtr)
	vTable.IUnknownVtbl = ole.IUnknownVtbl{
		QueryInterface: callbacks.QueryInterface,
		AddRef:         callbacks.AddRef,
		Release:        callbacks.Release,
	}

	vTable.GetCurrent = getCurrentCallback
	vTable.GetHasCurrent = getHasCurrentCallback
	vTable.GetMany = getManyCallback
	vTable.MoveNext = moveNextCallback

	arrayItems.add(instPtr, items)

	// Initialize all properties: the malloc may contain garbage
	inst.IID = *iid // copy contents
	inst.Mutex = sync.Mutex{}
	inst.refs = 0
	inst.itemSignature = itemSignature
	inst.index = -1 // not initialized

	inst.AddRef()
	return &inst.IIterator // ugly but works
}

func (r *collectionsIterator) GetIID() *ole.GUID {
	return &r.IID
}

// addRef increments the reference counter by one
func (r *collectionsIterator) AddRef() uintptr {
	r.Lock()
	defer r.Unlock()
	r.refs++
	return r.refs
}

func (r *collectionsIterator) removeRef() uintptr {
	r.Lock()
	defer r.Unlock()

	if r.refs > 0 {
		r.refs--
	}

	return r.refs
}

func (instance *collectionsIterator) Invoke(instancePtr, rawArgs0, rawArgs1, rawArgs2, rawArgs3, rawArgs4, rawArgs5, rawArgs6, rawArgs7, rawArgs8 unsafe.Pointer) uintptr {
	log.Println("Invoke not implemented")
	return ole.S_OK
}

// removeRef decrements the reference counter by one. If it was already zero, it will just return zero.
func (r *collectionsIterator) Release() uintptr {
	rem := r.removeRef()
	if rem == 0 {
		// We're done.
		instancePtr := unsafe.Pointer(r)
		arrayItems.remove(instancePtr)
		kernel32.Free(unsafe.Pointer(r.RawVTable))
		kernel32.Free(instancePtr)
	}
	return rem
}

func getCurrent(inst, out unsafe.Pointer) uintptr {
	// good old C pointer magic tricks
	offset := unsafe.Offsetof(collectionsIterator{}.IIterator)
	it := (*collectionsIterator)(unsafe.Pointer(uintptr(inst) - offset))

	items, ok := arrayItems.get(unsafe.Pointer(it))
	if !ok {
		return ole.E_FAIL
	}

	current := items[it.index]

	copyItemToPointer(current, out)

	return ole.S_OK
}

func getHasCurrent(inst, out unsafe.Pointer) uintptr {
	// good old C pointer magic tricks
	offset := unsafe.Offsetof(collectionsIterator{}.IIterator)
	it := (*collectionsIterator)(unsafe.Pointer(uintptr(inst) - offset))

	items, ok := arrayItems.get(unsafe.Pointer(it))
	if !ok {
		return ole.E_FAIL
	}

	if it.index > -1 && it.index < len(items) {
		// index is within items length
		*(*uintptr)(out) = uintptr(1)
	} else {
		*(*uintptr)(out) = uintptr(0)
	}
	return ole.S_OK
}

func moveNext(inst, out unsafe.Pointer) uintptr {
	// good old C pointer magic tricks
	offset := unsafe.Offsetof(collectionsIterator{}.IIterator)
	it := (*collectionsIterator)(unsafe.Pointer(uintptr(inst) - offset))

	it.index++
	return getHasCurrent(inst, out)
}

func getMany(inst, itemsAmount, outItems, outItemsSize unsafe.Pointer) uintptr {
	// good old C pointer magic tricks
	offset := unsafe.Offsetof(collectionsIterator{}.IIterator)
	it := (*collectionsIterator)(unsafe.Pointer(uintptr(inst) - offset))

	items, ok := arrayItems.get(unsafe.Pointer(it))
	if !ok {
		return ole.E_FAIL
	}

	// requested itemsAmount
	requestedItems := int(uintptr(itemsAmount))
	availableItems := len(items) - it.index - 1
	if availableItems < requestedItems {
		// not enough items available
		requestedItems = availableItems
	}

	// write items size
	value := uint32(requestedItems)
	log.Printf("Writing %d to outItemsSize at %p\n", value, outItemsSize)
	*(*uint32)(outItemsSize) = value

	// copy items
	n := uintptr(0)
	hasCurrent := false // unused
	for i := 0; i < requestedItems; i++ {
		moveNext(inst, unsafe.Pointer(&hasCurrent))
		n += copyItemToPointer(items[it.index], unsafe.Pointer(uintptr(outItems)+n))
	}

	return ole.S_OK
}

func copyItemToPointer(item any, out unsafe.Pointer) uintptr {
	var size uintptr = 0
	switch t := item.(type) {
	case uint:
		*(*uint)(out) = t
		size = unsafe.Sizeof(uint(0))
	case uint8:
		*(*uint8)(out) = t
		size = unsafe.Sizeof(uint8(0))
	case uint16:
		*(*uint16)(out) = t
		size = unsafe.Sizeof(uint16(0))
	case uint32:
		*(*uint32)(out) = t
		size = unsafe.Sizeof(uint32(0))
	case uint64:
		*(*uint64)(out) = t
		size = unsafe.Sizeof(uint64(0))
	case int:
		*(*int)(out) = t
		size = unsafe.Sizeof(int(0))
	case int8:
		*(*int8)(out) = t
		size = unsafe.Sizeof(int8(0))
	case int16:
		*(*int16)(out) = t
		size = unsafe.Sizeof(int16(0))
	case int32:
		*(*int32)(out) = t
		size = unsafe.Sizeof(int32(0))
	case int64:
		*(*int64)(out) = t
		size = unsafe.Sizeof(int64(0))
	case *any:
		*(*any)(out) = t
		size = unsafe.Sizeof(uintptr(0))
	}

	return size
}
