package server

import (
	"fmt"
	"os"
	"queue/queue"
	"queue/store"
	"strconv"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

type HttpServer struct {
	server *fiber.App
	store  store.Store
	queues map[string]queue.Queue
	mu     sync.Mutex
}

// NewHttpServer cria um novo servidor HTTP com dependências injetáveis
func NewHttpServer(s store.Store) *HttpServer {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		AppName:       "Queue",
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	app.Use(logger.New())

	return &HttpServer{
		server: app,
		store:  s,
		queues: make(map[string]queue.Queue, 0), // queueName => queue
	}
}

// NewHttpServerWithConfig permite customizar o logger e outras configurações
func NewHttpServerWithConfig(s store.Store, enableLogger bool) *HttpServer {
	app := fiber.New(fiber.Config{
		CaseSensitive: true,
		StrictRouting: true,
		AppName:       "Queue",
	})

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	if enableLogger {
		app.Use(logger.New())
	}

	return &HttpServer{
		server: app,
		store:  s,
		queues: make(map[string]queue.Queue, 0),
	}
}

func (s *HttpServer) Start() error {
	port := os.Getenv("API_PORT")

	if port == "" {
		return fmt.Errorf("failed to start http server, port is empty")
	}

	s.configRoutes()

	err := s.server.Listen(":" + port)
	if err != nil {
		return fmt.Errorf("failed to start http server on port %s, error: %v ", port, err)
	}

	return nil
}

func (s *HttpServer) getQueue(queueName string) queue.Queue {
	if s.queues[queueName] == nil {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.queues[queueName] = queue.NewSequentialQueue(queue.QueueConfig{
			Store: s.store,
			Name:  queueName,
		})
	}

	return s.queues[queueName]
}

func (s *HttpServer) configRoutes() {
	api := s.server.Group("/api/v1", func(c fiber.Ctx) error { return c.Next() })

	api.Post("/queue/:queueName/enqueue", s.enqueueHandler)
	api.Get("/queue/:queueName/dequeue", s.dequeueHandler)
	api.Get("/queue/:queueName/list-items", s.listItemsHandler)
}

func (s *HttpServer) enqueueHandler(c fiber.Ctx) error {
	queueName := c.Params("queueName")
	if queueName == "" {
		return c.Status(400).SendString("queueName param is required")
	}

	queue := s.getQueue(queueName)

	index, err := queue.Enqueue(string(c.Body()))
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	response, err := sonic.Marshal(map[string]any{
		"queueLength": index,
	})
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(200).Send(response)
}

func (s *HttpServer) dequeueHandler(c fiber.Ctx) error {
	queueName := c.Params("queueName")

	if queueName == "" {
		return c.Status(400).SendString("queueName param is required")
	}

	queue := s.getQueue(queueName)

	data, err := queue.Dequeue()

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	if data == nil {
		return c.SendStatus(200)
	}

	return c.Status(200).SendString(data.Payload)
}

func (s *HttpServer) listItemsHandler(c fiber.Ctx) error {
	queueName := c.Params("queueName")

	if queueName == "" {
		return c.Status(fiber.StatusBadRequest).SendString("The 'queueName' parameter is required")
	}

	page := c.Query("page")
	if page == "" {
		return c.Status(fiber.StatusBadRequest).SendString("The 'page' query parameter is required")
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to convert 'page' to integer")
	}

	if pageInt < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("The 'page' query parameter must be greater than 0")
	}

	perPage := c.Query("perPage")
	if perPage == "" {
		return c.Status(fiber.StatusBadRequest).SendString("The 'perPage' query parameter is required")
	}

	perPageInt, err := strconv.Atoi(perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to convert 'perPage' to integer")
	}

	if perPageInt < 1 {
		return c.Status(fiber.StatusBadRequest).SendString("The 'perPage' query parameter must be greater than 0")
	}

	queue := s.getQueue(queueName)

	response, err := queue.ListItems(store.ListOptions{
		Page:    pageInt,
		PerPage: perPageInt,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to list items from the queue")
	}

	return c.JSON(response)
}
