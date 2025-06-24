package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bytebros.ti/database"
	"bytebros.ti/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado - usando variáveis de ambiente do sistema")
	}

	gin.SetMode(gin.DebugMode) // Manter o modo debug é útil

	database.InitDB()
	defer database.CloseDB()

	if err := database.CreateTables(); err != nil {
		log.Fatalf("Erro ao criar tabelas: %v", err)
	}

	handlers.InitializeGeminiClient()
	log.SetOutput(os.Stderr)

	router := gin.Default()
	router.RedirectTrailingSlash = false

	// Configuração de CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://bytebros.netlify.app"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	router.Use(cors.New(config))

	router.Use(func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	})

	// --- ROTAS ORIGINAIS ---
	router.GET("/api/noticias", handlers.ListarNoticias)
	router.GET("/api/noticias/:id", handlers.ObterNoticia)

	produtoRoutes := router.Group("/api/produtos")
	{
		produtoRoutes.POST("", handlers.CriarProduto)
		produtoRoutes.GET("", handlers.ListarProdutos)
		produtoRoutes.GET("/:id", handlers.ObterProduto)
		produtoRoutes.PUT("/:id", handlers.AtualizarProduto)
		produtoRoutes.DELETE("/:id", handlers.DeletarProduto)
	}

	orcamentoRoutes := router.Group("/api/orcamentos")
	{
		orcamentoRoutes.POST("", handlers.CriarOrcamento)
	}

	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/registrar", handlers.RegistrarUsuario)
		authRoutes.POST("/login", handlers.LoginUsuario)
		authRoutes.POST("/funcionarios/registrar", handlers.RegistrarFuncionario)
		authRoutes.POST("/funcionarios/login", handlers.LoginFuncionario)
	}

    // ... (COLE O RESTANTE DAS SUAS ROTAS AQUI, COMO ESTAVAM ORIGINALMENTE)
    protected := router.Group("/api")
	protected.Use(handlers.AuthMiddleware())
	{
		// ...
	}
    // ... etc

	// Handler para rotas não encontradas (manter é útil para depuração)
	router.NoRoute(func(c *gin.Context) {
		log.Printf("DEBUG: Rota não encontrada. Método: %s, Caminho: %s", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusNotFound, gin.H{"erro": "Rota não encontrada", "caminho_requisitado": c.Request.URL.Path})
	})

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Servidor iniciado na porta %s", os.Getenv("PORT"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Falha ao iniciar servidor: %v", err)
		}
	}()

	<-quit
	log.Println("Desligando servidor...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Falha ao desligar servidor: %v", err)
	}

	log.Println("Servidor desligado com sucesso")
}
