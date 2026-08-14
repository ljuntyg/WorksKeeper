package mock

import "log"

func GetMockProtocol() string {
	log.Println("calling MOCK GetMockProtocol()")
	return "http://"
}

func GetMockHost() string {
	log.Println("calling MOCK GetMockHost()")
	return "localhost"
}

func GetMockPort() string {
	log.Println("calling MOCK GetMockPort()")
	return ":8080"
}
