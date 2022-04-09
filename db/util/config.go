package util

import "github.com/spf13/viper"

type Config struct {
	DBDriver     string `mapstructure:"DB_DRIVER"`
	DBSource     string `mapstructure:"DB_SOURCE"`
	ServerAdress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)  //指定配置文件所在的路径
	viper.SetConfigName("app") //配置文件的名称
	viper.SetConfigType("env") //配置文件的类型,可以是yaml,json等,这里是env类型

	viper.AutomaticEnv() //如果有环境变量,则自动覆盖配置文件中的值

	err = viper.ReadInConfig() //读取变量
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config) //将获取的配置信息写入结构体
	return
}
