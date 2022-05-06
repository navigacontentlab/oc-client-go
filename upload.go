package oc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

type FileSet map[string]File

type File struct {
	Name     string
	Reader   io.Reader
	Mimetype string
}

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

type mimeFields map[string]string

func (fields mimeFields) write(writer *multipart.Writer) error {
	for field, value := range fields {
		part, err := writer.CreateFormField(field)
		if err != nil {
			return fmt.Errorf("failed to create form field %q: %w", field, err)
		}

		_, err = part.Write([]byte(value))
		if err != nil {
			return fmt.Errorf("failed to write to form field %q: %w", field, err)
		}
	}

	return nil
}

type UploadRequest struct {
	UUID    string
	Source  string
	Files   FileSet
	Unit    string
	Batch   bool
	IfMatch string
}

type UploadResponse struct {
	UUID    string
	ETag    string
	Version int64
}

// Upload saves the fileset in the OC database.
func (c *Client) Upload(ctx context.Context, req UploadRequest) (*UploadResponse, error) { //nolint:gocognit
	start := time.Now()
	url := c.url("objectupload", nil)
	errChan := make(chan error)
	done := make(chan bool)
	pipeOut, pipeIn := io.Pipe()
	writer := multipart.NewWriter(pipeIn)

	var res UploadResponse

	go func() {
		defer writer.Close()
		defer pipeIn.Close()

		fields := mimeFields{
			"source": req.Source,
			"batch":  strconv.FormatBool(req.Batch),
		}

		if req.UUID != "" {
			fields["id"] = req.UUID
		}

		for field, file := range req.Files {
			fields[field] = file.Name
			fields[field+"-mimetype"] = file.Mimetype
		}

		err := fields.write(writer)
		if err != nil {
			errChan <- err
			return
		}

		for _, file := range req.Files {
			if file.Reader == nil {
				continue
			}

			header := make(textproto.MIMEHeader)
			header.Set(
				"Content-Disposition",
				fmt.Sprintf(
					`form-data; name="%s"; filename="%s"`,
					escapeQuotes(file.Name),
					escapeQuotes(file.Name),
				))
			header.Set("Content-Type", file.Mimetype)

			part, err := writer.CreatePart(header)
			if err != nil {
				errChan <- err
				return
			}

			_, err = io.Copy(part, file.Reader)
			if err != nil {
				errChan <- err
				return
			}
		}

		writer.Close()
	}()

	go func() {
		r, err := http.NewRequest("POST", url, pipeOut)
		if err != nil {
			errChan <- err
			return
		}

		r = r.WithContext(ctx)

		if req.Unit != "" {
			r.Header.Set("X-Imid-Unit", req.Unit)
		}

		r.Header.Set("Content-Type", writer.FormDataContentType())

		if req.IfMatch != "" {
			r.Header.Set("If-Match", req.IfMatch)
		}

		if c.auth != nil {
			c.auth(r)
		}

		resp, err := c.httpClient.Do(r)

		if err != nil {
			errChan <- err
			return
		}

		defer safeClose(c.logger, "upload response", resp.Body)

		if c.metrics != nil {
			c.metrics.incStatusCode(ctx, "objectupload", resp.StatusCode)
		}

		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			res.ETag = resp.Header.Get("Etag")

			uuidBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.logger.Logf("failed to read response UUID for upload: %w", err)
			}

			v, err := objectVersionFromHeader(resp.Header, versionOptional)
			if err != nil {
				c.logger.Logf("failed to parse existing version: %w", err)
			}

			res.Version = v

			res.UUID = string(bytes.TrimSpace(uuidBytes))

			done <- true
		} else {
			errChan <- newResponseError(resp)
		}
	}()

	select {
	case err := <-errChan:
		if c.metrics != nil {
			duration := time.Since(start)
			c.metrics.addDuration(ctx, "objectupload", float64(duration.Milliseconds()))
		}

		return nil, err
	case <-done:
		if c.metrics != nil {
			duration := time.Since(start)
			c.metrics.addDuration(ctx, "objectupload", float64(duration.Milliseconds()))
		}

		return &res, nil
	}
}
