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
