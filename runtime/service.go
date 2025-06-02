package runtime

import "fmt"

var (
	serviceFormat = `{ "type":"call_service", "domain":"%s", "service":"%s", "service_data": {%s}, "target":{"entity_id":"%s"},`
	idstr         = `"id":%d }`
)

type ServiceData struct {
	Key   string
	Value string
}

func ServiceCmd(domain string, service string, entityID string, sd ...ServiceData) string {
	var serviceData string
	for _, s := range sd {
		serviceData += fmt.Sprintf(`"%s":%s,`, s.Key, s.Value)
	}
	if len(serviceData) > 0 {
		serviceData = serviceData[:len(serviceData)-1]
	}
	cmd := fmt.Sprintf(serviceFormat, domain, service, serviceData, entityID) + idstr
	return cmd
}

func LightCmd(entityID string, serviceData ...ServiceData) string {
	return ServiceCmd("light", "turn_on", entityID, serviceData...)
}

func LightCmdOff(entityID string, serviceData ...ServiceData) string {
	return ServiceCmd("light", "turn_off", entityID, serviceData...)
}

func NumberCmd(entityID string, serviceData ...ServiceData) string {
	return ServiceCmd("number", "set_value", entityID, serviceData...)
}
