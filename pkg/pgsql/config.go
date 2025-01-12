package pgsql

type Config struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     uint16
	SSLMode  string
	Timezone string
}