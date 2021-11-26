package main

import (
	"github.com/knlambert/unit-tests-demo/internal"
	"io/fs"
	"log"
	"os"
)

type IPGetter interface {
	GetPublicIP() (*string, error)
}

type FileWriter interface {
	Write(filename string, data []byte, perm fs.FileMode) error
}

func Execute(
	ipGetter IPGetter,
	fileWriter FileWriter,
	outputFile string,
) error {
	//Get request on the API.
	publicIP, err := ipGetter.GetPublicIP()

	if err != nil {
		return err
	}

	//Write its content to a file.
	return fileWriter.Write(outputFile, []byte(*publicIP), 0644)
}

func main() {
	if err := Execute(&internal.Ipify{}, &internal.FileRepository{}, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
