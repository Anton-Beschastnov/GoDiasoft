#!/bin/sh

# Заменяем переменные окружения в конфигурационном файле
envsubst < /etc/calendar/config.yaml.template > /etc/calendar/config.yaml

# Запускаем приложение
exec "$@"
