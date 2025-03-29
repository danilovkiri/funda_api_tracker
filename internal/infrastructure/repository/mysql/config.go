package mysql

type Config struct { //nolint:golint
	DNS string `env:"MYSQL_DNS" env-default:"./example.db"`
}
