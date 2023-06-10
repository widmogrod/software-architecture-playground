package question

import (
	"context"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/gm"
	"github.com/widmogrod/software-architecture-playground/opus/essence/algebra/kv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ QuestionUseCasesServer = &DefaultQuestionUseCasesServer{}

type DefaultQuestionUseCasesServer struct {
	UnimplementedQuestionUseCasesServer

	Reg   *gm.SchemaRegistry
	Acl   *gm.Guard
	Store *kv.Store
}

func (d *DefaultQuestionUseCasesServer) CreateQuestion(ctx context.Context, request *CreateQuestionRequest) (*CreateQuestionResponse, error) {
	r := gm.InvokeRequest{
		Action:  "CreateQuestionRequest",
		Payload: request,
	}

	err := d.Acl.EvalRule(request.GetSource().GetType(), r)
	if err != nil {
		// wrap error with details
		return nil, status.Errorf(codes.Internal, "invoke1: %s: %w", err)
	}

	data := gm.DefaultQuestion()
	data.Content = kv.PtrString(request.Content)
	data.SourceId = kv.PtrString(request.Source.Id)
	data.SourceType = kv.PtrString(request.Source.Type)

	err = d.Reg.Validate(data.SchemaID(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invoke2: %s: %w", err)
	}

	err = d.Store.SetAttributes(data.ToKey(), data.ToAttr())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invoke3: %s: %w", err)
	}

	return &CreateQuestionResponse{
		State: &QuestionState{
			Source:      nil,
			Question:    nil,
			CqaTaxonomy: nil,
		},
	}, nil
}
