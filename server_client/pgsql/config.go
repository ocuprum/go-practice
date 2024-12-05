package pgsql

type Config struct {
	Host     string
	User     string
	Password string
	DBname   string
	Port     uint16
	SSLmode  string
	Timezone string
}