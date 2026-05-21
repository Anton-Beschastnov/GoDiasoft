@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo Запуск системы Календарь с Kafka
echo ========================================
echo.

cd /d "%~dp0.."

echo [1/3] Сборка проектов...
go build -o bin\calendar.exe -ldflags "-X main.release=develop" ./cmd/calendar
if errorlevel 1 (
    echo Ошибка при сборке calendar
    exit /b 1
)
go build -o bin\calendar_scheduler.exe -ldflags "-X main.release=develop" ./cmd/scheduler
if errorlevel 1 (
    echo Ошибка при сборке calendar_scheduler
    exit /b 1
)
go build -o bin\calendar_storer.exe -ldflags "-X main.release=develop" ./cmd/storer
if errorlevel 1 (
    echo Ошибка при сборке calendar_storer
    exit /b 1
)
echo.

echo [2/3] Запуск сервисов...
start "Calendar API" cmd /c "cd /d %~dp0.. && .\bin\calendar.exe --config=.\configs\config.yaml & pause"
timeout /t 2 /nobreak >nul

start "Calendar Scheduler" cmd /c "cd /d %~dp0.. && .\bin\calendar_scheduler.exe --config=.\configs\scheduler_config.yaml & pause"
timeout /t 2 /nobreak >nul

start "Calendar Storer" cmd /c "cd /d %~dp0.. && .\bin\calendar_storer.exe --config=.\configs\storer_config.yaml & pause"
timeout /t 2 /nobreak >nul

echo.
echo ========================================
echo Система запущена!
echo ========================================
echo.
echo Запущены следующие сервисы:
echo   - Calendar API (http://localhost:8080)
echo   - Calendar Scheduler
echo   - Calendar Storer
echo.
echo Для остановки сервисов закройте окна терминалов.
echo Для остановки PostgreSQL и Kafka выполните:
echo   docker-compose -f docker\docker-compose.yml down
echo.
pause