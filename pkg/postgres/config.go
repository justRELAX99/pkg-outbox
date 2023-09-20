package postgres

const (
	driverName = "postgres"
)

type DataBases struct {
	EnkodPG Config `json:"enkodPG"`
}

type Config struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	DBName       string `json:"dbName"`
	Password     string `json:"password"`
	MaxOpenConns int    `json:"maxOpenConns"`
	MaxAttempts  int    `json:"maxAttempts"`
	MaxDelay     int    `json:"maxDelay"`
}
