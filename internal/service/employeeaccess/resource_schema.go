// Code generated by terraform-plugin-framework-generator DO NOT EDIT.

package employeeaccess

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func EmployeeAccessResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Human-readable label that identifies this cluster.",
				MarkdownDescription: "Human-readable label that identifies this cluster.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexp.MustCompile("^([a-zA-Z0-9][a-zA-Z0-9-]*)?[a-zA-Z0-9]+$"), ""),
				},
			},
			"expiration_time": schema.StringAttribute{
				Required:            true,
				Description:         "Expiration date for the employee access grant.",
				MarkdownDescription: "Expiration date for the employee access grant.",
			},
			"grant_type": schema.StringAttribute{
				Required:            true,
				Description:         "Level of access to grant to MongoDB Employees.",
				MarkdownDescription: "Level of access to grant to MongoDB Employees.",
			},
			"group_id": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "Unique 24-hexadecimal digit string that identifies your project. Use the [/groups](#tag/Projects/operation/listProjects) endpoint to retrieve all projects to which the authenticated user has access.\n\n**NOTE**: Groups and projects are synonymous terms. Your group id is the same as your project id. For existing groups, your group/project id remains the same. The resource and corresponding endpoints use the term groups.",
				MarkdownDescription: "Unique 24-hexadecimal digit string that identifies your project. Use the [/groups](#tag/Projects/operation/listProjects) endpoint to retrieve all projects to which the authenticated user has access.\n\n**NOTE**: Groups and projects are synonymous terms. Your group id is the same as your project id. For existing groups, your group/project id remains the same. The resource and corresponding endpoints use the term groups.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(24, 24),
					stringvalidator.RegexMatches(regexp.MustCompile("^([a-f0-9]{24})$"), ""),
				},
			},
			"links": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"href": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "Uniform Resource Locator (URL) that points another API resource to which this response has some relationship. This URL often begins with `https://cloud.mongodb.com/api/atlas`.",
							MarkdownDescription: "Uniform Resource Locator (URL) that points another API resource to which this response has some relationship. This URL often begins with `https://cloud.mongodb.com/api/atlas`.",
						},
						"rel": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Description:         "Uniform Resource Locator (URL) that defines the semantic relationship between this resource and another API resource. This URL often begins with `https://cloud.mongodb.com/api/atlas`.",
							MarkdownDescription: "Uniform Resource Locator (URL) that defines the semantic relationship between this resource and another API resource. This URL often begins with `https://cloud.mongodb.com/api/atlas`.",
						},
					},
					CustomType: LinksType{
						ObjectType: types.ObjectType{
							AttrTypes: LinksValue{}.AttributeTypes(ctx),
						},
					},
				},
				Optional:            true,
				Computed:            true,
				Description:         "List of one or more Uniform Resource Locators (URLs) that point to API sub-resources, related API resources, or both. RFC 5988 outlines these relationships.",
				MarkdownDescription: "List of one or more Uniform Resource Locators (URLs) that point to API sub-resources, related API resources, or both. RFC 5988 outlines these relationships.",
			},
		},
	}
}

type EmployeeAccessModel struct {
	ClusterName    types.String `tfsdk:"cluster_name"`
	ExpirationTime types.String `tfsdk:"expiration_time"`
	GrantType      types.String `tfsdk:"grant_type"`
	GroupId        types.String `tfsdk:"group_id"`
	Links          types.List   `tfsdk:"links"`
}

var _ basetypes.ObjectTypable = LinksType{}

type LinksType struct {
	basetypes.ObjectType
}

func (t LinksType) Equal(o attr.Type) bool {
	other, ok := o.(LinksType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t LinksType) String() string {
	return "LinksType"
}

func (t LinksType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributes := in.Attributes()

	hrefAttribute, ok := attributes["href"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`href is missing from object`)

		return nil, diags
	}

	hrefVal, ok := hrefAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`href expected to be basetypes.StringValue, was: %T`, hrefAttribute))
	}

	relAttribute, ok := attributes["rel"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`rel is missing from object`)

		return nil, diags
	}

	relVal, ok := relAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`rel expected to be basetypes.StringValue, was: %T`, relAttribute))
	}

	if diags.HasError() {
		return nil, diags
	}

	return LinksValue{
		Href:  hrefVal,
		Rel:   relVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewLinksValueNull() LinksValue {
	return LinksValue{
		state: attr.ValueStateNull,
	}
}

func NewLinksValueUnknown() LinksValue {
	return LinksValue{
		state: attr.ValueStateUnknown,
	}
}

func NewLinksValue(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) (LinksValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/521
	ctx := context.Background()

	for name, attributeType := range attributeTypes {
		attribute, ok := attributes[name]

		if !ok {
			diags.AddError(
				"Missing LinksValue Attribute Value",
				"While creating a LinksValue value, a missing attribute value was detected. "+
					"A LinksValue must contain values for all attributes, even if null or unknown. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("LinksValue Attribute Name (%s) Expected Type: %s", name, attributeType.String()),
			)

			continue
		}

		if !attributeType.Equal(attribute.Type(ctx)) {
			diags.AddError(
				"Invalid LinksValue Attribute Type",
				"While creating a LinksValue value, an invalid attribute value was detected. "+
					"A LinksValue must use a matching attribute type for the value. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("LinksValue Attribute Name (%s) Expected Type: %s\n", name, attributeType.String())+
					fmt.Sprintf("LinksValue Attribute Name (%s) Given Type: %s", name, attribute.Type(ctx)),
			)
		}
	}

	for name := range attributes {
		_, ok := attributeTypes[name]

		if !ok {
			diags.AddError(
				"Extra LinksValue Attribute Value",
				"While creating a LinksValue value, an extra attribute value was detected. "+
					"A LinksValue must not contain values beyond the expected attribute types. "+
					"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Extra LinksValue Attribute Name: %s", name),
			)
		}
	}

	if diags.HasError() {
		return NewLinksValueUnknown(), diags
	}

	hrefAttribute, ok := attributes["href"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`href is missing from object`)

		return NewLinksValueUnknown(), diags
	}

	hrefVal, ok := hrefAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`href expected to be basetypes.StringValue, was: %T`, hrefAttribute))
	}

	relAttribute, ok := attributes["rel"]

	if !ok {
		diags.AddError(
			"Attribute Missing",
			`rel is missing from object`)

		return NewLinksValueUnknown(), diags
	}

	relVal, ok := relAttribute.(basetypes.StringValue)

	if !ok {
		diags.AddError(
			"Attribute Wrong Type",
			fmt.Sprintf(`rel expected to be basetypes.StringValue, was: %T`, relAttribute))
	}

	if diags.HasError() {
		return NewLinksValueUnknown(), diags
	}

	return LinksValue{
		Href:  hrefVal,
		Rel:   relVal,
		state: attr.ValueStateKnown,
	}, diags
}

func NewLinksValueMust(attributeTypes map[string]attr.Type, attributes map[string]attr.Value) LinksValue {
	object, diags := NewLinksValue(attributeTypes, attributes)

	if diags.HasError() {
		// This could potentially be added to the diag package.
		diagsStrings := make([]string, 0, len(diags))

		for _, diagnostic := range diags {
			diagsStrings = append(diagsStrings, fmt.Sprintf(
				"%s | %s | %s",
				diagnostic.Severity(),
				diagnostic.Summary(),
				diagnostic.Detail()))
		}

		panic("NewLinksValueMust received error(s): " + strings.Join(diagsStrings, "\n"))
	}

	return object
}

func (t LinksType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.Type() == nil {
		return NewLinksValueNull(), nil
	}

	if !in.Type().Equal(t.TerraformType(ctx)) {
		return nil, fmt.Errorf("expected %s, got %s", t.TerraformType(ctx), in.Type())
	}

	if !in.IsKnown() {
		return NewLinksValueUnknown(), nil
	}

	if in.IsNull() {
		return NewLinksValueNull(), nil
	}

	attributes := map[string]attr.Value{}

	val := map[string]tftypes.Value{}

	err := in.As(&val)

	if err != nil {
		return nil, err
	}

	for k, v := range val {
		a, err := t.AttrTypes[k].ValueFromTerraform(ctx, v)

		if err != nil {
			return nil, err
		}

		attributes[k] = a
	}

	return NewLinksValueMust(LinksValue{}.AttributeTypes(ctx), attributes), nil
}

func (t LinksType) ValueType(ctx context.Context) attr.Value {
	return LinksValue{}
}

var _ basetypes.ObjectValuable = LinksValue{}

type LinksValue struct {
	Href  basetypes.StringValue `tfsdk:"href"`
	Rel   basetypes.StringValue `tfsdk:"rel"`
	state attr.ValueState
}

func (v LinksValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	attrTypes := make(map[string]tftypes.Type, 2)

	var val tftypes.Value
	var err error

	attrTypes["href"] = basetypes.StringType{}.TerraformType(ctx)
	attrTypes["rel"] = basetypes.StringType{}.TerraformType(ctx)

	objectType := tftypes.Object{AttributeTypes: attrTypes}

	switch v.state {
	case attr.ValueStateKnown:
		vals := make(map[string]tftypes.Value, 2)

		val, err = v.Href.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["href"] = val

		val, err = v.Rel.ToTerraformValue(ctx)

		if err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		vals["rel"] = val

		if err := tftypes.ValidateValue(objectType, vals); err != nil {
			return tftypes.NewValue(objectType, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(objectType, vals), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(objectType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(objectType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Object state in ToTerraformValue: %s", v.state))
	}
}

func (v LinksValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

func (v LinksValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

func (v LinksValue) String() string {
	return "LinksValue"
}

func (v LinksValue) ToObjectValue(ctx context.Context) (basetypes.ObjectValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	attributeTypes := map[string]attr.Type{
		"href": basetypes.StringType{},
		"rel":  basetypes.StringType{},
	}

	if v.IsNull() {
		return types.ObjectNull(attributeTypes), diags
	}

	if v.IsUnknown() {
		return types.ObjectUnknown(attributeTypes), diags
	}

	objVal, diags := types.ObjectValue(
		attributeTypes,
		map[string]attr.Value{
			"href": v.Href,
			"rel":  v.Rel,
		})

	return objVal, diags
}

func (v LinksValue) Equal(o attr.Value) bool {
	other, ok := o.(LinksValue)

	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	if !v.Href.Equal(other.Href) {
		return false
	}

	if !v.Rel.Equal(other.Rel) {
		return false
	}

	return true
}

func (v LinksValue) Type(ctx context.Context) attr.Type {
	return LinksType{
		basetypes.ObjectType{
			AttrTypes: v.AttributeTypes(ctx),
		},
	}
}

func (v LinksValue) AttributeTypes(ctx context.Context) map[string]attr.Type {
	return map[string]attr.Type{
		"href": basetypes.StringType{},
		"rel":  basetypes.StringType{},
	}
}
