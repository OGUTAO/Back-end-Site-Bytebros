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

	// ---> ADIÇÃO 1: HABILITA LOGS DETALHADOS DO GIN
	gin.SetMode(gin.DebugMode)

	database.InitDB()
	defer database.CloseDB()

	if err := database.CreateTables(); err != nil {
		log.Fatalf("Erro ao criar tabelas: %v", err)
	}

	handlers.InitializeGeminiClient()
	log.SetOutput(os.Stderr)

	router := gin.Default()

	// ... (Sua configuração de CORS que já está funcionando)
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://bytebros.netlify.app"} // Mantém esta linha que já corrigimos
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	router.Use(cors.New(config))
	
	// ... (Todas as suas definições de rotas como /api/noticias, /api/auth, etc.)
    // ...
	authRoutes := router.Group("/api/auth")
	{
		authRoutes.POST("/registrar", handlers.RegistrarUsuario)
		authRoutes.POST("/login", handlers.LoginUsuario)
        //...
	}
    // ...

	// ---> ADIÇÃO 2: CAPTURA QUALQUER ROTA NÃO ENCONTRADA PARA DEPURAÇÃO
	router.NoRoute(func(c *gin.Context) {
		log.Printf("DEBUG: Rota não encontrada. Método: %s, Caminho: %s", c.Request.Method, c.Request.URL.Path)
		c.JSON(http.StatusNotFound, gin.H{"erro": "Rota não encontrada", "caminho_requisitado": c.Request.URL.Path})
	})

	server := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	// ... (Resto do seu código para desligar o servidor)
}
