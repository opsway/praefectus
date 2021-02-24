package server

import (
	"log"

	"github.com/boodmo/praefectus/internal/storage"
)
import "github.com/gofiber/fiber/v2"

type Server struct {
	stateStorage *storage.ProcStorage
}

func New(ps *storage.ProcStorage) *Server {
	return &Server{stateStorage: ps}
}

func (s *Server) Start() {
	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		wsList := s.stateStorage.GetList()
		result := make([]map[string]interface{}, 0, len(wsList))
		for _, ws := range wsList {
			result = append(result, map[string]interface{}{
				"pid":   ws.Process.Pid,
				"state": storage.StateName(ws.State),
			})
		}

		return ctx.JSON(result)
	})

	log.Fatal(app.Listen("0.0.0.0:9000"))
}
