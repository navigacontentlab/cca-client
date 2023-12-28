package cca

import (
	context "context"
	"errors"
	"fmt"
	io "io"
	"net/http"
	"time"
)

//nolint:cyclop
func Upload(
	ctx context.Context, files Files, client *http.Client, data io.Reader, size int64,
) (string, *Manifest, error) {
	upload, err := files.CreateUpload(ctx, &CreateUploadReq{})
	if err != nil {
		return "", nil, fmt.Errorf("failed to create upload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, upload.UploadUrl, data)
	if err != nil {
		return upload.UploadId, nil, fmt.Errorf(
			"failed to create upload request: %w", err)
	}

	req.ContentLength = size

	res, err := client.Do(req)
	if err != nil {
		return upload.UploadId, nil, fmt.Errorf(
			"failed to perform upload request: %w", err)
	}

	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return upload.UploadId, nil, fmt.Errorf(
			"server responded with %q", res.Status)
	}

	for {
		status, err := files.GetStatus(ctx, &GetStatusReq{
			UploadId: upload.UploadId,
		})
		if err != nil {
			return upload.UploadId, nil, fmt.Errorf(
				"failed to check upload status: %w", err)
		}

		switch status.Status {
		case ProcessingStatus_ERROR:
			time.Sleep(1 * time.Second)
			// return nil, fmt.Errorf("upload handling failed: %s", status.Message)
		case ProcessingStatus_IN_PROGRESS:
			time.Sleep(1 * time.Second)
		case ProcessingStatus_DONE:
			return upload.UploadId, status.Manifest, nil
		case ProcessingStatus_ASSET_EXISTS:
			return status.Manifest.Uuid, status.Manifest, nil
		case ProcessingStatus_UNKNOWN:
			return upload.UploadId, nil, errors.New(
				"unknown status returned for upload")
		}
	}
}

func (m *Manifest) GetArtifact(artifactType string) *Artifact {
	for i := range m.Artifacts {
		if m.Artifacts[i].Type == artifactType {
			return m.Artifacts[i]
		}
	}

	return nil
}
