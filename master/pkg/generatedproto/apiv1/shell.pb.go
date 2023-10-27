// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// source: determined/api/v1/shell.proto

package apiv1

import (
	shellv1 "github.com/determined-ai/determined/proto/pkg/shellv1"
	utilv1 "github.com/determined-ai/determined/proto/pkg/utilv1"
	_struct "github.com/golang/protobuf/ptypes/struct"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Sorts shells by the given field.
type GetShellsRequest_SortBy int32

const (
	// Returns shells in an unsorted list.
	GetShellsRequest_SORT_BY_UNSPECIFIED GetShellsRequest_SortBy = 0
	// Returns shells sorted by id.
	GetShellsRequest_SORT_BY_ID GetShellsRequest_SortBy = 1
	// Returns shells sorted by description.
	GetShellsRequest_SORT_BY_DESCRIPTION GetShellsRequest_SortBy = 2
	// Return shells sorted by start time.
	GetShellsRequest_SORT_BY_START_TIME GetShellsRequest_SortBy = 4
	// Return shells sorted by workspace_id.
	GetShellsRequest_SORT_BY_WORKSPACE_ID GetShellsRequest_SortBy = 5
)

// Enum value maps for GetShellsRequest_SortBy.
var (
	GetShellsRequest_SortBy_name = map[int32]string{
		0: "SORT_BY_UNSPECIFIED",
		1: "SORT_BY_ID",
		2: "SORT_BY_DESCRIPTION",
		4: "SORT_BY_START_TIME",
		5: "SORT_BY_WORKSPACE_ID",
	}
	GetShellsRequest_SortBy_value = map[string]int32{
		"SORT_BY_UNSPECIFIED":  0,
		"SORT_BY_ID":           1,
		"SORT_BY_DESCRIPTION":  2,
		"SORT_BY_START_TIME":   4,
		"SORT_BY_WORKSPACE_ID": 5,
	}
)

func (x GetShellsRequest_SortBy) Enum() *GetShellsRequest_SortBy {
	p := new(GetShellsRequest_SortBy)
	*p = x
	return p
}

func (x GetShellsRequest_SortBy) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (GetShellsRequest_SortBy) Descriptor() protoreflect.EnumDescriptor {
	return file_determined_api_v1_shell_proto_enumTypes[0].Descriptor()
}

func (GetShellsRequest_SortBy) Type() protoreflect.EnumType {
	return &file_determined_api_v1_shell_proto_enumTypes[0]
}

func (x GetShellsRequest_SortBy) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use GetShellsRequest_SortBy.Descriptor instead.
func (GetShellsRequest_SortBy) EnumDescriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{0, 0}
}

// Get a list of shells.
type GetShellsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Sort shells by the given field.
	SortBy GetShellsRequest_SortBy `protobuf:"varint,1,opt,name=sort_by,json=sortBy,proto3,enum=determined.api.v1.GetShellsRequest_SortBy" json:"sort_by,omitempty"`
	// Order shells in either ascending or descending order.
	OrderBy OrderBy `protobuf:"varint,2,opt,name=order_by,json=orderBy,proto3,enum=determined.api.v1.OrderBy" json:"order_by,omitempty"`
	// Skip the number of shells before returning results. Negative values
	// denote number of shells to skip from the end before returning results.
	Offset int32 `protobuf:"varint,3,opt,name=offset,proto3" json:"offset,omitempty"`
	// Limit the number of shells. A value of 0 denotes no limit.
	Limit int32 `protobuf:"varint,4,opt,name=limit,proto3" json:"limit,omitempty"`
	// Limit shells to those that are owned by users with the specified usernames.
	Users []string `protobuf:"bytes,5,rep,name=users,proto3" json:"users,omitempty"`
	// Limit shells to those that are owned by users with the specified userIds.
	UserIds []int32 `protobuf:"varint,6,rep,packed,name=user_ids,json=userIds,proto3" json:"user_ids,omitempty"`
	// Limit to those within a specified workspace, or 0 for all
	// accessible workspaces.
	WorkspaceId int32 `protobuf:"varint,7,opt,name=workspace_id,json=workspaceId,proto3" json:"workspace_id,omitempty"`
}

func (x *GetShellsRequest) Reset() {
	*x = GetShellsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetShellsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShellsRequest) ProtoMessage() {}

func (x *GetShellsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShellsRequest.ProtoReflect.Descriptor instead.
func (*GetShellsRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{0}
}

func (x *GetShellsRequest) GetSortBy() GetShellsRequest_SortBy {
	if x != nil {
		return x.SortBy
	}
	return GetShellsRequest_SORT_BY_UNSPECIFIED
}

func (x *GetShellsRequest) GetOrderBy() OrderBy {
	if x != nil {
		return x.OrderBy
	}
	return OrderBy_ORDER_BY_UNSPECIFIED
}

func (x *GetShellsRequest) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

func (x *GetShellsRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *GetShellsRequest) GetUsers() []string {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *GetShellsRequest) GetUserIds() []int32 {
	if x != nil {
		return x.UserIds
	}
	return nil
}

func (x *GetShellsRequest) GetWorkspaceId() int32 {
	if x != nil {
		return x.WorkspaceId
	}
	return 0
}

// Response to GetShellsRequest.
type GetShellsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The list of returned shells.
	Shells []*shellv1.Shell `protobuf:"bytes,1,rep,name=shells,proto3" json:"shells,omitempty"`
	// Pagination information of the full dataset.
	Pagination *Pagination `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (x *GetShellsResponse) Reset() {
	*x = GetShellsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetShellsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShellsResponse) ProtoMessage() {}

func (x *GetShellsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShellsResponse.ProtoReflect.Descriptor instead.
func (*GetShellsResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{1}
}

func (x *GetShellsResponse) GetShells() []*shellv1.Shell {
	if x != nil {
		return x.Shells
	}
	return nil
}

func (x *GetShellsResponse) GetPagination() *Pagination {
	if x != nil {
		return x.Pagination
	}
	return nil
}

// Get the requested shell.
type GetShellRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the shell.
	ShellId string `protobuf:"bytes,1,opt,name=shell_id,json=shellId,proto3" json:"shell_id,omitempty"`
}

func (x *GetShellRequest) Reset() {
	*x = GetShellRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetShellRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShellRequest) ProtoMessage() {}

func (x *GetShellRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShellRequest.ProtoReflect.Descriptor instead.
func (*GetShellRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{2}
}

func (x *GetShellRequest) GetShellId() string {
	if x != nil {
		return x.ShellId
	}
	return ""
}

// Response to GetShellRequest.
type GetShellResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The requested shell.
	Shell *shellv1.Shell `protobuf:"bytes,1,opt,name=shell,proto3" json:"shell,omitempty"`
	// The shell config.
	Config *_struct.Struct `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
}

func (x *GetShellResponse) Reset() {
	*x = GetShellResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetShellResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetShellResponse) ProtoMessage() {}

func (x *GetShellResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetShellResponse.ProtoReflect.Descriptor instead.
func (*GetShellResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{3}
}

func (x *GetShellResponse) GetShell() *shellv1.Shell {
	if x != nil {
		return x.Shell
	}
	return nil
}

func (x *GetShellResponse) GetConfig() *_struct.Struct {
	if x != nil {
		return x.Config
	}
	return nil
}

// Kill the requested shell.
type KillShellRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the shell.
	ShellId string `protobuf:"bytes,1,opt,name=shell_id,json=shellId,proto3" json:"shell_id,omitempty"`
}

func (x *KillShellRequest) Reset() {
	*x = KillShellRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KillShellRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KillShellRequest) ProtoMessage() {}

func (x *KillShellRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KillShellRequest.ProtoReflect.Descriptor instead.
func (*KillShellRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{4}
}

func (x *KillShellRequest) GetShellId() string {
	if x != nil {
		return x.ShellId
	}
	return ""
}

// Response to KillShellRequest.
type KillShellResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The requested shell.
	Shell *shellv1.Shell `protobuf:"bytes,1,opt,name=shell,proto3" json:"shell,omitempty"`
}

func (x *KillShellResponse) Reset() {
	*x = KillShellResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KillShellResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KillShellResponse) ProtoMessage() {}

func (x *KillShellResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KillShellResponse.ProtoReflect.Descriptor instead.
func (*KillShellResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{5}
}

func (x *KillShellResponse) GetShell() *shellv1.Shell {
	if x != nil {
		return x.Shell
	}
	return nil
}

// Set the priority of the requested shell.
type SetShellPriorityRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the shell.
	ShellId string `protobuf:"bytes,1,opt,name=shell_id,json=shellId,proto3" json:"shell_id,omitempty"`
	// The new priority.
	Priority int32 `protobuf:"varint,2,opt,name=priority,proto3" json:"priority,omitempty"`
}

func (x *SetShellPriorityRequest) Reset() {
	*x = SetShellPriorityRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetShellPriorityRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetShellPriorityRequest) ProtoMessage() {}

func (x *SetShellPriorityRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetShellPriorityRequest.ProtoReflect.Descriptor instead.
func (*SetShellPriorityRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{6}
}

func (x *SetShellPriorityRequest) GetShellId() string {
	if x != nil {
		return x.ShellId
	}
	return ""
}

func (x *SetShellPriorityRequest) GetPriority() int32 {
	if x != nil {
		return x.Priority
	}
	return 0
}

// Response to SetShellPriorityRequest.
type SetShellPriorityResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The requested shell.
	Shell *shellv1.Shell `protobuf:"bytes,1,opt,name=shell,proto3" json:"shell,omitempty"`
}

func (x *SetShellPriorityResponse) Reset() {
	*x = SetShellPriorityResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetShellPriorityResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetShellPriorityResponse) ProtoMessage() {}

func (x *SetShellPriorityResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetShellPriorityResponse.ProtoReflect.Descriptor instead.
func (*SetShellPriorityResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{7}
}

func (x *SetShellPriorityResponse) GetShell() *shellv1.Shell {
	if x != nil {
		return x.Shell
	}
	return nil
}

// Request to launch a shell.
type LaunchShellRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Shell config (JSON).
	Config *_struct.Struct `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	// Template name.
	TemplateName string `protobuf:"bytes,2,opt,name=template_name,json=templateName,proto3" json:"template_name,omitempty"`
	// The files to run with the command.
	Files []*utilv1.File `protobuf:"bytes,3,rep,name=files,proto3" json:"files,omitempty"`
	// Additional data.
	Data []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	// Workspace ID. Defaults to 'Uncategorized' workspace if not specified.
	WorkspaceId int32 `protobuf:"varint,5,opt,name=workspace_id,json=workspaceId,proto3" json:"workspace_id,omitempty"`
}

func (x *LaunchShellRequest) Reset() {
	*x = LaunchShellRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LaunchShellRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LaunchShellRequest) ProtoMessage() {}

func (x *LaunchShellRequest) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LaunchShellRequest.ProtoReflect.Descriptor instead.
func (*LaunchShellRequest) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{8}
}

func (x *LaunchShellRequest) GetConfig() *_struct.Struct {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *LaunchShellRequest) GetTemplateName() string {
	if x != nil {
		return x.TemplateName
	}
	return ""
}

func (x *LaunchShellRequest) GetFiles() []*utilv1.File {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *LaunchShellRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *LaunchShellRequest) GetWorkspaceId() int32 {
	if x != nil {
		return x.WorkspaceId
	}
	return 0
}

// Response to LaunchShellRequest.
type LaunchShellResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The requested shell.
	Shell *shellv1.Shell `protobuf:"bytes,1,opt,name=shell,proto3" json:"shell,omitempty"`
	// The config;
	Config *_struct.Struct `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	// List of any related warnings.
	Warnings []LaunchWarning `protobuf:"varint,3,rep,packed,name=warnings,proto3,enum=determined.api.v1.LaunchWarning" json:"warnings,omitempty"`
}

func (x *LaunchShellResponse) Reset() {
	*x = LaunchShellResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_determined_api_v1_shell_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LaunchShellResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LaunchShellResponse) ProtoMessage() {}

func (x *LaunchShellResponse) ProtoReflect() protoreflect.Message {
	mi := &file_determined_api_v1_shell_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LaunchShellResponse.ProtoReflect.Descriptor instead.
func (*LaunchShellResponse) Descriptor() ([]byte, []int) {
	return file_determined_api_v1_shell_proto_rawDescGZIP(), []int{9}
}

func (x *LaunchShellResponse) GetShell() *shellv1.Shell {
	if x != nil {
		return x.Shell
	}
	return nil
}

func (x *LaunchShellResponse) GetConfig() *_struct.Struct {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *LaunchShellResponse) GetWarnings() []LaunchWarning {
	if x != nil {
		return x.Warnings
	}
	return nil
}

var File_determined_api_v1_shell_proto protoreflect.FileDescriptor

var file_determined_api_v1_shell_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x11, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x22, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x76, 0x31, 0x2f, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65,
	0x64, 0x2f, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x68, 0x65, 0x6c, 0x6c,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e,
	0x65, 0x64, 0x2f, 0x75, 0x74, 0x69, 0x6c, 0x2f, 0x76, 0x31, 0x2f, 0x75, 0x74, 0x69, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2c, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65,
	0x6e, 0x2d, 0x73, 0x77, 0x61, 0x67, 0x67, 0x65, 0x72, 0x2f, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x8e, 0x03, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x43, 0x0a, 0x07, 0x73, 0x6f, 0x72, 0x74,
	0x5f, 0x62, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x2a, 0x2e, 0x64, 0x65, 0x74, 0x65,
	0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65,
	0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x2e, 0x53,
	0x6f, 0x72, 0x74, 0x42, 0x79, 0x52, 0x06, 0x73, 0x6f, 0x72, 0x74, 0x42, 0x79, 0x12, 0x35, 0x0a,
	0x08, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x5f, 0x62, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x1a, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x76, 0x31, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x79, 0x52, 0x07, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x42, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x12, 0x14, 0x0a, 0x05,
	0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d,
	0x69, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x09, 0x52, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x12, 0x19, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72,
	0x5f, 0x69, 0x64, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x05, 0x52, 0x07, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x64, 0x73, 0x12, 0x21, 0x0a, 0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x77, 0x6f, 0x72, 0x6b, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x49, 0x64, 0x22, 0x7c, 0x0a, 0x06, 0x53, 0x6f, 0x72, 0x74, 0x42, 0x79,
	0x12, 0x17, 0x0a, 0x13, 0x53, 0x4f, 0x52, 0x54, 0x5f, 0x42, 0x59, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x4f, 0x52,
	0x54, 0x5f, 0x42, 0x59, 0x5f, 0x49, 0x44, 0x10, 0x01, 0x12, 0x17, 0x0a, 0x13, 0x53, 0x4f, 0x52,
	0x54, 0x5f, 0x42, 0x59, 0x5f, 0x44, 0x45, 0x53, 0x43, 0x52, 0x49, 0x50, 0x54, 0x49, 0x4f, 0x4e,
	0x10, 0x02, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x4f, 0x52, 0x54, 0x5f, 0x42, 0x59, 0x5f, 0x53, 0x54,
	0x41, 0x52, 0x54, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x04, 0x12, 0x18, 0x0a, 0x14, 0x53, 0x4f,
	0x52, 0x54, 0x5f, 0x42, 0x59, 0x5f, 0x57, 0x4f, 0x52, 0x4b, 0x53, 0x50, 0x41, 0x43, 0x45, 0x5f,
	0x49, 0x44, 0x10, 0x05, 0x22, 0x96, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c,
	0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x32, 0x0a, 0x06, 0x73, 0x68,
	0x65, 0x6c, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x64, 0x65, 0x74,
	0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x2e, 0x76, 0x31,
	0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x06, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x73, 0x12, 0x3d,
	0x0a, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x0a, 0x70, 0x61, 0x67, 0x69, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x3a, 0x0e, 0x92,
	0x41, 0x0b, 0x0a, 0x09, 0xd2, 0x01, 0x06, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x73, 0x22, 0x2c, 0x0a,
	0x0f, 0x47, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x19, 0x0a, 0x08, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x49, 0x64, 0x22, 0x8d, 0x01, 0x0a, 0x10,
	0x47, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x30, 0x0a, 0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x1a, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x73, 0x68, 0x65,
	0x6c, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x05, 0x73, 0x68, 0x65,
	0x6c, 0x6c, 0x12, 0x2f, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x06, 0x63, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x3a, 0x16, 0x92, 0x41, 0x13, 0x0a, 0x11, 0xd2, 0x01, 0x05, 0x73, 0x68, 0x65,
	0x6c, 0x6c, 0xd2, 0x01, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x2d, 0x0a, 0x10, 0x4b,
	0x69, 0x6c, 0x6c, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x19, 0x0a, 0x08, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x49, 0x64, 0x22, 0x45, 0x0a, 0x11, 0x4b, 0x69,
	0x6c, 0x6c, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x30, 0x0a, 0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x73, 0x68, 0x65, 0x6c,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x05, 0x73, 0x68, 0x65, 0x6c,
	0x6c, 0x22, 0x50, 0x0a, 0x17, 0x53, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69,
	0x6f, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x08,
	0x73, 0x68, 0x65, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x73, 0x68, 0x65, 0x6c, 0x6c, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72,
	0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x72, 0x69, 0x6f, 0x72,
	0x69, 0x74, 0x79, 0x22, 0x4c, 0x0a, 0x18, 0x53, 0x65, 0x74, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x50,
	0x72, 0x69, 0x6f, 0x72, 0x69, 0x74, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x30, 0x0a, 0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a,
	0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x73, 0x68, 0x65, 0x6c,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x05, 0x73, 0x68, 0x65, 0x6c,
	0x6c, 0x22, 0xd1, 0x01, 0x0a, 0x12, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x53, 0x68, 0x65, 0x6c,
	0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2f, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63,
	0x74, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x23, 0x0a, 0x0d, 0x74, 0x65, 0x6d,
	0x70, 0x6c, 0x61, 0x74, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x74, 0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e,
	0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x75, 0x74, 0x69, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x12,
	0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61,
	0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x5f,
	0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x49, 0x64, 0x22, 0xce, 0x01, 0x0a, 0x13, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68,
	0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x30, 0x0a,
	0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x64,
	0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x53, 0x68, 0x65, 0x6c, 0x6c, 0x52, 0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0x12,
	0x2f, 0x0a, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x06, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x12, 0x3c, 0x0a, 0x08, 0x77, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0e, 0x32, 0x20, 0x2e, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x61, 0x75, 0x6e, 0x63, 0x68, 0x57, 0x61, 0x72,
	0x6e, 0x69, 0x6e, 0x67, 0x52, 0x08, 0x77, 0x61, 0x72, 0x6e, 0x69, 0x6e, 0x67, 0x73, 0x3a, 0x16,
	0x92, 0x41, 0x13, 0x0a, 0x11, 0xd2, 0x01, 0x05, 0x73, 0x68, 0x65, 0x6c, 0x6c, 0xd2, 0x01, 0x06,
	0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x42, 0x35, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2d,
	0x61, 0x69, 0x2f, 0x64, 0x65, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x65, 0x64, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x76, 0x31, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_determined_api_v1_shell_proto_rawDescOnce sync.Once
	file_determined_api_v1_shell_proto_rawDescData = file_determined_api_v1_shell_proto_rawDesc
)

func file_determined_api_v1_shell_proto_rawDescGZIP() []byte {
	file_determined_api_v1_shell_proto_rawDescOnce.Do(func() {
		file_determined_api_v1_shell_proto_rawDescData = protoimpl.X.CompressGZIP(file_determined_api_v1_shell_proto_rawDescData)
	})
	return file_determined_api_v1_shell_proto_rawDescData
}

var file_determined_api_v1_shell_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_determined_api_v1_shell_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_determined_api_v1_shell_proto_goTypes = []interface{}{
	(GetShellsRequest_SortBy)(0),     // 0: determined.api.v1.GetShellsRequest.SortBy
	(*GetShellsRequest)(nil),         // 1: determined.api.v1.GetShellsRequest
	(*GetShellsResponse)(nil),        // 2: determined.api.v1.GetShellsResponse
	(*GetShellRequest)(nil),          // 3: determined.api.v1.GetShellRequest
	(*GetShellResponse)(nil),         // 4: determined.api.v1.GetShellResponse
	(*KillShellRequest)(nil),         // 5: determined.api.v1.KillShellRequest
	(*KillShellResponse)(nil),        // 6: determined.api.v1.KillShellResponse
	(*SetShellPriorityRequest)(nil),  // 7: determined.api.v1.SetShellPriorityRequest
	(*SetShellPriorityResponse)(nil), // 8: determined.api.v1.SetShellPriorityResponse
	(*LaunchShellRequest)(nil),       // 9: determined.api.v1.LaunchShellRequest
	(*LaunchShellResponse)(nil),      // 10: determined.api.v1.LaunchShellResponse
	(OrderBy)(0),                     // 11: determined.api.v1.OrderBy
	(*shellv1.Shell)(nil),            // 12: determined.shell.v1.Shell
	(*Pagination)(nil),               // 13: determined.api.v1.Pagination
	(*_struct.Struct)(nil),           // 14: google.protobuf.Struct
	(*utilv1.File)(nil),              // 15: determined.util.v1.File
	(LaunchWarning)(0),               // 16: determined.api.v1.LaunchWarning
}
var file_determined_api_v1_shell_proto_depIdxs = []int32{
	0,  // 0: determined.api.v1.GetShellsRequest.sort_by:type_name -> determined.api.v1.GetShellsRequest.SortBy
	11, // 1: determined.api.v1.GetShellsRequest.order_by:type_name -> determined.api.v1.OrderBy
	12, // 2: determined.api.v1.GetShellsResponse.shells:type_name -> determined.shell.v1.Shell
	13, // 3: determined.api.v1.GetShellsResponse.pagination:type_name -> determined.api.v1.Pagination
	12, // 4: determined.api.v1.GetShellResponse.shell:type_name -> determined.shell.v1.Shell
	14, // 5: determined.api.v1.GetShellResponse.config:type_name -> google.protobuf.Struct
	12, // 6: determined.api.v1.KillShellResponse.shell:type_name -> determined.shell.v1.Shell
	12, // 7: determined.api.v1.SetShellPriorityResponse.shell:type_name -> determined.shell.v1.Shell
	14, // 8: determined.api.v1.LaunchShellRequest.config:type_name -> google.protobuf.Struct
	15, // 9: determined.api.v1.LaunchShellRequest.files:type_name -> determined.util.v1.File
	12, // 10: determined.api.v1.LaunchShellResponse.shell:type_name -> determined.shell.v1.Shell
	14, // 11: determined.api.v1.LaunchShellResponse.config:type_name -> google.protobuf.Struct
	16, // 12: determined.api.v1.LaunchShellResponse.warnings:type_name -> determined.api.v1.LaunchWarning
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_determined_api_v1_shell_proto_init() }
func file_determined_api_v1_shell_proto_init() {
	if File_determined_api_v1_shell_proto != nil {
		return
	}
	file_determined_api_v1_command_proto_init()
	file_determined_api_v1_pagination_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_determined_api_v1_shell_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetShellsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetShellsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetShellRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetShellResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KillShellRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KillShellResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetShellPriorityRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetShellPriorityResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LaunchShellRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_determined_api_v1_shell_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LaunchShellResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_determined_api_v1_shell_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_determined_api_v1_shell_proto_goTypes,
		DependencyIndexes: file_determined_api_v1_shell_proto_depIdxs,
		EnumInfos:         file_determined_api_v1_shell_proto_enumTypes,
		MessageInfos:      file_determined_api_v1_shell_proto_msgTypes,
	}.Build()
	File_determined_api_v1_shell_proto = out.File
	file_determined_api_v1_shell_proto_rawDesc = nil
	file_determined_api_v1_shell_proto_goTypes = nil
	file_determined_api_v1_shell_proto_depIdxs = nil
}
