package transport

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"code.storageos.net/storageos/c2-cli/pkg/entity"
	"code.storageos.net/storageos/c2-cli/pkg/id"
	"code.storageos.net/storageos/c2-cli/pkg/node"
	"code.storageos.net/storageos/c2-cli/pkg/volume"

	"code.storageos.net/storageos/openapi"
)

type openAPICodec struct{}

func (c openAPICodec) decodeGetNode(model openapi.Node) (*node.Resource, error) {
	node := &node.Resource{
		ID:     id.Node(model.Id),
		Name:   model.Name,
		Health: entity.HealthFromString(string(model.Health)),

		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,

		Labels:  map[string]string(model.Labels),
		Version: entity.VersionFromString(model.Version),
	}

	return node, nil
}

func (c openAPICodec) decodeDescribeNode(model openapi.Node) (*node.Resource, error) {
	n, err := c.decodeGetNode(model)
	if err != nil {
		return nil, err
	}

	n.Configuration = &node.Configuration{
		IOAddr:         model.IoEndpoint,
		SupervisorAddr: model.SupervisorEndpoint,
		GossipAddr:     model.GossipEndpoint,
		ClusteringAddr: model.ClusteringEndpoint,
	}

	return n, nil
}

func (c openAPICodec) decodeVolume(model openapi.Volume) (*volume.Resource, error) {

	// TODO: Validate if fields ok? (complete fields too)
	volume := &volume.Resource{
		ID:   id.Volume(model.Id),
		Name: model.Name,
	}

	return volume, nil
}

type OpenAPI struct {
	mu *sync.RWMutex

	client *openapi.APIClient
	codec  openAPICodec
}

func (o *OpenAPI) Authenticate(ctx context.Context, username, password string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	_, resp, err := o.client.DefaultApi.AuthenticateUser(
		ctx,
		openapi.AuthUserData{
			Username: username,
			Password: password,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	token := strings.TrimPrefix(resp.Header.Get("Authorization"), "Bearer ")
	o.client.GetConfig().AddDefaultHeader("Authorization", token)

	return nil
}

// -----------------------------------------------------------------------------
// 								GET
// -----------------------------------------------------------------------------

func (o *OpenAPI) GetNode(ctx context.Context, uid id.Node) (*node.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	model, _, err := o.client.DefaultApi.GetNode(ctx, uid.String())
	if err != nil {
		// TODO: Maybe do the error mapping at the transport level?
		// → if so change below as well.
		// → Error mapping could use the resp object to be a bit
		// intelligent?
		return nil, err
	}

	n, err := o.codec.decodeGetNode(model)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (o *OpenAPI) GetListNodes(ctx context.Context) ([]*node.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	models, _, err := o.client.DefaultApi.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*node.Resource, len(models))
	for i, m := range models {

		// → If we error here then there's an incompatibility somewhere so
		// aborting is probably a good shout.
		n, err := o.codec.decodeGetNode(m)
		if err != nil {
			return nil, err
		}

		nodes[i] = n
	}

	return nodes, nil
}

func (o *OpenAPI) GetVolume(ctx context.Context, namespace id.Namespace, uid id.Volume) (*volume.Resource, error) {
	model, _, err := o.client.DefaultApi.GetVolume(ctx, namespace.String(), uid.String())

	v, err := o.codec.decodeVolume(model)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// -----------------------------------------------------------------------------
// 								DESCRIBE
// -----------------------------------------------------------------------------

func (o *OpenAPI) DescribeNode(ctx context.Context, uid id.Node) (*node.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	model, _, err := o.client.DefaultApi.GetNode(ctx, uid.String())
	if err != nil {
		// TODO: Maybe do the error mapping at the transport level?
		// → if so change below as well.
		// → Error mapping could use the resp object to be a bit
		// intelligent?
		return nil, err
	}

	n, err := o.codec.decodeDescribeNode(model)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (o *OpenAPI) DescribeListNodes(ctx context.Context) ([]*node.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	models, _, err := o.client.DefaultApi.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodes := make([]*node.Resource, len(models))
	for i, m := range models {

		// → If we error here then there's an incompatibility somewhere so
		// aborting is probably a good shout.
		n, err := o.codec.decodeDescribeNode(m)
		if err != nil {
			return nil, err
		}

		nodes[i] = n
	}

	return nodes, nil
}

func NewOpenAPI(apiEndpoint, userAgent string) *OpenAPI {
	// Init the OpenAPI client
	cfg := &openapi.Configuration{
		BasePath:      "v2",
		DefaultHeader: map[string]string{},
		Host:          apiEndpoint,
		Scheme:        "http",
		UserAgent:     userAgent,
	}

	client := openapi.NewAPIClient(cfg)

	return &OpenAPI{
		mu: &sync.RWMutex{},

		client: client,
		codec:  openAPICodec{},
	}
}
