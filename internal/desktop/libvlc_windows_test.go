//go:build windows && amd64

package desktop

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"absolute_cinema/internal/player"
)

// Opt-in integration test uses a real local movie and real LibVLC, without
// opening an external player or changing the source media.
func TestLibVLCPlayback(t *testing.T) {
	media := os.Getenv("AC_LIBVLC_TEST_MEDIA")
	if media == "" {
		t.Skip("set AC_LIBVLC_TEST_MEDIA to test real decoding")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	class, _ := syscall.UTF16PtrFromString("STATIC")
	hwnd, _, err := createWindow.Call(0, uintptr(unsafe.Pointer(class)), 0, 0x06000000, 0, 0, 640, 360, 0, 0, 0, 0)
	if hwnd == 0 {
		t.Fatal(err)
	}
	defer user32.NewProc("DestroyWindow").Call(hwnd)
	engine, err := newVLC(player.LibraryDirectory(""), hwnd)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	if err := engine.Play([]string{media}, "", 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := engine.Command("volume", 0); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(20 * time.Second)
	var status playbackState
	for time.Now().Before(deadline) {
		status, err = engine.Status()
		if err != nil {
			t.Fatal(err)
		}
		if status.State == 7 {
			t.Fatal("LibVLC decoding error")
		}
		if status.State == 3 && status.Position > 1 && status.Duration > 0 && status.VideoWidth > 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if status.Position <= 1 || status.Duration <= 0 {
		t.Fatalf("playback did not advance: %#v", status)
	}
	if status.VideoWidth == 0 || status.VideoHeight == 0 {
		t.Fatalf("no native video output: %#v", status)
	}
	if status.Volume != 0 {
		t.Fatalf("volume not retained: %#v", status)
	}
	if err := engine.Command("pause", 1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	paused, _ := engine.Status()
	if paused.State != 4 {
		t.Fatalf("not paused: %#v", paused)
	}
	if err := engine.Command("seek", 120); err != nil {
		t.Fatal(err)
	}
	if err := engine.Command("pause", 0); err != nil {
		t.Fatal(err)
	}
	time.Sleep(400 * time.Millisecond)
	seeked, _ := engine.Status()
	if seeked.Position < 119 {
		t.Fatalf("seek failed: %#v", seeked)
	}
	if len(seeked.Audio) > 1 {
		if err := engine.Command("audio", float64(seeked.Audio[1].ID)); err != nil {
			t.Fatal(err)
		}
	}
	if err := engine.Command("subtitle", -1); err != nil {
		t.Fatal(err)
	}
	disabled, _ := engine.Status()
	if disabled.SubtitleTrack != -1 {
		t.Fatal("subtitle disabling failed")
	}
	external := filepath.Join(t.TempDir(), "external subtitle.srt")
	if err := os.WriteFile(external, []byte("1\n00:02:00,000 --> 00:02:05,000\nNative LibVLC subtitle\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := engine.AddSubtitle(external); err != nil {
		t.Fatal(err)
	}
	withExternal, _ := engine.Status()
	for until := time.Now().Add(5 * time.Second); len(withExternal.Subtitles) <= len(seeked.Subtitles) && time.Now().Before(until); {
		time.Sleep(100 * time.Millisecond)
		withExternal, _ = engine.Status()
	}
	if len(withExternal.Subtitles) <= len(seeked.Subtitles) {
		t.Fatalf("external subtitle track not added: before=%#v after=%#v path=%s", seeked.Subtitles, withExternal.Subtitles, external)
	}
	t.Logf("Real LibVLC playback: duration %.1fs, time %.1fs, audio tracks %d, subtitle tracks %d", seeked.Duration, seeked.Position, len(seeked.Audio), len(seeked.Subtitles))
}
