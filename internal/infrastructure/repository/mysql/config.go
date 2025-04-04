package mysql

type Config struct { //nolint:golint
	DNS string `env:"SQLITE_DNS" env-default:"./example.db?_loc=auto"`
}
