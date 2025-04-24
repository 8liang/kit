package excel

var _Config *config

type config struct {
	excelDir    string
	clientDir   string
	serverDir   string
	packageName string
}

type Option func(*config)
