#!/bin/bash

echo "========================================"
echo "Запуск системы Календарь с Kafka"
echo "========================================"
echo ""

cd "$(dirname "$0")/.."

echo "[1/3] Сборка проектов..."
go build -o bin/calendar -ldflags "-X main.release=develop" ./cmd/calendar
if [ $? -ne 0 ]; then
    echo "Ошибка при сборке calendar"
    exit 1
fi
go build -o bin/calendar_scheduler -ldflags "-X main.release=develop" ./cmd/scheduler
if [ $? -ne 0 ]; then
    echo "Ошибка при сборке calendar_scheduler"
    exit 1
fi
go build -o bin/calendar_storer -ldflags "-X main.release=develop" ./cmd/storer
if [ $? -ne 0 ]; then
    echo "Ошибка при сборке calendar_storer"
    exit 1
fi
echo ""

echo "[2/3] Запуск сервисов..."
cd bin && ./calendar --config=../configs/config.yaml &
sleep 2

cd .. && cd bin && ./calendar_scheduler --config=../configs/scheduler_config.yaml &
sleep 2

cd .. && cd bin && ./calendar_storer --config=../configs/storer_config.yaml &
sleep 2

echo ""
echo "========================================"
echo "Система запущена!"
echo "========================================"
echo ""
echo "Запущены следующие сервисы:"
echo "  - Calendar API (http://localhost:8080)"
echo "  - Calendar Scheduler"
echo "  - Calendar Storer"
echo ""
echo "Для остановки сервисов закройте терминалы."
echo "Для остановки PostgreSQL и Kafka выполните:"
echo "  docker-compose -f docker/docker-compose.yml down"
echo ""

wait