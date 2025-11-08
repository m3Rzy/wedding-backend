package bot

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

type Config struct {
	BotToken     string
	AllowedUsers []string
	AdminChatIDs []int64
}

type BotApp struct {
	bot    *telebot.Bot
	config *Config
}

// LoadConfig загружает конфигурацию из .env файла
func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	allowedUsersStr := os.Getenv("ALLOWED_USERS")
	if allowedUsersStr == "" {
		return nil, fmt.Errorf("ALLOWED_USERS is required")
	}

	allowedUsers := strings.Split(allowedUsersStr, ",")
	for i, user := range allowedUsers {
		allowedUsers[i] = strings.TrimSpace(strings.TrimPrefix(user, "@"))
	}

	// Пытаемся загрузить Chat IDs если они есть
	var adminChatIDs []int64
	if chatIDsStr := os.Getenv("ADMIN_CHAT_IDS"); chatIDsStr != "" {
		chatIDStrings := strings.Split(chatIDsStr, ",")
		for _, chatIDStr := range chatIDStrings {
			chatIDStr = strings.TrimSpace(chatIDStr)
			chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
			if err == nil {
				adminChatIDs = append(adminChatIDs, chatID)
			}
		}
	}

	return &Config{
		BotToken:     botToken,
		AllowedUsers: allowedUsers,
		AdminChatIDs: adminChatIDs,
	}, nil
}

// NewBotApp создает новое приложение бота
func NewBotApp(config *Config) (*BotApp, error) {
	pref := telebot.Settings{
		Token:  config.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &BotApp{
		bot:    bot,
		config: config,
	}, nil
}

// SendUserData отправляет данные пользователя всем администраторам
func (app *BotApp) SendUserData(fio string, telephone string, transport string, carNumber string) error {
	// Формируем сообщение в зависимости от типа транспорта
	var transportInfo string
	if transport == "car" && carNumber != "" {
		transportInfo = fmt.Sprintf("🚗 Личный автомобиль\n🚙 Госномер: %s", carNumber)
	} else {
		transportInfo = "🚌 Трансфер"
	}

	userData := fmt.Sprintf(
		"📨 Новый гость:\n"+
			"👤 ФИО: %s\n"+
			"📞 Телефон: %s\n"+
			"📍 Транспорт: %s\n"+
			"📅 Время: %s",
		fio,
		telephone,
		transportInfo,
		time.Now().Format("2006-01-02 15:04:05"),
	)

	log.Printf("Отправка данных в Telegram: %s, %s, транспорт: %s, номер: %s", 
		fio, telephone, transport, carNumber)

	var successCount int
	var errors []string

	// Если есть зарегистрированные Chat IDs, отправляем по ним
	if len(app.config.AdminChatIDs) > 0 {
		for _, adminChatID := range app.config.AdminChatIDs {
			recipient := &telebot.Chat{ID: adminChatID}
			_, err := app.bot.Send(recipient, userData)
			if err != nil {
				errorMsg := fmt.Sprintf("Ошибка отправки сообщения администратору (Chat ID: %d): %v", adminChatID, err)
				log.Printf(errorMsg)
				errors = append(errors, errorMsg)
			} else {
				log.Printf("Сообщение успешно отправлено администратору (Chat ID: %d)", adminChatID)
				successCount++
			}
		}
	} else {
		// Если Chat IDs нет, пытаемся отправить по username (может не работать)
		log.Printf("Chat IDs не зарегистрированы, пытаемся отправить по username")
		for _, username := range app.config.AllowedUsers {
			recipient := &telebot.User{Username: username}
			_, err := app.bot.Send(recipient, userData)
			if err != nil {
				errorMsg := fmt.Sprintf("Ошибка отправки сообщения пользователю %s: %v", username, err)
				log.Printf(errorMsg)
				errors = append(errors, errorMsg)
			} else {
				log.Printf("Сообщение успешно отправлено пользователю %s", username)
				successCount++
			}
		}
	}

	if successCount == 0 && len(errors) > 0 {
		return fmt.Errorf("не удалось отправить ни одного сообщения: %v", errors)
	}

	if len(errors) > 0 {
		log.Printf("Успешно отправлено %d сообщений, ошибок: %d", successCount, len(errors))
	}

	return nil
}

// GetBot возвращает экземпляр бота для использования в хендлерах
func (app *BotApp) GetBot() *telebot.Bot {
	return app.bot
}

// isAdmin проверяет, является ли пользователь администратором
func (app *BotApp) isAdmin(user *telebot.User) bool {
	username := strings.TrimPrefix(user.Username, "@")
	for _, adminUser := range app.config.AllowedUsers {
		if strings.EqualFold(username, adminUser) {
			return true
		}
	}
	return false
}

// registerAdmin регистрирует Chat ID администратора
func (app *BotApp) registerAdmin(user *telebot.User, chatID int64) bool {
	if app.isAdmin(user) {
		// Проверяем, нет ли уже этого Chat ID
		for _, existingID := range app.config.AdminChatIDs {
			if existingID == chatID {
				return false // Уже зарегистрирован
			}
		}
		app.config.AdminChatIDs = append(app.config.AdminChatIDs, chatID)
		log.Printf("Зарегистрирован Chat ID %d для администратора @%s", chatID, user.Username)
		return true
	}
	return false
}

// Start запускает бота
func (app *BotApp) Start() {
	log.Printf("Бот запущен! Ожидаем сообщения...")
	log.Printf("Разрешенные пользователи: %v", app.config.AllowedUsers)
	log.Printf("Зарегистрированные Chat IDs: %v", app.config.AdminChatIDs)

	// Команда /start
	app.bot.Handle("/start", func(ctx telebot.Context) error {
		user := ctx.Sender()
		chatID := ctx.Chat().ID

		// Если пользователь администратор, регистрируем его Chat ID
		if app.isAdmin(user) {
			if app.registerAdmin(user, chatID) {
				return ctx.Send(fmt.Sprintf(
					"👋 Привет, администратор @%s!\n"+
						"✅ Ваш Chat ID (%d) зарегистрирован!\n"+
						"Теперь вы будете получать уведомления о новых сообщениях.",
					user.Username, chatID,
				))
			}
			return ctx.Send(fmt.Sprintf(
				"👋 Привет, администратор @%s!\n"+
					"✅ Вы уже зарегистрированы для получения уведомлений.",
				user.Username,
			))
		}

		// Обычный пользователь
		return ctx.Send(
			"👋 Привет! Я бот для передачи сообщений администраторам.\n" +
				"Просто отправьте мне любое сообщение, и я перешлю его ответственным лицам.\n\n" +
				"📝 Отправьте текст, фото или файл - всё будет доставлено!",
		)
	})

	// Команда /chatid для получения Chat ID
	app.bot.Handle("/chatid", func(ctx telebot.Context) error {
		chat := ctx.Chat()
		user := ctx.Sender()

		response := fmt.Sprintf(
			"📋 Ваши ID:\n"+
				"💬 Chat ID: `%d`\n"+
				"👤 User ID: `%d`\n"+
				"🔹 Username: @%s",
			chat.ID,
			user.ID,
			user.Username,
		)

		return ctx.Send(response, &telebot.SendOptions{
			ParseMode: telebot.ModeMarkdown,
		})
	})

	// Запускаем бота
	app.bot.Start()
}