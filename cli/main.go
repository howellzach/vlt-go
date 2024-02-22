package main

import (
	"fmt"
	"log"
	"os"

	"github.com/howellzach/vlt-go"
)

func main() {

	client, err := vlt.NewClient(
		os.Getenv("HCP_ORGANIZATION_ID"), // null
		os.Getenv("HCP_PROJECT_ID"),      // null
		os.Getenv("HCP_APPLICATION_NAME"),
		os.Getenv("HCP_CLIENT_ID"),
		os.Getenv("HCP_CLIENT_SECRET"),
		os.Getenv("HCP_PROJECT_NAME"), // null
	)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Organization ID: ", client.OrganizationID)
	fmt.Println("Project ID: ", client.ProjectID)
	fmt.Println("Project Name: ", client.ProjectName)

	secrets, err := client.GetAllSecrets()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Secrets: ", secrets)

}
