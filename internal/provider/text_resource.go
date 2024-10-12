// Copyright (c) Plex, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TextResource{}
var _ resource.ResourceWithImportState = &TextResource{}

func NewTextResource() resource.Resource {
	return &TextResource{}
}

// TextResource defines the resource implementation.
type TextResource struct {
	providerData ProviderData
}

type SecretPayload struct {
	Password   string  `json:"payload"`
	Passphrase *string `json:"passphrase"`
	// ExpireAfterDays   int    `json:"expire_after_days"`
	// ExpireAfterViews  int    `json:"expire_after_views"`
	DeletableByViewer bool   `json:"deletable_by_viewer"`
	RetrievalStep     bool   `json:"retrieval_step"`
	Kind              string `json:"kind"`
}

// Secret -
type Secret struct {
	ID                string `json:"url_token"`
	ExpireAfterDays   int    `json:"expire_after_days"`
	ExpireAfterViews  int    `json:"expire_after_views"`
	Expired           bool   `json:"expired"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	Deleted           bool   `json:"deleted"`
	DeletableByViewer bool   `json:"deletable_by_viewer"`
	RetrievalStep     bool   `json:"retrieval_step"`
	ExpiredAt         string `json:"expired_on"`
	DaysRemaining     int    `json:"days_remaining"`
	ViewsRemaining    int    `json:"views_remaining"`
}

// TextResourceModel describes the resource data model.
type TextResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Password          types.String `tfsdk:"password"`
	Passphrase        *string      `tfsdk:"passphrase"`
	ExpireAfterDays   types.Int32  `tfsdk:"expire_after_days"`
	ExpireAfterViews  types.Int32  `tfsdk:"expire_after_views"`
	Expired           types.Bool   `tfsdk:"expired"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	Deleted           types.Bool   `tfsdk:"deleted"`
	DeletableByViewer types.Bool   `tfsdk:"deletable_by_viewer"`
	RetrievalStep     types.Bool   `tfsdk:"retrieval_step"`
	ExpiredAt         types.String `tfsdk:"expired_on"`
	DaysRemaining     types.Int32  `tfsdk:"days_remaining"`
	ViewsRemaining    types.Int32  `tfsdk:"views_remaining"`
}

func (r *TextResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_text"
}

func (r *TextResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The Text resource that will get pushed to the secret server",

		Attributes: map[string]schema.Attribute{
			"password": schema.StringAttribute{
				MarkdownDescription: "The password payload",
				Required:            true,
				Sensitive:           true,
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "Require recipients to enter this passphrase to view the created item",
				Optional:            true,
				Sensitive:           true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier of the secret in the pwpusher app",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expire_after_days": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Expire secret link and delete after this many days",
			},
			"expire_after_views": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Expire secret link and delete after this many views",
			},
			"expired": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "If the secret has expired",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp that the secret was created",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp that the secret was updated",
			},
			"deleted": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "If the secret has been deleted",
			},
			"deletable_by_viewer": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Allow users to delete passwords once retrieved",
			},
			"retrieval_step": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Helps to avoid chat systems and URL scanners from eating up views",
			},
			"expired_on": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp that the secret expired",
			},
			"days_remaining": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The number of days left that the secret can be viewed",
			},
			"views_remaining": schema.Int32Attribute{
				Computed:            true,
				MarkdownDescription: "The number of times that the secret can be viewed",
			},
		},
	}
}

func (r *TextResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = providerData
}

func (r *TextResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TextResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	payload := SecretPayload{
		Password:   data.Password.ValueString(),
		Passphrase: data.Passphrase,
		// ExpireAfterDays:   data.ExpireAfterDays.ValueInt32Pointer(),
		// ExpireAfterViews:  int(data.ExpireAfterViews.ValueInt32()),
		DeletableByViewer: data.DeletableByViewer.ValueBool(),
		RetrievalStep:     data.RetrievalStep.ValueBool(),
		Kind:              "text",
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	res, err := r.providerData.client.Post(r.providerData.url.ValueString()+"/p.json", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return
	}

	newSecret := Secret{}
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &newSecret)
	if err != nil {
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	tflog.Trace(ctx, bodyString)

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue(newSecret.ID)
	data.ExpireAfterDays = types.Int32Value(int32(newSecret.ExpireAfterDays))
	data.ExpireAfterViews = types.Int32Value(int32(newSecret.ExpireAfterViews))
	data.Expired = types.BoolValue(newSecret.Expired)
	data.CreatedAt = types.StringValue(newSecret.CreatedAt)
	data.UpdatedAt = types.StringValue(newSecret.UpdatedAt)
	data.Deleted = types.BoolValue(newSecret.Deleted)
	data.DeletableByViewer = types.BoolValue(newSecret.DeletableByViewer)
	data.RetrievalStep = types.BoolValue(newSecret.RetrievalStep)
	data.ExpiredAt = types.StringValue(newSecret.ExpiredAt)
	data.DaysRemaining = types.Int32Value(int32(newSecret.DaysRemaining))
	data.ViewsRemaining = types.Int32Value(int32(newSecret.ViewsRemaining))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log\
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TextResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TextResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TextResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TextResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update entry, not a permitted action"))
	if resp.Diagnostics.HasError() {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TextResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TextResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *TextResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import entry, not a permitted action"))

}
