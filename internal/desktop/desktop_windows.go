//go:build windows && amd64

package desktop

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"absolute_cinema/internal/config"
	"absolute_cinema/internal/player"
	"absolute_cinema/internal/server"
	webview "github.com/jchv/go-webview2"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var createWindow = user32.NewProc("CreateWindowExW")
var setWindowPos = user32.NewProc("SetWindowPos")
var showWindow = user32.NewProc("ShowWindow")
var getWindowLong = user32.NewProc("GetWindowLongPtrW")
var setWindowLong = user32.NewProc("SetWindowLongPtrW")
var getWindowRect = user32.NewProc("GetWindowRect")
var getMonitor = user32.NewProc("MonitorFromWindow")
var getMonitorInfo = user32.NewProc("GetMonitorInfoW")
var getClientRect = user32.NewProc("GetClientRect")

type rectangle struct{ Left, Top, Right, Bottom int32 }
type point struct{ X, Y int32 }
type windowPlacement struct {
	Length, Flags, ShowCommand uint32
	MinPosition, MaxPosition   point
	NormalPosition             rectangle
}

func Run(baseURL string, paths config.Paths, srv *server.Server) (bool, error) {
	directory := player.LibraryDirectory(paths.MediaRoot)
	if directory == "" {
		return false, fmt.Errorf("VLC requis pour la lecture native")
	}
	if _, err := os.Stat(filepath.Join(directory, "libvlc.dll")); err != nil {
		return false, err
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	setApplicationIdentity()
	cache, err := os.UserCacheDir()
	if err != nil {
		cache = os.TempDir()
	}
	w := webview.NewWithOptions(webview.WebViewOptions{
		DataPath: filepath.Join(cache, "AbsoluteCinema", "WebView2"), AutoFocus: true,
		WindowOptions: webview.WindowOptions{Title: "Absolute Cinema", Width: 1280, Height: 820, Center: true, IconId: 1},
	})
	if w == nil {
		return false, fmt.Errorf("WebView2 indisponible")
	}
	defer w.Destroy()
	w.SetSize(800, 560, webview.HintMin)
	hwnd := uintptr(w.Window())
	styleWindow(hwnd)
	class, _ := syscall.UTF16PtrFromString("STATIC")
	// WS_CHILD | WS_CLIPSIBLINGS | WS_CLIPCHILDREN. The native video is a
	// sibling of WebView2; only its rectangle covers the HTML placeholder.
	// SS_BLACKRECT keeps loading/letterbox areas black even without a decoder.
	surface, _, createErr := createWindow.Call(0, uintptr(unsafe.Pointer(class)), 0, 0x46000004, 0, 0, 1, 1, hwnd, 0, 0, 0)
	if surface == 0 {
		return false, fmt.Errorf("surface vidéo : %w", createErr)
	}
	// USB plugin discovery can take seconds while the drive is busy. Never
	// perform it on the window thread: the catalogue must keep responding.
	var engine *vlcEngine
	var engineError error
	engineLoaded := make(chan struct{})
	var shuttingDown atomic.Bool
	go func() {
		if delay, _ := strconv.Atoi(os.Getenv("AC_NATIVE_BOOT_DELAY_MS")); delay > 0 && delay <= 10000 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		engine, engineError = newVLC(directory, surface)
		if shuttingDown.Load() && engine != nil {
			engine.Close()
		}
		close(engineLoaded)
	}()
	loading := errors.New("lecteur en cours de chargement")
	currentEngine := func() (*vlcEngine, error) {
		select {
		case <-engineLoaded:
			return engine, engineError
		default:
			return nil, loading
		}
	}
	defer func() {
		shuttingDown.Store(true)
		if engine, err := currentEngine(); err == nil && engine != nil {
			engine.Close()
		}
	}()
	pendingVolume := 100
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		return false, err
	}
	token := hex.EncodeToString(key[:])
	valid := func(candidate string) error {
		if candidate != token {
			return errors.New("session native invalide")
		}
		return nil
	}
	currentPath := ""
	// Save from Go as well as the HTML controls: closing the native window
	// must preserve progress even when JavaScript has already been torn down.
	save := func() {
		if currentPath == "" {
			return
		}
		engine, err := currentEngine()
		if err != nil || engine == nil {
			return
		}
		if state, err := engine.Status(); err == nil && state.Duration > 0 {
			_ = srv.RecordNativeProgress(currentPath, state.Position, state.Duration, state.Part, state.Part+1 == state.Parts && state.State == 6)
		}
	}
	defer save()
	// WebView2's window procedure destroys child windows on WM_CLOSE. Stop
	// decoding first: LibVLC requires its HWND to live until playback stops.
	var previousProc uintptr
	callWindowProc := user32.NewProc("CallWindowProcW")
	closeProc := syscall.NewCallback(func(window uintptr, message uint32, wp, lp uintptr) uintptr {
		if message == 0x0010 {
			save()
			currentPath = ""
			if engine, err := currentEngine(); err == nil && engine != nil {
				_ = engine.Command("stop", 0)
			}
			shuttingDown.Store(true)
		}
		result, _, _ := callWindowProc.Call(previousProc, window, uintptr(message), wp, lp)
		return result
	})
	previousProc, _, _ = setWindowLong.Call(hwnd, ^uintptr(3), closeProc)
	defer setWindowLong.Call(hwnd, ^uintptr(3), previousProc)
	bind := func(name string, fn any) error { return w.Bind(name, fn) }
	if err := bind("_acNativeReady", func(key string) (bool, error) {
		if err := valid(key); err != nil {
			return false, err
		}
		_, err := currentEngine()
		if err == loading {
			return false, nil
		}
		return err == nil, err
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativePlay", func(key, path, subtitle string, resume float64, part int) error {
		if err := valid(key); err != nil {
			return err
		}
		if math.IsNaN(resume) || math.IsInf(resume, 0) || resume < 0 {
			return errors.New("position invalide")
		}
		media, err := srv.ResolveNativePlayback(path, subtitle)
		if err != nil {
			return err
		}
		save()
		engine, err := currentEngine()
		if err != nil {
			return err
		}
		_ = engine.Command("volume", float64(pendingVolume))
		if err := engine.Play(media.Paths, media.Subtitle, resume, part); err != nil {
			return err
		}
		currentPath = path
		return nil
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativeStatus", func(key string) (playbackState, error) {
		if err := valid(key); err != nil {
			return playbackState{}, err
		}
		engine, err := currentEngine()
		if err != nil {
			return playbackState{}, err
		}
		return engine.Status()
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativeCommand", func(key, action string, value float64) error {
		if err := valid(key); err != nil {
			return err
		}
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return errors.New("valeur invalide")
		}
		if action == "volume" {
			pendingVolume = int(math.Max(0, math.Min(100, value)))
		}
		if action == "stop" {
			save()
			currentPath = ""
			showWindow.Call(surface, 0)
		}
		engine, err := currentEngine()
		if err == loading && (action == "stop" || action == "volume") {
			return nil
		}
		if err != nil {
			return err
		}
		return engine.Command(action, value)
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativeSubtitle", func(key, relative string) error {
		if err := valid(key); err != nil {
			return err
		}
		media, err := srv.ResolveNativePlayback(currentPath, relative)
		if err != nil {
			return err
		}
		engine, err := currentEngine()
		if err != nil {
			return err
		}
		return engine.AddSubtitle(media.Subtitle)
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativeBounds", func(key string, x, y, width, height int) error {
		if err := valid(key); err != nil {
			return err
		}
		if width <= 0 || height <= 0 {
			showWindow.Call(surface, 0)
			return nil
		}
		if x < 0 || y < 0 || width > 32768 || height > 32768 {
			return errors.New("dimensions invalides")
		}
		setWindowPos.Call(surface, 0, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0x0010|0x0040)
		return nil
	}); err != nil {
		return false, err
	}
	fullscreen := false
	var restored windowPlacement
	var previousStyle uintptr
	var wasTopmost bool
	if err := bind("_acNativeFullscreen", func(key string) (bool, error) {
		if err := valid(key); err != nil {
			return false, err
		}
		styleIndex := ^uintptr(15) // GWL_STYLE = -16
		if !fullscreen {
			restored.Length = uint32(unsafe.Sizeof(restored))
			if ok, _, err := user32.NewProc("GetWindowPlacement").Call(hwnd, uintptr(unsafe.Pointer(&restored))); ok == 0 {
				return false, err
			}
			previousStyle, _, _ = getWindowLong.Call(hwnd, styleIndex)
			extendedStyle, _, _ := getWindowLong.Call(hwnd, ^uintptr(19))
			wasTopmost = extendedStyle&0x00000008 != 0
			monitor, _, _ := getMonitor.Call(hwnd, 2)
			info := struct {
				Size          uint32
				Monitor, Work rectangle
				Flags         uint32
			}{}
			info.Size = uint32(unsafe.Sizeof(info))
			if ok, _, err := getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); ok == 0 {
				return false, err
			}
			setWindowLong.Call(hwnd, styleIndex, (previousStyle&^(0x00CF0000|0x01000000))|0x80000000|0x02000000)
			// HWND_TOPMOST covers the taskbar as well as the complete monitor.
			setWindowPos.Call(hwnd, ^uintptr(0), uintptr(info.Monitor.Left), uintptr(info.Monitor.Top), uintptr(info.Monitor.Right-info.Monitor.Left), uintptr(info.Monitor.Bottom-info.Monitor.Top), 0x0020|0x0040)
		} else {
			setWindowLong.Call(hwnd, styleIndex, previousStyle)
			order := ^uintptr(1) // HWND_NOTOPMOST
			if wasTopmost {
				order = ^uintptr(0)
			}
			setWindowPos.Call(hwnd, order, 0, 0, 0, 0, 0x0001|0x0002|0x0020)
			user32.NewProc("SetWindowPlacement").Call(hwnd, uintptr(unsafe.Pointer(&restored)))
		}
		fullscreen = !fullscreen
		return fullscreen, nil
	}); err != nil {
		return false, err
	}
	if err := bind("_acNativeExternal", func(key, target string) error {
		if err := valid(key); err != nil {
			return err
		}
		parsed, err := url.Parse(target)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return errors.New("lien invalide")
		}
		player.OpenBrowser(target)
		return nil
	}); err != nil {
		return false, err
	}
	// Bindings exist in WebView2 documents; the unguessable session key is
	// delivered only to the local app page, never injected into other origins.
	origin, _ := json.Marshal(baseURL)
	w.Init(`window.acNativeAvailable = location.origin === ` + string(origin) + `;`)
	w.Navigate(baseURL + "/?native=" + token)
	// Video occupies a native child window, so DOM mousemove events do not
	// cover it. Observe cursor motion only while our app is foreground; the
	// page can then reveal its cinema controls when the mouse moves on video.
	quitPointer := make(chan struct{})
	defer close(quitPointer)
	go func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		defer ticker.Stop()
		var previous point
		getCursor := user32.NewProc("GetCursorPos")
		foreground := user32.NewProc("GetForegroundWindow")
		for {
			select {
			case <-quitPointer:
				return
			case <-ticker.C:
				front, _, _ := foreground.Call()
				if front != hwnd {
					continue
				}
				var cursor point
				if ok, _, _ := getCursor.Call(uintptr(unsafe.Pointer(&cursor))); ok == 0 || cursor == previous {
					continue
				}
				previous = cursor
				user32.NewProc("ScreenToClient").Call(hwnd, uintptr(unsafe.Pointer(&cursor)))
				var client rectangle
				getClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))
				if cursor.X < 0 || cursor.Y < 0 || cursor.X >= client.Right || cursor.Y >= client.Bottom {
					continue
				}
				w.Dispatch(func() { w.Eval(`window._acNativePointer?.()`) })
			}
		}
	}()
	w.Run()
	return true, nil
}
