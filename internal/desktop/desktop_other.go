//go:build !windows || !amd64

package desktop

import (
	"absolute_cinema/internal/config"
	"absolute_cinema/internal/server"
)

// The native host currently targets Windows amd64. Other platforms retain
// the web interface and external VLC player.
func Run(url string, paths config.Paths, srv *server.Server) (bool, error) { return false, nil }
