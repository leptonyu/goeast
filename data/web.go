package data

type Data struct {
	BaseUrl string
	Config  map[string]string
	Map     map[string]interface{}
}

func NewData(conf map[string]string) Data {
	d := Data{BaseUrl: conf["url"], Config: conf, Map: make(map[string]interface{})}
	return d
}

func (d *Data) Url() string {
	return `<a hrep='` + d.BaseUrl + `'>GoEast Language Centers</a>`
}
