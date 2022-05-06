# oc-client-go

A project for simplifying when working with Open Content API.

## Search example

    import "github.com/navigacontentlab/oc-client-go"

    client, err := oc.New(oc.Options{
		BaseURL: "https://host:8443/opencontent",
		Auth:    oc.BearerAuth("<token>"),
	})

	req := oc.SearchRequest{
		Properties: "uuid,updated",
		Query:      "contenttype:Image",
		Sort: []oc.SearchSort{{
			IndexField: "updated",
			Descending: true,
		}},
	}

	resp, err := client.Search(context.Background(), req)


## Upload example

    import (
        "github.com/navigacontentlab/oc-client-go"
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/clientcredentials"
    )

    config := clientcredentials.Config{
		ClientID:     "<client-id>",
		ClientSecret: "<client-secret>",
		AuthStyle:    oauth2.AuthStyleInParams,
		TokenURL:     "https://access-token.stage.id.navigacloud.com/v1/token",
	}

	httpClient := config.Client(context.Background())
	httpClient.Timeout = time.Second * 5

	client, err := oc.New(oc.Options{
		BaseURL:    "https://<host>:7777/opencontent",
		HTTPClient: httpClient,
	})

	reader, err := os.Open("sample.jpeg")
	metadataReader, err := os.Open("sample-image-metadata.xml")

	req := oc.UploadRequest{
		Files: oc.FileSet{
			"file": oc.File{
				Name:     "sample.jpeg",
				Reader:   reader,
				Mimetype: "image/jpeg",
			},
			"metadata": oc.File{
				Name:     "sample-image.metadata.xml",
				Reader:   metadataReader,
				Mimetype: "application/vnd.iptc.g2.newsitem+xml.picture",
			},
		},
	}

	resp, err := client.Upload(context.Background(), req)


## Metrics

This library provides metrics in the form of a prometheus collector.

The following metrics are supported:

* statusCodes
* responseTimes