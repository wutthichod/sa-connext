package config

import "github.com/wutthichod/sa-connext/shared/utils"

type Config struct {
    Addr      string
    MongoURI  string
    RabbitURI string
    Email     string
    EmailPW   string
}

func LoadConfig() Config {
    return Config{
        Addr:      utils.GetEnvString("ADDR", "9093"),
        MongoURI:  utils.GetEnvString("MONGODB_URI", "mongodb://localhost:27017"),
        RabbitURI: utils.GetEnvString("RABBITMQ_URI", ""),
        Email:     utils.GetEnvString("EMAIL", ""),
        EmailPW:   utils.GetEnvString("EMAIL_PW", ""),
    }
}