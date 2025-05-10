package configs

var Config = struct {
	AIPredictURL          string
	EtcdAddress           string
	RedisAddress          string
	PostgresConnectString string
	WebApiPort            string
}{
	AIPredictURL:          "http://localhost:5000/predict",
	EtcdAddress:           "localhost:2379",
	RedisAddress:          "localhost:6379",
	PostgresConnectString: "host=localhost port=5432 user=postgres password=postgres dbname=tasks sslmode=disable",
	WebApiPort:            ":8080",
}
