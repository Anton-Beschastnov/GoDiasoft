@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Запуск системы Календарь с Kafka
echo ========================================
echo.

cd /d "%~dp0.."

echo [1/5] Остановка старых контейнеров...
docker-compose -f "%~dp0..\docker\docker-compose.yml" down --remove-orphans 2>nul
echo.

echo [2/5] Поднятие Docker контейнеров (PostgreSQL + Kafka)...
docker-compose -f "%~dp0..\docker\docker-compose.yml" up -d
if errorlevel 1 (
    echo Ошибка при запуске docker-compose
    exit /b 1
)
echo.

echo [3/5] Ожидание запуска PostgreSQL...
timeout /t 10 /nobreak >nul
echo.

echo [4/5] Применение миграций базы данных...
go run github.com/pressly/goose/v3/cmd/goose@v3.19.0 migrate -dir "%~dp0..\migrations" postgres "postgres://calendar:calendar@localhost:5432/calendar?sslmode=disable" up
if errorlevel 1 (
    echo Ошибка при применении миграций
    exit /b 1
)
echo.

echo [5/5] Сборка проектов...
go build -o "%~dp0..\bin\calendar.exe" -ldflags "-X main.release=develop" "%~dp0..\cmd\calendar"
if errorlevel 1 (
    echo Ошибка при сборке calendar
    exit /b 1
)
go build -o "%~dp0..\bin\calendar_scheduler.exe" -ldflags "-X main.release=develop" "%~dp0..\cmd\scheduler"
if errorlevel 1 (
    echo Ошибка при сборке calendar_scheduler
    exit /b 1
)
go build -o "%~dp0..\bin\calendar_storer.exe" -ldflags "-X main.release=develop" "%~dp0..\cmd\storer"
if errorlevel 1 (
    echo Ошибка при сборке calendar_storer
    exit /b 1
)
echo.

echo.
echo ========================================
echo Запуск всех сервисов...
echo ========================================
echo.

start "Calendar API" cmd /c "cd /d %~dp0.. && .\bin\calendar.exe --config=.\configs\config.yaml"
timeout /t 2 /nobreak >nul

start "Calendar Scheduler" cmd /c "cd /d %~dp0.. && .\bin\calendar_scheduler.exe --config=.\configs\scheduler_config.yaml"
timeout /t 2 /nobreak >nul

start "Calendar Storer" cmd /c "cd /d %~dp0.. && .\bin\calendar_storer.exe --config=.\configs\storer_config.yaml"
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo Система запущена!
echo ========================================
echo.
echo Запущены следующие сервисы:
echo   - PostgreSQL (localhost:5432)
echo   - Kafka (localhost:9092)
echo   - Kafka UI (http://localhost:8080)
echo   - Calendar API (http://localhost:8080)
echo   - Calendar Scheduler
echo   - Calendar Storer
echo.
echo Для остановки системы выполните:
echo   docker-compose -f "%~dp0..\docker\docker-compose.yml" down
echo.
echo Для просмотра логов откройте окна терминала Calendar API, Calendar Scheduler и Calendar Storer.
echo.
pause