// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: proto.proto

package prototictactoe

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type CreateGameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GameID        string `protobuf:"bytes,1,opt,name=gameID,proto3" json:"gameID,omitempty"`
	FirstPlayerID string `protobuf:"bytes,2,opt,name=firstPlayerID,proto3" json:"firstPlayerID,omitempty"`
}

func (x *CreateGameRequest) Reset() {
	*x = CreateGameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateGameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateGameRequest) ProtoMessage() {}

func (x *CreateGameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateGameRequest.ProtoReflect.Descriptor instead.
func (*CreateGameRequest) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{0}
}

func (x *CreateGameRequest) GetGameID() string {
	if x != nil {
		return x.GameID
	}
	return ""
}

func (x *CreateGameRequest) GetFirstPlayerID() string {
	if x != nil {
		return x.FirstPlayerID
	}
	return ""
}

type CreateGameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *GameState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *CreateGameResponse) Reset() {
	*x = CreateGameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateGameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateGameResponse) ProtoMessage() {}

func (x *CreateGameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateGameResponse.ProtoReflect.Descriptor instead.
func (*CreateGameResponse) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{1}
}

func (x *CreateGameResponse) GetState() *GameState {
	if x != nil {
		return x.State
	}
	return nil
}

type JoinGameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GameID         string `protobuf:"bytes,1,opt,name=gameID,proto3" json:"gameID,omitempty"`
	SecondPlayerID string `protobuf:"bytes,2,opt,name=secondPlayerID,proto3" json:"secondPlayerID,omitempty"`
}

func (x *JoinGameRequest) Reset() {
	*x = JoinGameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinGameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinGameRequest) ProtoMessage() {}

func (x *JoinGameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinGameRequest.ProtoReflect.Descriptor instead.
func (*JoinGameRequest) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{2}
}

func (x *JoinGameRequest) GetGameID() string {
	if x != nil {
		return x.GameID
	}
	return ""
}

func (x *JoinGameRequest) GetSecondPlayerID() string {
	if x != nil {
		return x.SecondPlayerID
	}
	return ""
}

type JoinGameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *GameState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *JoinGameResponse) Reset() {
	*x = JoinGameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinGameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinGameResponse) ProtoMessage() {}

func (x *JoinGameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinGameResponse.ProtoReflect.Descriptor instead.
func (*JoinGameResponse) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{3}
}

func (x *JoinGameResponse) GetState() *GameState {
	if x != nil {
		return x.State
	}
	return nil
}

type MoveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GameID   string `protobuf:"bytes,1,opt,name=gameID,proto3" json:"gameID,omitempty"`
	PlayerID string `protobuf:"bytes,2,opt,name=playerID,proto3" json:"playerID,omitempty"`
	Move     string `protobuf:"bytes,3,opt,name=move,proto3" json:"move,omitempty"`
}

func (x *MoveRequest) Reset() {
	*x = MoveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MoveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoveRequest) ProtoMessage() {}

func (x *MoveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoveRequest.ProtoReflect.Descriptor instead.
func (*MoveRequest) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{4}
}

func (x *MoveRequest) GetGameID() string {
	if x != nil {
		return x.GameID
	}
	return ""
}

func (x *MoveRequest) GetPlayerID() string {
	if x != nil {
		return x.PlayerID
	}
	return ""
}

func (x *MoveRequest) GetMove() string {
	if x != nil {
		return x.Move
	}
	return ""
}

type MoveResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *GameState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *MoveResponse) Reset() {
	*x = MoveResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MoveResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MoveResponse) ProtoMessage() {}

func (x *MoveResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MoveResponse.ProtoReflect.Descriptor instead.
func (*MoveResponse) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{5}
}

func (x *MoveResponse) GetState() *GameState {
	if x != nil {
		return x.State
	}
	return nil
}

type GetGameRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	GameID string `protobuf:"bytes,1,opt,name=gameID,proto3" json:"gameID,omitempty"`
}

func (x *GetGameRequest) Reset() {
	*x = GetGameRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetGameRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGameRequest) ProtoMessage() {}

func (x *GetGameRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGameRequest.ProtoReflect.Descriptor instead.
func (*GetGameRequest) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{6}
}

func (x *GetGameRequest) GetGameID() string {
	if x != nil {
		return x.GameID
	}
	return ""
}

type GetGameResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	State *GameState `protobuf:"bytes,1,opt,name=state,proto3" json:"state,omitempty"`
}

func (x *GetGameResponse) Reset() {
	*x = GetGameResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetGameResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetGameResponse) ProtoMessage() {}

func (x *GetGameResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetGameResponse.ProtoReflect.Descriptor instead.
func (*GetGameResponse) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{7}
}

func (x *GetGameResponse) GetState() *GameState {
	if x != nil {
		return x.State
	}
	return nil
}

type GameState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to State:
	//	*GameState_Waiting
	//	*GameState_Progress
	//	*GameState_Result
	State isGameState_State `protobuf_oneof:"state"`
}

func (x *GameState) Reset() {
	*x = GameState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameState) ProtoMessage() {}

func (x *GameState) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameState.ProtoReflect.Descriptor instead.
func (*GameState) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{8}
}

func (m *GameState) GetState() isGameState_State {
	if m != nil {
		return m.State
	}
	return nil
}

func (x *GameState) GetWaiting() *GameState_GameWaitingForPlayer {
	if x, ok := x.GetState().(*GameState_Waiting); ok {
		return x.Waiting
	}
	return nil
}

func (x *GameState) GetProgress() *GameState_GameProgress {
	if x, ok := x.GetState().(*GameState_Progress); ok {
		return x.Progress
	}
	return nil
}

func (x *GameState) GetResult() *GameState_GameResult {
	if x, ok := x.GetState().(*GameState_Result); ok {
		return x.Result
	}
	return nil
}

type isGameState_State interface {
	isGameState_State()
}

type GameState_Waiting struct {
	Waiting *GameState_GameWaitingForPlayer `protobuf:"bytes,1,opt,name=waiting,proto3,oneof"`
}

type GameState_Progress struct {
	Progress *GameState_GameProgress `protobuf:"bytes,2,opt,name=progress,proto3,oneof"`
}

type GameState_Result struct {
	Result *GameState_GameResult `protobuf:"bytes,3,opt,name=result,proto3,oneof"`
}

func (*GameState_Waiting) isGameState_State() {}

func (*GameState_Progress) isGameState_State() {}

func (*GameState_Result) isGameState_State() {}

type GameState_GameWaitingForPlayer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NeedsPlayers int32 `protobuf:"zigzag32,1,opt,name=needsPlayers,proto3" json:"needsPlayers,omitempty"`
}

func (x *GameState_GameWaitingForPlayer) Reset() {
	*x = GameState_GameWaitingForPlayer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameState_GameWaitingForPlayer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameState_GameWaitingForPlayer) ProtoMessage() {}

func (x *GameState_GameWaitingForPlayer) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameState_GameWaitingForPlayer.ProtoReflect.Descriptor instead.
func (*GameState_GameWaitingForPlayer) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{8, 0}
}

func (x *GameState_GameWaitingForPlayer) GetNeedsPlayers() int32 {
	if x != nil {
		return x.NeedsPlayers
	}
	return 0
}

type GameState_GameProgress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NextMovePlayerID string   `protobuf:"bytes,1,opt,name=nextMovePlayerID,proto3" json:"nextMovePlayerID,omitempty"`
	AvailableMoves   []string `protobuf:"bytes,2,rep,name=availableMoves,proto3" json:"availableMoves,omitempty"`
}

func (x *GameState_GameProgress) Reset() {
	*x = GameState_GameProgress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameState_GameProgress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameState_GameProgress) ProtoMessage() {}

func (x *GameState_GameProgress) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameState_GameProgress.ProtoReflect.Descriptor instead.
func (*GameState_GameProgress) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{8, 1}
}

func (x *GameState_GameProgress) GetNextMovePlayerID() string {
	if x != nil {
		return x.NextMovePlayerID
	}
	return ""
}

func (x *GameState_GameProgress) GetAvailableMoves() []string {
	if x != nil {
		return x.AvailableMoves
	}
	return nil
}

type GameState_GameResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Winner         string   `protobuf:"bytes,1,opt,name=winner,proto3" json:"winner,omitempty"`
	WiningSequence []string `protobuf:"bytes,2,rep,name=winingSequence,proto3" json:"winingSequence,omitempty"`
}

func (x *GameState_GameResult) Reset() {
	*x = GameState_GameResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GameState_GameResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GameState_GameResult) ProtoMessage() {}

func (x *GameState_GameResult) ProtoReflect() protoreflect.Message {
	mi := &file_proto_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GameState_GameResult.ProtoReflect.Descriptor instead.
func (*GameState_GameResult) Descriptor() ([]byte, []int) {
	return file_proto_proto_rawDescGZIP(), []int{8, 2}
}

func (x *GameState_GameResult) GetWinner() string {
	if x != nil {
		return x.Winner
	}
	return ""
}

func (x *GameState_GameResult) GetWiningSequence() []string {
	if x != nil {
		return x.WiningSequence
	}
	return nil
}

var File_proto_proto protoreflect.FileDescriptor

var file_proto_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x51, 0x0a,
	0x11, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x67, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x67, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x12, 0x24, 0x0a, 0x0d, 0x66, 0x69,
	0x72, 0x73, 0x74, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x66, 0x69, 0x72, 0x73, 0x74, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x44,
	0x22, 0x36, 0x0a, 0x12, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74,
	0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x51, 0x0a, 0x0f, 0x4a, 0x6f, 0x69, 0x6e,
	0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x67,
	0x61, 0x6d, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67, 0x61, 0x6d,
	0x65, 0x49, 0x44, 0x12, 0x26, 0x0a, 0x0e, 0x73, 0x65, 0x63, 0x6f, 0x6e, 0x64, 0x50, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x73, 0x65, 0x63,
	0x6f, 0x6e, 0x64, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x44, 0x22, 0x34, 0x0a, 0x10, 0x4a,
	0x6f, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x20, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a,
	0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x22, 0x55, 0x0a, 0x0b, 0x4d, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x67, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x67, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x49, 0x44, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6d, 0x6f, 0x76, 0x65, 0x22, 0x30, 0x0a, 0x0c, 0x4d, 0x6f, 0x76, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0x28, 0x0a, 0x0e, 0x47, 0x65,
	0x74, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x67, 0x61, 0x6d, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x67, 0x61,
	0x6d, 0x65, 0x49, 0x44, 0x22, 0x33, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x47, 0x61, 0x6d, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x20, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x22, 0xa7, 0x03, 0x0a, 0x09, 0x47, 0x61,
	0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x3b, 0x0a, 0x07, 0x77, 0x61, 0x69, 0x74, 0x69,
	0x6e, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x57, 0x61, 0x69, 0x74, 0x69, 0x6e, 0x67,
	0x46, 0x6f, 0x72, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x48, 0x00, 0x52, 0x07, 0x77, 0x61, 0x69,
	0x74, 0x69, 0x6e, 0x67, 0x12, 0x35, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x48,
	0x00, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2f, 0x0a, 0x06, 0x72,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x47, 0x61,
	0x6d, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x2e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x48, 0x00, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x1a, 0x3a, 0x0a, 0x14,
	0x47, 0x61, 0x6d, 0x65, 0x57, 0x61, 0x69, 0x74, 0x69, 0x6e, 0x67, 0x46, 0x6f, 0x72, 0x50, 0x6c,
	0x61, 0x79, 0x65, 0x72, 0x12, 0x22, 0x0a, 0x0c, 0x6e, 0x65, 0x65, 0x64, 0x73, 0x50, 0x6c, 0x61,
	0x79, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x11, 0x52, 0x0c, 0x6e, 0x65, 0x65, 0x64,
	0x73, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x1a, 0x62, 0x0a, 0x0c, 0x47, 0x61, 0x6d, 0x65,
	0x50, 0x72, 0x6f, 0x67, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2a, 0x0a, 0x10, 0x6e, 0x65, 0x78, 0x74,
	0x4d, 0x6f, 0x76, 0x65, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x10, 0x6e, 0x65, 0x78, 0x74, 0x4d, 0x6f, 0x76, 0x65, 0x50, 0x6c, 0x61, 0x79,
	0x65, 0x72, 0x49, 0x44, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x76, 0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c,
	0x65, 0x4d, 0x6f, 0x76, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x61, 0x76,
	0x61, 0x69, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x4d, 0x6f, 0x76, 0x65, 0x73, 0x1a, 0x4c, 0x0a, 0x0a,
	0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x69,
	0x6e, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x77, 0x69, 0x6e, 0x6e,
	0x65, 0x72, 0x12, 0x26, 0x0a, 0x0e, 0x77, 0x69, 0x6e, 0x69, 0x6e, 0x67, 0x53, 0x65, 0x71, 0x75,
	0x65, 0x6e, 0x63, 0x65, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0e, 0x77, 0x69, 0x6e, 0x69,
	0x6e, 0x67, 0x53, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x42, 0x07, 0x0a, 0x05, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x32, 0xd7, 0x01, 0x0a, 0x12, 0x54, 0x69, 0x63, 0x54, 0x61, 0x63, 0x54, 0x6f,
	0x65, 0x41, 0x67, 0x67, 0x72, 0x65, 0x67, 0x61, 0x74, 0x65, 0x12, 0x37, 0x0a, 0x0a, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x00, 0x12, 0x31, 0x0a, 0x08, 0x4a, 0x6f, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x12,
	0x10, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x11, 0x2e, 0x4a, 0x6f, 0x69, 0x6e, 0x47, 0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x25, 0x0a, 0x04, 0x4d, 0x6f, 0x76, 0x65, 0x12, 0x0c,
	0x2e, 0x4d, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x4d,
	0x6f, 0x76, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x2e, 0x0a,
	0x07, 0x47, 0x65, 0x74, 0x47, 0x61, 0x6d, 0x65, 0x12, 0x0f, 0x2e, 0x47, 0x65, 0x74, 0x47, 0x61,
	0x6d, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10, 0x2e, 0x47, 0x65, 0x74, 0x47,
	0x61, 0x6d, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x12, 0x5a,
	0x10, 0x2e, 0x3b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x74, 0x69, 0x63, 0x74, 0x61, 0x63, 0x74, 0x6f,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_proto_rawDescOnce sync.Once
	file_proto_proto_rawDescData = file_proto_proto_rawDesc
)

func file_proto_proto_rawDescGZIP() []byte {
	file_proto_proto_rawDescOnce.Do(func() {
		file_proto_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_proto_rawDescData)
	})
	return file_proto_proto_rawDescData
}

var file_proto_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_proto_proto_goTypes = []interface{}{
	(*CreateGameRequest)(nil),              // 0: CreateGameRequest
	(*CreateGameResponse)(nil),             // 1: CreateGameResponse
	(*JoinGameRequest)(nil),                // 2: JoinGameRequest
	(*JoinGameResponse)(nil),               // 3: JoinGameResponse
	(*MoveRequest)(nil),                    // 4: MoveRequest
	(*MoveResponse)(nil),                   // 5: MoveResponse
	(*GetGameRequest)(nil),                 // 6: GetGameRequest
	(*GetGameResponse)(nil),                // 7: GetGameResponse
	(*GameState)(nil),                      // 8: GameState
	(*GameState_GameWaitingForPlayer)(nil), // 9: GameState.GameWaitingForPlayer
	(*GameState_GameProgress)(nil),         // 10: GameState.GameProgress
	(*GameState_GameResult)(nil),           // 11: GameState.GameResult
}
var file_proto_proto_depIdxs = []int32{
	8,  // 0: CreateGameResponse.state:type_name -> GameState
	8,  // 1: JoinGameResponse.state:type_name -> GameState
	8,  // 2: MoveResponse.state:type_name -> GameState
	8,  // 3: GetGameResponse.state:type_name -> GameState
	9,  // 4: GameState.waiting:type_name -> GameState.GameWaitingForPlayer
	10, // 5: GameState.progress:type_name -> GameState.GameProgress
	11, // 6: GameState.result:type_name -> GameState.GameResult
	0,  // 7: TicTacToeAggregate.CreateGame:input_type -> CreateGameRequest
	2,  // 8: TicTacToeAggregate.JoinGame:input_type -> JoinGameRequest
	4,  // 9: TicTacToeAggregate.Move:input_type -> MoveRequest
	6,  // 10: TicTacToeAggregate.GetGame:input_type -> GetGameRequest
	1,  // 11: TicTacToeAggregate.CreateGame:output_type -> CreateGameResponse
	3,  // 12: TicTacToeAggregate.JoinGame:output_type -> JoinGameResponse
	5,  // 13: TicTacToeAggregate.Move:output_type -> MoveResponse
	7,  // 14: TicTacToeAggregate.GetGame:output_type -> GetGameResponse
	11, // [11:15] is the sub-list for method output_type
	7,  // [7:11] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_proto_proto_init() }
func file_proto_proto_init() {
	if File_proto_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateGameRequest); i {
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
		file_proto_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateGameResponse); i {
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
		file_proto_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinGameRequest); i {
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
		file_proto_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinGameResponse); i {
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
		file_proto_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MoveRequest); i {
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
		file_proto_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MoveResponse); i {
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
		file_proto_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetGameRequest); i {
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
		file_proto_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetGameResponse); i {
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
		file_proto_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameState); i {
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
		file_proto_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameState_GameWaitingForPlayer); i {
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
		file_proto_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameState_GameProgress); i {
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
		file_proto_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GameState_GameResult); i {
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
	file_proto_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*GameState_Waiting)(nil),
		(*GameState_Progress)(nil),
		(*GameState_Result)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_proto_goTypes,
		DependencyIndexes: file_proto_proto_depIdxs,
		MessageInfos:      file_proto_proto_msgTypes,
	}.Build()
	File_proto_proto = out.File
	file_proto_proto_rawDesc = nil
	file_proto_proto_goTypes = nil
	file_proto_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TicTacToeAggregateClient is the client API for TicTacToeAggregate service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TicTacToeAggregateClient interface {
	CreateGame(ctx context.Context, in *CreateGameRequest, opts ...grpc.CallOption) (*CreateGameResponse, error)
	JoinGame(ctx context.Context, in *JoinGameRequest, opts ...grpc.CallOption) (*JoinGameResponse, error)
	Move(ctx context.Context, in *MoveRequest, opts ...grpc.CallOption) (*MoveResponse, error)
	GetGame(ctx context.Context, in *GetGameRequest, opts ...grpc.CallOption) (*GetGameResponse, error)
}

type ticTacToeAggregateClient struct {
	cc grpc.ClientConnInterface
}

func NewTicTacToeAggregateClient(cc grpc.ClientConnInterface) TicTacToeAggregateClient {
	return &ticTacToeAggregateClient{cc}
}

func (c *ticTacToeAggregateClient) CreateGame(ctx context.Context, in *CreateGameRequest, opts ...grpc.CallOption) (*CreateGameResponse, error) {
	out := new(CreateGameResponse)
	err := c.cc.Invoke(ctx, "/TicTacToeAggregate/CreateGame", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ticTacToeAggregateClient) JoinGame(ctx context.Context, in *JoinGameRequest, opts ...grpc.CallOption) (*JoinGameResponse, error) {
	out := new(JoinGameResponse)
	err := c.cc.Invoke(ctx, "/TicTacToeAggregate/JoinGame", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ticTacToeAggregateClient) Move(ctx context.Context, in *MoveRequest, opts ...grpc.CallOption) (*MoveResponse, error) {
	out := new(MoveResponse)
	err := c.cc.Invoke(ctx, "/TicTacToeAggregate/Move", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ticTacToeAggregateClient) GetGame(ctx context.Context, in *GetGameRequest, opts ...grpc.CallOption) (*GetGameResponse, error) {
	out := new(GetGameResponse)
	err := c.cc.Invoke(ctx, "/TicTacToeAggregate/GetGame", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TicTacToeAggregateServer is the server API for TicTacToeAggregate service.
type TicTacToeAggregateServer interface {
	CreateGame(context.Context, *CreateGameRequest) (*CreateGameResponse, error)
	JoinGame(context.Context, *JoinGameRequest) (*JoinGameResponse, error)
	Move(context.Context, *MoveRequest) (*MoveResponse, error)
	GetGame(context.Context, *GetGameRequest) (*GetGameResponse, error)
}

// UnimplementedTicTacToeAggregateServer can be embedded to have forward compatible implementations.
type UnimplementedTicTacToeAggregateServer struct {
}

func (*UnimplementedTicTacToeAggregateServer) CreateGame(context.Context, *CreateGameRequest) (*CreateGameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGame not implemented")
}
func (*UnimplementedTicTacToeAggregateServer) JoinGame(context.Context, *JoinGameRequest) (*JoinGameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method JoinGame not implemented")
}
func (*UnimplementedTicTacToeAggregateServer) Move(context.Context, *MoveRequest) (*MoveResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Move not implemented")
}
func (*UnimplementedTicTacToeAggregateServer) GetGame(context.Context, *GetGameRequest) (*GetGameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetGame not implemented")
}

func RegisterTicTacToeAggregateServer(s *grpc.Server, srv TicTacToeAggregateServer) {
	s.RegisterService(&_TicTacToeAggregate_serviceDesc, srv)
}

func _TicTacToeAggregate_CreateGame_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateGameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TicTacToeAggregateServer).CreateGame(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TicTacToeAggregate/CreateGame",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TicTacToeAggregateServer).CreateGame(ctx, req.(*CreateGameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TicTacToeAggregate_JoinGame_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(JoinGameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TicTacToeAggregateServer).JoinGame(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TicTacToeAggregate/JoinGame",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TicTacToeAggregateServer).JoinGame(ctx, req.(*JoinGameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TicTacToeAggregate_Move_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TicTacToeAggregateServer).Move(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TicTacToeAggregate/Move",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TicTacToeAggregateServer).Move(ctx, req.(*MoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TicTacToeAggregate_GetGame_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetGameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TicTacToeAggregateServer).GetGame(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/TicTacToeAggregate/GetGame",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TicTacToeAggregateServer).GetGame(ctx, req.(*GetGameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _TicTacToeAggregate_serviceDesc = grpc.ServiceDesc{
	ServiceName: "TicTacToeAggregate",
	HandlerType: (*TicTacToeAggregateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateGame",
			Handler:    _TicTacToeAggregate_CreateGame_Handler,
		},
		{
			MethodName: "JoinGame",
			Handler:    _TicTacToeAggregate_JoinGame_Handler,
		},
		{
			MethodName: "Move",
			Handler:    _TicTacToeAggregate_Move_Handler,
		},
		{
			MethodName: "GetGame",
			Handler:    _TicTacToeAggregate_GetGame_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto.proto",
}
