//go:build windows && amd64

package desktop

import (
	"syscall"
	"unsafe"
)

// Keep the standard Windows caption buttons, resize handles and accessibility,
// while making the native frame match the catalogue's dark cinema palette.
func styleWindow(hwnd uintptr) {
	// A stock brush also keeps the host dark while WebView2 paints its first page.
	brush, _, _ := syscall.NewLazyDLL("gdi32.dll").NewProc("GetStockObject").Call(4) // BLACK_BRUSH
	user32.NewProc("SetClassLongPtrW").Call(hwnd, ^uintptr(9), brush)                // GCLP_HBRBACKGROUND
	setAttribute := syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")
	if setAttribute.Find() != nil {
		return
	}
	set := func(attribute, value uint32) uintptr {
		result, _, _ := setAttribute.Call(hwnd, uintptr(attribute), uintptr(unsafe.Pointer(&value)), unsafe.Sizeof(value))
		return result
	}
	if set(20, 1) != 0 {
		set(19, 1)
	} // immersive dark mode, including older Windows 10
	set(35, 0x00140f0d) // caption: #0d0f14 (COLORREF uses BGR)
	set(36, 0x00e3e6e8) // text: #e8e6e3
	set(34, 0x003a2723) // border: #23273a
	set(33, 2)          // rounded corners where supported (Windows 11)
}

func setApplicationIdentity() {
	setID := syscall.NewLazyDLL("shell32.dll").NewProc("SetCurrentProcessExplicitAppUserModelID")
	if setID.Find() != nil {
		return
	}
	appID, _ := syscall.UTF16PtrFromString("Bist0uille.AbsoluteCinema")
	setID.Call(uintptr(unsafe.Pointer(appID)))
}
