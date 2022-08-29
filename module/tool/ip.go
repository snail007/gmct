package tool

import "fmt"

func (s *Tool) ip() {
	for _, v := range getLocalIP() {
		fmt.Println(v)
	}
}

