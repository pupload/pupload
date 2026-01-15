package stores

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pupload/pupload/internal/models"
	locals3 "github.com/pupload/pupload/internal/stores/local_s3"
	"github.com/pupload/pupload/internal/stores/s3"
	s3fs "github.com/pupload/pupload/internal/stores/s3_fs"
)

func UnmarshalStore(input models.StoreInput) (models.Store, error) {

	switch input.Type {

	case "s3":
		var params s3.S3StoreInput
		if err := decodeParams(input.Params, &params); err != nil {
			return nil, fmt.Errorf("error decoding params for s3local, %w", err)
		}

		store, err := s3.NewS3Store(params)
		if err != nil {
			return nil, fmt.Errorf("Unable to create s3 store: %w", err)
		}

		return store, nil

	case "s3local":
		var params locals3.LocalS3StoreInput
		if err := decodeParams(input.Params, &params); err != nil {
			return nil, fmt.Errorf("decode params for s3local: %w", err)
		}

		store, err := locals3.NewLocalS3Store(params)
		if err != nil {
			return nil, fmt.Errorf("Unable to create locals3 store: %w", err)
		}
		return store, nil

	case "s3fs":
		var params s3fs.FilesystemS3StoreInput
		if err := decodeParams(input.Params, &params); err != nil {
			return nil, fmt.Errorf("decode params for s3fs store: %w", err)
		}

		store, err := s3fs.NewLocalS3Store(params)
		if err != nil {
			return nil, fmt.Errorf("Unable to create s3fs store: %w", err)
		}

		return store, nil

	default:
		return nil, nil

	}

}

func decodeParams(raw json.RawMessage, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}

	if dec.More() {
		return fmt.Errorf("Unexpected extra data in params")
	}

	return nil
}
