// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mail.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	mail.proto

It has these top-level messages:
	VerificationRequest
	VerificationResponse
	MailRequest
	MailResponse
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type VerificationRequest_Purpose int32

const (
	VerificationRequest_CONFIRMATION   VerificationRequest_Purpose = 0
	VerificationRequest_PASSWORD_RESET VerificationRequest_Purpose = 1
)

var VerificationRequest_Purpose_name = map[int32]string{
	0: "CONFIRMATION",
	1: "PASSWORD_RESET",
}
var VerificationRequest_Purpose_value = map[string]int32{
	"CONFIRMATION":   0,
	"PASSWORD_RESET": 1,
}

func (x VerificationRequest_Purpose) String() string {
	return proto.EnumName(VerificationRequest_Purpose_name, int32(x))
}
func (VerificationRequest_Purpose) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0, 0}
}

type VerificationRequest struct {
	Token   string                      `protobuf:"bytes,1,opt,name=token" json:"token,omitempty"`
	Purpose VerificationRequest_Purpose `protobuf:"varint,2,opt,name=purpose,enum=api.VerificationRequest_Purpose" json:"purpose,omitempty"`
}

func (m *VerificationRequest) Reset()                    { *m = VerificationRequest{} }
func (m *VerificationRequest) String() string            { return proto.CompactTextString(m) }
func (*VerificationRequest) ProtoMessage()               {}
func (*VerificationRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *VerificationRequest) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *VerificationRequest) GetPurpose() VerificationRequest_Purpose {
	if m != nil {
		return m.Purpose
	}
	return VerificationRequest_CONFIRMATION
}

type VerificationResponse struct {
	Email  string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	UserID uint32 `protobuf:"varint,2,opt,name=userID" json:"userID,omitempty"`
}

func (m *VerificationResponse) Reset()                    { *m = VerificationResponse{} }
func (m *VerificationResponse) String() string            { return proto.CompactTextString(m) }
func (*VerificationResponse) ProtoMessage()               {}
func (*VerificationResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *VerificationResponse) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *VerificationResponse) GetUserID() uint32 {
	if m != nil {
		return m.UserID
	}
	return 0
}

type MailRequest struct {
	Email  string `protobuf:"bytes,1,opt,name=email" json:"email,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	UserID uint32 `protobuf:"varint,3,opt,name=userID" json:"userID,omitempty"`
}

func (m *MailRequest) Reset()                    { *m = MailRequest{} }
func (m *MailRequest) String() string            { return proto.CompactTextString(m) }
func (*MailRequest) ProtoMessage()               {}
func (*MailRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MailRequest) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *MailRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *MailRequest) GetUserID() uint32 {
	if m != nil {
		return m.UserID
	}
	return 0
}

type MailResponse struct {
	Status string `protobuf:"bytes,1,opt,name=status" json:"status,omitempty"`
	Code   int32  `protobuf:"varint,2,opt,name=code" json:"code,omitempty"`
}

func (m *MailResponse) Reset()                    { *m = MailResponse{} }
func (m *MailResponse) String() string            { return proto.CompactTextString(m) }
func (*MailResponse) ProtoMessage()               {}
func (*MailResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *MailResponse) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

func (m *MailResponse) GetCode() int32 {
	if m != nil {
		return m.Code
	}
	return 0
}

func init() {
	proto.RegisterType((*VerificationRequest)(nil), "api.VerificationRequest")
	proto.RegisterType((*VerificationResponse)(nil), "api.VerificationResponse")
	proto.RegisterType((*MailRequest)(nil), "api.MailRequest")
	proto.RegisterType((*MailResponse)(nil), "api.MailResponse")
	proto.RegisterEnum("api.VerificationRequest_Purpose", VerificationRequest_Purpose_name, VerificationRequest_Purpose_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for MailService service

type MailServiceClient interface {
	SendConfirmation(ctx context.Context, in *MailRequest, opts ...grpc.CallOption) (*MailResponse, error)
	SendPasswordReset(ctx context.Context, in *MailRequest, opts ...grpc.CallOption) (*MailResponse, error)
	VerifyToken(ctx context.Context, in *VerificationRequest, opts ...grpc.CallOption) (*VerificationResponse, error)
}

type mailServiceClient struct {
	cc *grpc.ClientConn
}

func NewMailServiceClient(cc *grpc.ClientConn) MailServiceClient {
	return &mailServiceClient{cc}
}

func (c *mailServiceClient) SendConfirmation(ctx context.Context, in *MailRequest, opts ...grpc.CallOption) (*MailResponse, error) {
	out := new(MailResponse)
	err := grpc.Invoke(ctx, "/api.MailService/SendConfirmation", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mailServiceClient) SendPasswordReset(ctx context.Context, in *MailRequest, opts ...grpc.CallOption) (*MailResponse, error) {
	out := new(MailResponse)
	err := grpc.Invoke(ctx, "/api.MailService/SendPasswordReset", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mailServiceClient) VerifyToken(ctx context.Context, in *VerificationRequest, opts ...grpc.CallOption) (*VerificationResponse, error) {
	out := new(VerificationResponse)
	err := grpc.Invoke(ctx, "/api.MailService/VerifyToken", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for MailService service

type MailServiceServer interface {
	SendConfirmation(context.Context, *MailRequest) (*MailResponse, error)
	SendPasswordReset(context.Context, *MailRequest) (*MailResponse, error)
	VerifyToken(context.Context, *VerificationRequest) (*VerificationResponse, error)
}

func RegisterMailServiceServer(s *grpc.Server, srv MailServiceServer) {
	s.RegisterService(&_MailService_serviceDesc, srv)
}

func _MailService_SendConfirmation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailServiceServer).SendConfirmation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.MailService/SendConfirmation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailServiceServer).SendConfirmation(ctx, req.(*MailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MailService_SendPasswordReset_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailServiceServer).SendPasswordReset(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.MailService/SendPasswordReset",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailServiceServer).SendPasswordReset(ctx, req.(*MailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MailService_VerifyToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(VerificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MailServiceServer).VerifyToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.MailService/VerifyToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MailServiceServer).VerifyToken(ctx, req.(*VerificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MailService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.MailService",
	HandlerType: (*MailServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendConfirmation",
			Handler:    _MailService_SendConfirmation_Handler,
		},
		{
			MethodName: "SendPasswordReset",
			Handler:    _MailService_SendPasswordReset_Handler,
		},
		{
			MethodName: "VerifyToken",
			Handler:    _MailService_VerifyToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mail.proto",
}

func init() { proto.RegisterFile("mail.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 335 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xc1, 0x6a, 0xf2, 0x40,
	0x14, 0x85, 0xcd, 0xef, 0xaf, 0xe2, 0xd5, 0x4a, 0x9c, 0x8a, 0x58, 0x57, 0x92, 0x95, 0xab, 0x14,
	0xec, 0xaa, 0xee, 0xc4, 0x58, 0x70, 0xa1, 0x91, 0x89, 0xb4, 0xcb, 0x32, 0xd5, 0x2b, 0x0c, 0xd5,
	0x4c, 0x3a, 0x33, 0x69, 0xe9, 0x9b, 0xf4, 0x99, 0xfa, 0x54, 0x25, 0x33, 0x11, 0x22, 0xa4, 0xd0,
	0xdd, 0x9c, 0x49, 0xce, 0x97, 0x73, 0xee, 0x0d, 0xc0, 0x89, 0xf1, 0xa3, 0x9f, 0x48, 0xa1, 0x05,
	0xa9, 0xb2, 0x84, 0x7b, 0x5f, 0x0e, 0x5c, 0x3f, 0xa2, 0xe4, 0x07, 0xbe, 0x63, 0x9a, 0x8b, 0x98,
	0xe2, 0x5b, 0x8a, 0x4a, 0x93, 0x1e, 0xd4, 0xb4, 0x78, 0xc5, 0x78, 0xe0, 0x8c, 0x9c, 0x71, 0x93,
	0x5a, 0x41, 0xa6, 0xd0, 0x48, 0x52, 0x99, 0x08, 0x85, 0x83, 0x7f, 0x23, 0x67, 0xdc, 0x99, 0x8c,
	0x7c, 0x96, 0x70, 0xbf, 0x04, 0xe0, 0x6f, 0xec, 0x7b, 0xf4, 0x6c, 0xf0, 0x6e, 0xa1, 0x91, 0xdf,
	0x11, 0x17, 0xda, 0xf3, 0x70, 0xfd, 0xb0, 0xa4, 0xab, 0xd9, 0x76, 0x19, 0xae, 0xdd, 0x0a, 0x21,
	0xd0, 0xd9, 0xcc, 0xa2, 0xe8, 0x29, 0xa4, 0xc1, 0x33, 0x5d, 0x44, 0x8b, 0xad, 0xeb, 0x78, 0x01,
	0xf4, 0x2e, 0xc1, 0x2a, 0x11, 0xb1, 0xc2, 0x2c, 0x1a, 0x66, 0x35, 0xce, 0xd1, 0x8c, 0x20, 0x7d,
	0xa8, 0xa7, 0x0a, 0xe5, 0x32, 0x30, 0xc9, 0xae, 0x68, 0xae, 0xbc, 0x10, 0x5a, 0x2b, 0xc6, 0x8f,
	0x85, 0x5e, 0x25, 0x66, 0x02, 0xff, 0x63, 0x76, 0xb2, 0xa5, 0x9a, 0xd4, 0x9c, 0x0b, 0xc0, 0xea,
	0x05, 0x70, 0x0a, 0x6d, 0x0b, 0xcc, 0xe3, 0xf4, 0xa1, 0xae, 0x34, 0xd3, 0xa9, 0xca, 0x91, 0xb9,
	0xca, 0x98, 0x3b, 0xb1, 0xb7, 0xcc, 0x1a, 0x35, 0xe7, 0xc9, 0xb7, 0x63, 0xd3, 0x44, 0x28, 0xdf,
	0xf9, 0x0e, 0xc9, 0x3d, 0xb8, 0x11, 0xc6, 0xfb, 0xb9, 0x88, 0x0f, 0x5c, 0x9e, 0x4c, 0x4d, 0xe2,
	0x9a, 0x91, 0x16, 0x32, 0x0f, 0xbb, 0x85, 0x1b, 0xfb, 0x51, 0xaf, 0x42, 0xa6, 0xd0, 0xcd, 0xac,
	0x1b, 0xa6, 0xd4, 0x87, 0x90, 0x7b, 0x8a, 0x0a, 0xf5, 0x5f, 0xbd, 0x01, 0xb4, 0xcc, 0x64, 0x3f,
	0xb7, 0x66, 0xab, 0x83, 0xdf, 0x96, 0x38, 0xbc, 0x29, 0x79, 0x72, 0xa6, 0xbc, 0xd4, 0xcd, 0x6f,
	0x74, 0xf7, 0x13, 0x00, 0x00, 0xff, 0xff, 0x90, 0x2c, 0xa7, 0xe0, 0x54, 0x02, 0x00, 0x00,
}