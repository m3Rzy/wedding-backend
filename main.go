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

	// Запускаем HTTP сервер
	log.Println("HTTP сервер запущен на порту 8080")
	log.Println("Server has been started successfully!")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
