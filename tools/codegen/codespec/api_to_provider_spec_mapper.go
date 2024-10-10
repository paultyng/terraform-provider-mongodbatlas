//nolint:gocritic
package codespec

import (
	"errors"
	"fmt"
	"log"
	"strings"

	high "github.com/pb33f/libopenapi/datamodel/high/v3"
	low "github.com/pb33f/libopenapi/datamodel/low/v3"

	"github.com/mongodb/terraform-provider-mongodbatlas/internal/common/conversion"
	"github.com/mongodb/terraform-provider-mongodbatlas/tools/codegen/config"
	"github.com/mongodb/terraform-provider-mongodbatlas/tools/codegen/openapi"
)

func ToProviderSpecModel(atlasAdminAPISpecFilePath, configPath string, resourceName string) (*Model, error) {
	apiSpec, err := openapi.ParseAtlasAdminAPI(atlasAdminAPISpecFilePath)
	if err != nil {
		panic(err)
	}

	genConfig, err := config.ParseGenConfigYAML(configPath)
	if err != nil {
		panic(err)
	}

	// find resource operations, schemas, etc from OAS
	oasResource, err := getOASResource(apiSpec.Model, genConfig.Resources[resourceName], resourceName)
	if err != nil {
		panic(err)
	}

	// map OAS resource model to CodeSpecModel
	codeSpecResource := oasResourceToProviderSpecModel(oasResource, genConfig.Resources[resourceName], resourceName)

	return &Model{Resources: []Resource{*codeSpecResource}}, nil
}

func oasResourceToProviderSpecModel(oasResource APISpecResource, resourceConfig config.Resource, name string) *Resource {
	createOp := oasResource.CreateOp
	readOp := oasResource.ReadOp

	pathParamAttributes := pathParamsToAttributes(createOp, readOp)
	createRequestAttributes := opRequestToAttributes(createOp)
	createResponseAttributes := opResponseToAttributes(createOp)
	readResponseAttributes := opResponseToAttributes(readOp)

	attributes := mergeAttributes(pathParamAttributes, createRequestAttributes, createResponseAttributes, readResponseAttributes)

	schema := &Schema{
		Description:        oasResource.Description,
		DeprecationMessage: oasResource.DeprecationMessage,
		Attributes:         attributes,
	}

	resource := &Resource{
		Name:   name,
		Schema: schema,
	}

	applyConfigSchemaOptions(resourceConfig, resource)

	return resource
}

func pathParamsToAttributes(createOp, readOp *high.Operation) Attributes {
	pathParams := createOp.Parameters
	pathParams = append(pathParams, readOp.Parameters...)

	// TODO: fix path param description not mapping
	pathAttributes := Attributes{}
	for _, param := range pathParams {
		if param.In != OASPathParam && param.In != OASQueryParam {
			continue
		}

		s, err := BuildSchema(param.Schema)
		if err != nil {
			continue
		}

		paramName := param.Name
		parameterAttribute, err := s.buildResourceAttr(paramName, ComputedOptional)
		if err != nil {
			log.Printf("[WARN] Path param %s could not be mapped: %s", paramName, err)
			continue
		}
		pathAttributes = append(pathAttributes, *parameterAttribute)
	}
	return pathAttributes
}

func opRequestToAttributes(op *high.Operation) Attributes {
	var requestAttributes Attributes
	requestSchema, err := buildSchemaFromRequest(op)
	if err != nil {
		log.Printf("[WARN] Request schema could not be mapped (OperationId: %s): %s", op.OperationId, err)
		return nil
	}

	requestAttributes, err = buildResourceAttrs(requestSchema)
	if err != nil {
		log.Printf("[WARN] Request attributes could not be mapped (OperationId: %s): %s", op.OperationId, err)
		return nil
	}

	return requestAttributes
}

func opResponseToAttributes(op *high.Operation) Attributes {
	var responseAttributes Attributes
	responseSchema, err := buildSchemaFromResponse(op)
	if err != nil {
		if errors.Is(err, errSchemaNotFound) {
			log.Printf("[INFO] Operation response body schema not found (OperationId: %s)", op.OperationId)
		} else {
			log.Printf("[WARN] Operation response body schema could not be mapped (OperationId: %s): %s", op.OperationId, err)
		}
	} else {
		responseAttributes, err = buildResourceAttrs(responseSchema)
		if err != nil {
			log.Printf("[WARN] Operation response body schema could not be mapped (OperationId: %s): %s", op.OperationId, err)
		}
	}
	return responseAttributes
}

func getOASResource(spec high.Document, resourceConfig config.Resource, name string) (APISpecResource, error) {
	var errResult error
	var resourceDeprecationMsg *string

	createOp, err := extractOp(spec.Paths, resourceConfig.Create)
	if err != nil {
		errResult = errors.Join(errResult, fmt.Errorf("unable to extract '%s.create' operation: %w", name, err))
	}
	readOp, err := extractOp(spec.Paths, resourceConfig.Read)
	if err != nil {
		errResult = errors.Join(errResult, fmt.Errorf("unable to extract '%s.read' operation: %w", name, err))
	}
	updateOp, err := extractOp(spec.Paths, resourceConfig.Update)
	if err != nil {
		errResult = errors.Join(errResult, fmt.Errorf("unable to extract '%s.update' operation: %w", name, err))
	}
	deleteOp, err := extractOp(spec.Paths, resourceConfig.Delete)
	if err != nil {
		errResult = errors.Join(errResult, fmt.Errorf("unable to extract '%s.delete' operation: %w", name, err))
	}

	commonParameters, err := extractCommonParameters(spec.Paths, resourceConfig.Read.Path)
	if err != nil {
		errResult = errors.Join(errResult, fmt.Errorf("unable to extract '%s' common parameters: %w", name, err))
	}

	if readOp.Deprecated != nil && *readOp.Deprecated {
		resourceDeprecationMsg = conversion.StringPtr(DefaultDeprecationMsg)
	}

	return APISpecResource{
		Description:        &createOp.Description,
		DeprecationMessage: resourceDeprecationMsg,
		CreateOp:           createOp,
		ReadOp:             readOp,
		UpdateOp:           updateOp,
		DeleteOp:           deleteOp,
		CommonParameters:   commonParameters,
	}, errResult
}

func extractOp(paths *high.Paths, apiOp *config.APIOperation) (*high.Operation, error) {
	if apiOp == nil {
		return nil, nil
	}

	if paths == nil || paths.PathItems == nil || paths.PathItems.GetOrZero(apiOp.Path) == nil {
		return nil, fmt.Errorf("path '%s' not found in OpenAPI spec", apiOp.Path)
	}

	pathItem, _ := paths.PathItems.Get(apiOp.Path)

	return extractOpFromPathItem(pathItem, apiOp)
}

func extractOpFromPathItem(pathItem *high.PathItem, apiOp *config.APIOperation) (*high.Operation, error) {
	if pathItem == nil || apiOp == nil {
		return nil, fmt.Errorf("pathItem or apiOp cannot be nil")
	}

	switch strings.ToLower(apiOp.Method) {
	case low.PostLabel:
		return pathItem.Post, nil
	case low.GetLabel:
		return pathItem.Get, nil
	case low.PutLabel:
		return pathItem.Put, nil
	case low.DeleteLabel:
		return pathItem.Delete, nil
	case low.PatchLabel:
		return pathItem.Patch, nil
	default:
		return nil, fmt.Errorf("method '%s' not found at OpenAPI path '%s'", apiOp.Method, apiOp.Path)
	}
}

func extractCommonParameters(paths *high.Paths, path string) ([]*high.Parameter, error) {
	if paths.PathItems.GetOrZero(path) == nil {
		return nil, fmt.Errorf("path '%s' not found in OpenAPI spec", path)
	}

	pathItem, _ := paths.PathItems.Get(path)

	return pathItem.Parameters, nil
}

func applyConfigSchemaOptions(resourceConfig config.Resource, resource *Resource) {
	// TODO: implement in follow-up PR
}
