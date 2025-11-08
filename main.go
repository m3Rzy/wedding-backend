package main

import (
	"log"
	"net/http"
	"wedding-backend/internal/bot"
	"wedding-backend/internal/guest/handler"
)

// Middleware для разрешения CORS
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разрешаем запросы с любого origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Разрешаем методы
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		// Разрешаем заголовки
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Обрабатываем preflight OPTIONS запрос
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем управление следующему обработчику
		next(w, r)
	}
}

func main() {
	// Загружаем конфигурацию бота
	config, err := bot.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем приложение бота
	botApp, err := bot.NewBotApp(config)
	if err != nil {
		log.Fatalf("Ошибка создания бота: %v", err)
	}

	// Инициализируем бота в хендлерах
	handler.InitBotApp(botApp)

	// Запускаем бота в отдельной горутине
	go func() {
		log.Println("Запуск телеграм бота...")
		botApp.Start()
	}()

	// Создаем новый роутер с CORS middleware
	mux := http.NewServeMux()

	// Настраиваем HTTP маршруты с CORS
	mux.HandleFunc("/get", enableCORS(handler.GetHandler))
	mux.HandleFunc("/", enableCORS(handler.ContactHandler))

	// ЗАПУСК С HTTPS НА ПОРТУ 443
	log.Println("HTTPS сервер запущен на порту 443")
	log.Println("Server has been started successfully!")
	
	// Для порта 443 нужны права суперпользователя на Linux
	err = http.ListenAndServeTLS(":443", "server.crt", "server.key", mux)
	if err != nil {
		log.Fatal("Ошибка запуска HTTPS сервера:", err)
	}
}