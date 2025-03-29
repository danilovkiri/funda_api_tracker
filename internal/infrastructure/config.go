// Package infrastructure реализует функционал инициализации и настройки инфраструктурных сервисов приложения.
package infrastructure

import (
	"fundaNotifier/internal/infrastructure/repository/mysql"
)

type Config struct {
	MySQLConfig mysql.Config
}
