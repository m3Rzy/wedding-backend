# Запуск на проде

docker stop go-backend
docker rm go-backend
docker rmi go-backend
docker build -t go-backend .
docker run -d \
  --name go-backend \
  --restart unless-stopped \
  -p 8080:8080 \
  go-backend
docker logs -f go-backend
