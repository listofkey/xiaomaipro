package logging

type Config struct {
	Level     string         `json:",default=info,options=[debug,info,warn,error]"`
	Encoding  string         `json:",default=json,options=[json,console]"`
	Console   bool           `json:",default=true"`
	Directory string         `json:",default=logs"`
	Filename  string         `json:",optional"`
	Logstash  LogstashConfig `json:",optional"`
}

type LogstashConfig struct {
	Enabled               bool   `json:",default=false"`
	Network               string `json:",default=tcp"`
	Address               string `json:",optional"`
	ConnectTimeoutSeconds int    `json:",default=3"`
	WriteTimeoutSeconds   int    `json:",default=3"`
}
