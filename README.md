# Запуск на проде

# Останавливаем контейнер (если запущен)
docker stop go-backend

# Удаляем контейнер
docker rm go-backend

# Удаляем образ
docker rmi go-backend

# Собираем новый образ
docker build -t go-backend .

# Запускаем новый контейнер
docker run -d \
  --name go-backend \
  --restart unless-stopped \
  -p 8080:8080 \
  go-backend

# Проверяем логи
docker logs -f go-backend
