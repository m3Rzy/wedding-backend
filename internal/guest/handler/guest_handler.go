package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"wedding-backend/internal/bot" // Импортируем пакет бота
	"wedding-backend/internal/guest/dto"
)

// BotAppInstance глобальная переменная для доступа к боту из хендлеров
var BotAppInstance *bot.BotApp

// InitBotApp инициализирует бота для использования в хендлерах
func InitBotApp(botApp *bot.BotApp) {
	BotAppInstance = botApp
	log.Println("Бот инициализирован для использования в хендлерах")
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Это простой текстовый ответ от сервера!")
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprint(w, "Добро пожаловать на сервер! Перейдите на /get")
}

// Обработчик для POST-запроса с данными контакта
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Ожидается application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Декодируем JSON данные
	var contact dto.GuestDto
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&contact)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	log.Println(contact)

	// Валидация данных
	if contact.Fio == "" || contact.Telephone == "" {
		http.Error(w, "ФИО и номер телефона обязательны", http.StatusBadRequest)
		return
	}

	// Выводим в консоль
	log.Printf("Получены данные: ФИО='%s', Телефон='%s'", contact.Fio, contact.Telephone)

	// Отправляем данные в телеграм бот
	if BotAppInstance != nil {
		err := BotAppInstance.SendUserData(contact.Fio, contact.Telephone, contact.Transport, contact.CarNumber)
		if err != nil {
			log.Printf("Ошибка отправки в Telegram: %v", err)
			http.Error(w, "Ошибка отправки уведомления", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("Бот не инициализирован, данные не отправлены в Telegram")
		http.Error(w, "Сервис уведомлений недоступен", http.StatusInternalServerError)
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"status":  "success",
		"message": "Данные успешно получены и отправлены администраторам",
	}
	json.NewEncoder(w).Encode(response)
}