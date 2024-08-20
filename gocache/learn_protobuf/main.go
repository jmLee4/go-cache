package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"learn_protobuf/studentpb"
	"log"
)

func main() {
	test := &studentpb.Student{
		Name:   "geektutu",
		Male:   true,
		Scores: []int32{98, 85, 88},
	}
	data, err := proto.Marshal(test)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	fmt.Println("Marshal:", data)

	newTest := &studentpb.Student{}
	err = proto.Unmarshal(data, newTest)
	if err != nil {
		log.Fatal("unmarshaling error:", err)
	}
	fmt.Printf("Unmarshal: %+v\n", newTest)

	if test.GetName() != newTest.GetName() {
		log.Fatalf("data mismatch %q != %q", test.GetName(), newTest.GetName())
	}
}
