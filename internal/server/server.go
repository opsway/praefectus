package server

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/boodmo/praefectus/internal/config"
	"github.com/boodmo/praefectus/internal/storage"
)

type Server struct {
	config       *config.Config
	stateStorage *storage.ProcStorage
}

func New(cfg *config.Config, ps *storage.ProcStorage) *Server {
	return &Server{
		config:       cfg,
		stateStorage: ps,
	}
}

func (s *Server) Start() {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		wsList := s.stateStorage.GetList()
		result := make([]map[string]interface{}, 0, len(wsList))
		for _, ws := range wsList {
			result = append(result, map[string]interface{}{
				"pid":        ws.Process.Pid,
				"state":      storage.StateName(ws.State),
				"updated_at": ws.UpdatedAt.Format(time.RFC3339),
			})
		}

		return ctx.JSON(result)
	})

	addr := net.JoinHostPort(s.config.Server.Host, strconv.Itoa(s.config.Server.Port))
	log.Fatal(app.Listen(addr))
}
