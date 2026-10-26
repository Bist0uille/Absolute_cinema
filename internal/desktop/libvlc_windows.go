//go:build windows && amd64

package desktop

import (
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type track struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type playbackState struct {
	State         int     `json:"state"`
	Position      float64 `json:"position"`
	Duration      float64 `json:"duration"`
	Volume        int     `json:"volume"`
	AudioTrack    int     `json:"audio_track"`
	SubtitleTrack int     `json:"subtitle_track"`
	Audio         []track `json:"audio"`
	Subtitles     []track `json:"subtitles"`
	Part          int     `json:"part"`
	Parts         int     `json:"parts"`
	VideoWidth    uint32  `json:"video_width"`
	VideoHeight   uint32  `json:"video_height"`
}

type vlcEngine struct {
	mu               sync.Mutex
	dll              *windows.DLL
	functions        map[string]*windows.Proc
	instance, player uintptr
	paths            []string
	part             int
	subtitle         string
	volume           int
	volumeApplied    bool
}

func newVLC(directory string, hwnd uintptr) (*vlcEngine, error) {
	dll, err := windows.LoadLibraryEx(filepath.Join(directory, "libvlc.dll"), 0,
		windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR|windows.LOAD_LIBRARY_SEARCH_DEFAULT_DIRS)
	if err != nil {
		return nil, fmt.Errorf("chargement de LibVLC : %w", err)
	}
	engine := &vlcEngine{dll: &windows.DLL{Handle: dll}, functions: map[string]*windows.Proc{}, volume: 100}
	for _, name := range []string{
		"libvlc_new", "libvlc_release", "libvlc_media_new_path", "libvlc_media_add_option", "libvlc_media_release",
		"libvlc_media_player_new", "libvlc_media_player_release", "libvlc_media_player_set_hwnd", "libvlc_media_player_set_media",
		"libvlc_media_player_play", "libvlc_media_player_stop", "libvlc_media_player_set_pause", "libvlc_media_player_get_state",
		"libvlc_media_player_get_time", "libvlc_media_player_get_length", "libvlc_media_player_set_time",
		"libvlc_audio_get_volume", "libvlc_audio_set_volume", "libvlc_audio_get_track", "libvlc_audio_set_track",
		"libvlc_audio_get_track_description", "libvlc_video_get_spu", "libvlc_video_set_spu", "libvlc_video_get_spu_description",
		"libvlc_track_description_list_release", "libvlc_video_set_key_input", "libvlc_video_set_mouse_input", "libvlc_media_player_add_slave",
		"libvlc_video_get_size",
		"libvlc_video_set_aspect_ratio", "libvlc_video_set_crop_geometry",
	} {
		proc, err := engine.dll.FindProc(name)
		if err != nil {
			engine.Close()
			return nil, fmt.Errorf("LibVLC 3 requis (%s) : %w", name, err)
		}
		engine.functions[name] = proc
	}
	options := []string{"--no-video-title-show", "--no-osd", "--quiet", "--file-caching=1000", "--autoscale"}
	arguments := make([]*byte, len(options))
	for i, option := range options {
		arguments[i], _ = syscall.BytePtrFromString(option)
	}
	engine.instance = engine.call("libvlc_new", uintptr(len(arguments)), uintptr(unsafe.Pointer(&arguments[0])))
	runtime.KeepAlive(arguments)
	if engine.instance == 0 {
		engine.Close()
		return nil, fmt.Errorf("initialisation de LibVLC impossible")
	}
	engine.player = engine.call("libvlc_media_player_new", engine.instance)
	if engine.player == 0 {
		engine.Close()
		return nil, fmt.Errorf("création du lecteur LibVLC impossible")
	}
	engine.call("libvlc_media_player_set_hwnd", engine.player, hwnd)
	engine.call("libvlc_video_set_key_input", engine.player, 0)
	engine.call("libvlc_video_set_mouse_input", engine.player, 0)
	engine.call("libvlc_video_set_aspect_ratio", engine.player, 0)
	engine.call("libvlc_video_set_crop_geometry", engine.player, 0)
	return engine, nil
}

func (e *vlcEngine) call(name string, args ...uintptr) uintptr {
	result, _, _ := e.functions[name].Call(args...)
	return result
}

func (e *vlcEngine) stringCall(name, value string, args ...uintptr) uintptr {
	pointer, err := syscall.BytePtrFromString(value)
	if err != nil {
		return ^uintptr(0)
	}
	args = append(args, uintptr(unsafe.Pointer(pointer)))
	result := e.call(name, args...)
	runtime.KeepAlive(pointer)
	return result
}

func (e *vlcEngine) Play(paths []string, subtitle string, resume float64, part int) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(paths) == 0 {
		return fmt.Errorf("aucun média")
	}
	e.call("libvlc_media_player_stop", e.player)
	if part < 0 || part >= len(paths) {
		part = 0
		resume = 0
	}
	e.paths, e.part, e.subtitle = append([]string(nil), paths...), part, subtitle
	return e.playPart(resume)
}

func (e *vlcEngine) playPart(resume float64) error {
	e.volumeApplied = false
	media := e.stringCall("libvlc_media_new_path", e.paths[e.part], e.instance)
	if media == 0 || media == ^uintptr(0) {
		return fmt.Errorf("ouverture du média impossible")
	}
	defer e.call("libvlc_media_release", media)
	if resume > 0 {
		e.stringCall("libvlc_media_add_option", fmt.Sprintf(":start-time=%.3f", resume), media)
	}
	if e.subtitle != "" {
		e.stringCall("libvlc_media_add_option", ":sub-file="+e.subtitle, media)
	}
	e.call("libvlc_media_player_set_media", e.player, media)
	if int32(e.call("libvlc_media_player_play", e.player)) == -1 {
		return fmt.Errorf("lecture LibVLC impossible")
	}
	return nil
}

func (e *vlcEngine) Status() (playbackState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	state := int(e.call("libvlc_media_player_get_state", e.player))
	if state == 6 && e.part+1 < len(e.paths) {
		e.part++
		if err := e.playPart(0); err != nil {
			return playbackState{}, err
		}
		state = 1
	}
	if (state == 3 || state == 4) && !e.volumeApplied {
		e.volumeApplied = int32(e.call("libvlc_audio_set_volume", e.player, uintptr(e.volume))) == 0
	}
	var width, height uint32
	e.call("libvlc_video_get_size", e.player, 0, uintptr(unsafe.Pointer(&width)), uintptr(unsafe.Pointer(&height)))
	return playbackState{
		State:         state,
		Position:      float64(int64(e.call("libvlc_media_player_get_time", e.player))) / 1000,
		Duration:      float64(int64(e.call("libvlc_media_player_get_length", e.player))) / 1000,
		Volume:        int(int32(e.call("libvlc_audio_get_volume", e.player))),
		AudioTrack:    int(int32(e.call("libvlc_audio_get_track", e.player))),
		SubtitleTrack: int(int32(e.call("libvlc_video_get_spu", e.player))),
		Audio:         e.tracks("libvlc_audio_get_track_description"),
		Subtitles:     e.tracks("libvlc_video_get_spu_description"),
		Part:          e.part, Parts: len(e.paths),
		VideoWidth: width, VideoHeight: height,
	}, nil
}

func (e *vlcEngine) tracks(function string) []track {
	first := e.call(function, e.player)
	if first == 0 {
		return []track{}
	}
	defer e.call("libvlc_track_description_list_release", first)
	result := []track{}
	type description struct {
		ID   int32
		_    uint32
		Name *byte
		Next uintptr
	}
	for current, count := first, 0; current != 0 && count < 256; count++ {
		desc := (*description)(unsafe.Pointer(current))
		result = append(result, track{int(desc.ID), windows.BytePtrToString(desc.Name)})
		current = desc.Next
	}
	return result
}

func (e *vlcEngine) Command(action string, value float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	switch action {
	case "pause":
		e.call("libvlc_media_player_set_pause", e.player, uintptr(int(value)))
	case "seek":
		if value < 0 {
			value = 0
		}
		length := int64(e.call("libvlc_media_player_get_length", e.player))
		millis := int64(value * 1000)
		if length > 0 && millis > length {
			millis = length
		}
		e.call("libvlc_media_player_set_time", e.player, uintptr(millis))
	case "volume":
		if value < 0 {
			value = 0
		}
		if value > 100 {
			value = 100
		}
		e.volume = int(value)
		e.volumeApplied = int32(e.call("libvlc_audio_set_volume", e.player, uintptr(e.volume))) == 0
	case "audio":
		if int32(e.call("libvlc_audio_set_track", e.player, uintptr(int(value)))) != 0 {
			return fmt.Errorf("piste audio indisponible")
		}
	case "subtitle":
		if int32(e.call("libvlc_video_set_spu", e.player, uintptr(int(value)))) != 0 {
			return fmt.Errorf("sous-titre indisponible")
		}
	case "stop":
		e.paths = nil
		e.call("libvlc_media_player_stop", e.player)
	default:
		return fmt.Errorf("commande inconnue")
	}
	return nil
}

func (e *vlcEngine) AddSubtitle(path string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	uri := (&url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(path)}).String()
	pointer, _ := syscall.BytePtrFromString(uri)
	result := int32(e.call("libvlc_media_player_add_slave", e.player, 0, uintptr(unsafe.Pointer(pointer)), 1))
	runtime.KeepAlive(pointer)
	if result != 0 {
		return fmt.Errorf("chargement du sous-titre impossible")
	}
	return nil
}

func (e *vlcEngine) Close() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.player != 0 {
		e.call("libvlc_media_player_stop", e.player)
		e.call("libvlc_media_player_release", e.player)
		e.player = 0
	}
	if e.instance != 0 {
		e.call("libvlc_release", e.instance)
		e.instance = 0
	}
	if e.dll != nil {
		e.dll.Release()
		e.dll = nil
	}
}
