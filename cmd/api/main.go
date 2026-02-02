package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"silsilah-keluarga/internal/config"
	"silsilah-keluarga/internal/handler"
	"silsilah-keluarga/internal/middleware"
	"silsilah-keluarga/internal/repository"
	"silsilah-keluarga/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	db, err := config.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	redis, err := config.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	minioClient, err := config.NewMinIOClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to MinIO: %v (media upload will not work)", err)
	}

	repos := repository.NewRepositories(db)
	services := service.NewServices(repos, redis, minioClient, cfg)
	handlers := handler.NewHandlers(services)

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.CORSOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	// Add middleware to extract real IP (for Cloudflare) and User-Agent
	app.Use(middleware.RequestInfo())

	setupRoutes(app, handlers, services.Auth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRoutes(app *fiber.App, h *handler.Handlers, authService service.AuthService) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	v1 := app.Group("/api/v1")

	public := v1.Group("/public")
	public.Get("/graph", h.Public.GetGraph)
	public.Get("/persons/:personId", h.Public.GetPerson)

	auth := v1.Group("/auth")
	auth.Post("/register", h.Auth.Register)
	auth.Post("/login", h.Auth.Login)
	auth.Post("/refresh", h.Auth.RefreshToken)
	auth.Post("/forgot-password", h.Auth.ForgotPassword)
	auth.Post("/reset-password", h.Auth.ResetPassword)
	auth.Get("/verify-email", h.Auth.VerifyEmail)
	auth.Post("/resend-verification", h.Auth.ResendVerificationEmail)

	protected := v1.Group("", middleware.AuthRequired(authService))

	users := protected.Group("/users")
	users.Get("/me", h.User.GetProfile)
	users.Get("/me/ancestors", h.User.GetAncestors)
	users.Put("/me", h.User.UpdateProfile)
	users.Post("/assign-role", middleware.RequireRole("developer"), h.User.AssignRole)
	users.Get("/by-role/:role", h.User.ListByRole)
	users.Get("/role-users", h.User.GetRoleUsers)
	users.Get("/", middleware.RequireRole("developer"), h.User.GetAllUsers)
	users.Delete("/:id", middleware.RequireRole("developer"), h.User.DeleteUser)

	persons := protected.Group("/persons")
	persons.Post("/", middleware.RequireRole("editor"), h.Person.Create)
	persons.Get("/", h.Person.List)
	persons.Get("/search", h.Person.Search)
	persons.Get("/:personId", h.Person.Get)
	persons.Put("/:personId", middleware.RequireRole("editor"), h.Person.Update)
	persons.Delete("/:personId", middleware.RequireRole("editor"), h.Person.Delete)

	relationships := protected.Group("/relationships")
	relationships.Post("/", middleware.RequireRole("editor"), h.Relationship.Create)
	relationships.Get("/", h.Relationship.List)
	relationships.Get("/:relationshipId", h.Relationship.Get)
	relationships.Put("/:relationshipId", middleware.RequireRole("editor"), h.Relationship.Update)
	relationships.Delete("/:relationshipId", middleware.RequireRole("editor"), h.Relationship.Delete)

	graph := protected.Group("/graph")
	graph.Get("/", h.Graph.GetFullGraph)
	graph.Get("/ancestors/:personId", h.Graph.GetAncestors)
	graph.Get("/ancestors/:personId/split", h.Graph.GetSplitAncestors)
	graph.Get("/descendants/:personId", h.Graph.GetDescendants)
	graph.Get("/path", h.Graph.FindRelationshipPath)

	changeRequests := protected.Group("/change-requests")
	changeRequests.Post("/", h.ChangeRequest.Create)
	changeRequests.Get("/", h.ChangeRequest.List)
	changeRequests.Get("/:requestId", h.ChangeRequest.Get)
	changeRequests.Post("/:requestId/approve", h.ChangeRequest.Approve)
	changeRequests.Post("/:requestId/reject", h.ChangeRequest.Reject)

	media := protected.Group("/media")
	media.Post("/", middleware.RequireRole("member"), h.Media.Upload)
	media.Get("/", h.Media.List)
	media.Get("/:mediaId", h.Media.Get)
	media.Delete("/:mediaId", middleware.RequireRole("member"), h.Media.Delete)

	comments := protected.Group("/persons/:personId/comments")
	comments.Post("/", h.Comment.Create)
	comments.Get("/", h.Comment.List)
	comments.Put("/:commentId", h.Comment.Update)
	comments.Delete("/:commentId", h.Comment.Delete)

	notifications := protected.Group("/notifications")
	notifications.Get("/", h.Notification.List)
	notifications.Get("/unread-count", h.Notification.GetUnreadCount)
	notifications.Patch("/:id/read", h.Notification.MarkAsRead)
	notifications.Post("/mark-all-read", h.Notification.MarkAllAsRead)

	audit := protected.Group("/audit")
	audit.Get("/recent", h.Audit.GetRecentActivities)
}
