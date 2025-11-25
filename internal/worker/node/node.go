package node

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"pupload/internal/models"
	"pupload/internal/worker/container"
	"slices"

	"github.com/google/uuid"
)

type NodeService struct {
	ContainerService *container.ContainerService
}

func CreateNodeService(cs *container.ContainerService) NodeService {
	return NodeService{
		ContainerService: cs,
	}
}

func (ns NodeService) GetEnvArray(nodeDef models.NodeDef, node models.Node) ([]string, error) {

	env := make([]string, 0)

	flags, err := ns.GetFlags(nodeDef, node)
	if err != nil {
		return []string{}, err
	}

	// TODO: ensure this can't be escaped
	for key, val := range flags {
		env = append(env, key+"="+val)
	}

	return env, nil
}

func (ns NodeService) GetOutputPaths() ([]string, error) {

	return nil, nil
}

type InputStreamOutput struct {
	reader    io.ReadCloser
	path      string
	base_path string
}

func (ns NodeService) GetInputStreams(inputs map[string]string, base_path string) map[string]InputStreamOutput {
	out := make(map[string]InputStreamOutput, len(inputs))

	for name, url := range inputs {
		resp, err := http.Get(url)

		if err != nil {
			continue
		}

		mimeBytes := make([]byte, 512)
		io.ReadFull(resp.Body, mimeBytes)

		mime := http.DetectContentType(mimeBytes)
		_ = mime

		reader := io.MultiReader(bytes.NewReader(mimeBytes), resp.Body)

		// get Tar reader

		id := uuid.Must(uuid.NewV7())

		size := resp.ContentLength

		pr, pw := io.Pipe()

		go func() {

			defer resp.Body.Close()
			defer pw.Close()
			tw := tar.NewWriter(pw)
			defer tw.Close()

			hdr := &tar.Header{
				Name: id.String(),
				Mode: 0600,
				Size: size,
			}

			if err := tw.WriteHeader(hdr); err != nil {
				pw.CloseWithError(err)
				return
			}

			if _, err := io.Copy(tw, reader); err != nil {
				pw.CloseWithError(err)
				return
			}

		}()

		path := path.Join(base_path, id.String())

		out[name] = InputStreamOutput{reader: pr, path: path, base_path: base_path}
	}

	return out
}

func (ns NodeService) GetFlags(nodeDef models.NodeDef, node models.Node) (map[string]string, error) {
	flagMap := make(map[string]string)

	for _, flagDef := range nodeDef.Flags {

		for _, flagVal := range node.Flags {
			if flagVal.Name == flagDef.Name {
				flagMap[flagVal.Name] = flagVal.Value
				break
			}
		}

		if _, ok := flagMap[flagDef.Name]; !ok && flagDef.Required {
			return flagMap, fmt.Errorf("Flag %s is required", flagDef.Name)
		}

	}

	return flagMap, nil
}

func (ns NodeService) CanWorkerRunContainer(nodeDef models.NodeDef) bool {

	imageList, err := ns.ContainerService.ListImages()
	if err != nil {
		return false
	}

	return slices.Contains(imageList, nodeDef.Image)

}
