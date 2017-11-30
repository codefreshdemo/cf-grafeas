package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-resty/resty"
	"github.com/grafeas/client-go/v1alpha1"
	"github.com/grafeas/grafeas/samples/server/go-server/api/server/name"
)

type cfImageDetails struct {
	dockerFile string
	//tags       []string
}

func main() {
	grafeas_host := os.Args[1]
	image_name := os.Args[2]
	grafeas_client := v1alpha1.NewGrafeasApiWithBasePath(grafeas_host)
	nPID := os.Getenv("CF_REPO_OWNER")
	nID := os.Getenv("CF_REPO_NAME")

	cfImage := getCfImageData(image_name)
	log.Printf("%v", cfImage)
	n := BuildDetailsNote(nPID, nID)
	if got, _, err := grafeas_client.GetNote(nPID, nID); err != nil {
		log.Printf("Pipeline details not found. Creating a new one")
		createdN, _, err := grafeas_client.CreateNote(nPID, nID, *n)
		if err != nil {
			log.Fatalf("Error creating note for pipeline: %v", err)
		} else {
			log.Printf("Succesfully created note for pipeline: %v", createdN)
		}
	} else {
		log.Printf("Note already exists: %v", got)
	}

	oPID := nPID + "-builds"
	o := BuildOccurrence(n.Name)
	createdO, _, err := grafeas_client.CreateOccurrence(oPID, *o)
	if err != nil {
		log.Fatalf("Error creating occurrence %v", err)
	} else {
		log.Printf("Succesfully created occurrence: %v", createdO)
	}

	_, oID, pErr := name.ParseOccurrence(createdO.Name)
	if pErr != nil {
		log.Fatalf("Unable to get occurenceId from occurrence name %v: %v", createdO.Name, err)
	}
	if got, _, err := grafeas_client.GetOccurrence(oPID, oID); err != nil {
		log.Printf("Error getting occurrence %v", err)
	} else {
		log.Printf("Succesfully got occurrence: %v", got)
	}

}

func getCfImageData(image_name string) *cfImageDetails {
	cf_img_api_prefix := "https://g.codefresh.io/api/images/"
	res, err := resty.R().
		SetQueryString("imageDisplayName="+image_name).
		SetHeader("Accept", "application/json").
		SetHeader("x-access-token", os.Getenv("CF_TOKEN")).
		Get(cf_img_api_prefix)
	if err != nil {
		panic(err)
	}
	//log.Printf("Image data : %v", res.String())
	var image_data []map[string]interface{}
	err = json.Unmarshal([]byte(res.String()), &image_data)
	if err != nil {
		panic(err)
	}
	return &cfImageDetails{
		dockerFile: image_data[0]["dockerFile"].(string),
		//	tags:       image_data[0]["tags"].([]string),
	}
}
func BuildDetailsNote(pID, nID string) *v1alpha1.Note {
	return &v1alpha1.Note{
		Name:             fmt.Sprintf("projects/%v/notes/%v", pID, nID),
		ShortDescription: nID,
		LongDescription:  fmt.Sprintf("Codefresh build %v for %v", nID, pID),
		Kind:             "BUILD_DETAILS",
		BuildType: v1alpha1.BuildType{
			BuilderVersion: pID,
			Signature: v1alpha1.BuildSignature{
				KeyId: "MY_COOL_KEY",
			},
		},
		RelatedUrl: []v1alpha1.RelatedUrl{
			{
				Url:   os.Getenv("CF_BUILD_URL"),
				Label: "Codefresh Build URL",
			},
			{
				Url:   os.Getenv("CF_COMMIT_URL"),
				Label: "Commit url",
			},
		},
	}
}

func BuildOccurrence(noteName string) *v1alpha1.Occurrence {

	p := v1alpha1.BuildProvenance{
		Id:        os.Getenv("CF_BUILD_ID"),
		ProjectId: os.Getenv("CF_REPO_NAME"),
		BuiltArtifacts: []v1alpha1.Artifact{
			{
				Id: os.Getenv("CF_BUILT_IMAGE"),
			},
		},
		BuilderVersion: noteName,
		SourceProvenance: v1alpha1.Source{
			SourceContext: v1alpha1.ExtendedSourceContext{
				Context: v1alpha1.SourceContext{
					Git: v1alpha1.GitSourceContext{
						Url:        os.Getenv("CF_COMMIT_URL"),
						RevisionId: os.Getenv("CF_REVISION"),
					},
				},
			},
		},
	}
	pb, _ := json.Marshal(p)

	return &v1alpha1.Occurrence{
		ResourceUrl: os.Getenv("CF_BUILD_URL"),
		NoteName:    noteName,
		Kind:        "BUILD_DETAILS",
		BuildDetails: v1alpha1.BuildDetails{
			Provenance:      p,
			ProvenanceBytes: string(pb),
		},
	}
}
