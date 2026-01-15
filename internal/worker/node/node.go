package node

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"slices"

	mimetypes "github.com/pupload/pupload/internal/mimetype"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/worker/container"

	"github.com/google/uuid"
)

type NodeService struct {
	SyncLayer syncplane.SyncLayer
	CS        *container.ContainerService
}

func CreateNodeService(cs *container.ContainerService, s syncplane.SyncLayer) NodeService {
	return NodeService{
		CS:        cs,
		SyncLayer: s,
	}
}

func (ns NodeService) AddEnvFlagMap(m map[string]string, nodeDef models.NodeDef, node models.Node) error {

	flags, err := ns.GetFlags(nodeDef, node)
	if err != nil {
		return err
	}

	// TODO: ensure this can't be escaped
	for key, val := range flags {
		m[key] = val
	}

	return nil
}

func (ns NodeService) GetFlagEnvArray(nodeDef models.NodeDef, node models.Node) ([]string, error) {

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

type preparedIO struct {
	name      string
	base_path string
	path      string
	filename  string
	url       string
}

func (ns *NodeService) prepareIO(inputs, outputs map[string]string, nodeDef models.NodeDef, basePath string) ([]preparedIO, []preparedIO, error) {
	in := make([]preparedIO, 0, len(inputs))
	out := make([]preparedIO, 0, len(outputs))

	for _, inputDef := range nodeDef.Inputs {
		inputURL, ok := inputs[inputDef.Name]
		if !ok {
			switch inputDef.Required {
			case true:
				return nil, nil, fmt.Errorf("PrepareInputs: node missing required input %s", inputDef.Name)
			case false:
				continue
			}
		}

		typeSet, err := mimetypes.CreateMimeSet(inputDef.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("PrepareInputs: error creating mimeset: %w", err)
		}

		ext, err := ns.ValidateInput(inputURL, *typeSet)
		if err != nil {
			return nil, nil, fmt.Errorf("PrepareInputs: error validating inputs: %w", err)
		}

		path, filename := ns.GetPath(basePath, ext)

		in = append(in, preparedIO{
			name:      inputDef.Name,
			url:       inputURL,
			base_path: basePath,
			path:      path,
			filename:  filename,
		})

	}

	for _, outputDef := range nodeDef.Outputs {
		outputURL, ok := outputs[outputDef.Name]
		if !ok {
			return nil, nil, fmt.Errorf("no output URL for output %s", outputDef.Name)
		}

		extension := ns.getOutputExtension(outputDef.Type)
		path, filename := ns.GetPath(basePath, extension)

		out = append(out, preparedIO{
			url:       outputURL,
			name:      outputDef.Name,
			base_path: basePath,
			path:      path,
			filename:  filename,
		})
	}

	return in, out, nil
}

func (ns NodeService) AddIOToEnvMap(env map[string]string, prepped []preparedIO) {
	for _, prep := range prepped {
		env[prep.name] = prep.path
	}
}

func (ns NodeService) getOutputExtension(types []models.MimeType) string {
	if len(types) != 1 {
		return ""
	}

	t := types[0]
	return mimetypes.GetExtensionFromMime(t)
}

func (ns NodeService) GetOutputPaths(outputs map[string]string, base_path string) (map[string]string, error) {

	out := make(map[string]string, len(outputs))

	for name := range outputs {
		id := uuid.Must(uuid.NewV7()).String()
		path := path.Join(base_path, id)

		out[name] = path
	}

	return out, nil
}

// Validates a given uploaded file against the qualified allowed mime types.
// Returns the appoprriate file extension
func (ns NodeService) ValidateInput(url string, mimeSet mimetypes.MimeSet) (ext string, err error) {

	resp, err := http.Get(url)

	if err != nil {
		return "", fmt.Errorf("error getting content from %s", url)
	}

	defer resp.Body.Close()
	mimeBytes := make([]byte, 512)

	io.ReadFull(resp.Body, mimeBytes)
	mime := http.DetectContentType(mimeBytes)

	if !mimeSet.Contains(models.MimeType(mime)) {
		return "", fmt.Errorf("invalid content type uploaded")
	}

	ext = mimetypes.GetExtensionFromMime(models.MimeType(mime))
	return ext, nil
}

func (ns NodeService) GetPath(base_path string, extension string) (path string, filename string) {
	filename = uuid.Must(uuid.NewV7()).String() + extension
	return filepath.Join(base_path, filename), filename
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

	imageList, err := ns.CS.ListImages()
	if err != nil {
		return false
	}

	return slices.Contains(imageList, nodeDef.Image)

}
